/*
Copyright 2020 Red Hat, Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/pkg/errors"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	serviceAccountManifest     = "../../manifests/kube-events-exporter-service-account.yaml"
	clusterRoleManifest        = "../../manifests/kube-events-exporter-cluster-role.yaml"
	clusterRoleBindingManifest = "../../manifests/kube-events-exporter-cluster-role-binding.yaml"
	deploymentManifest         = "../../manifests/kube-events-exporter-deployment.yaml"

	exporterNamespace = "default"
)

var (
	exporterService = &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app.kubernetes.io/component": "events-exporter",
				"app.kubernetes.io/name":      "kube-events-exporter",
			},
			Name:      "kube-events-exporter",
			Namespace: exporterNamespace,
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app.kubernetes.io/component": "events-exporter",
				"app.kubernetes.io/name":      "kube-events-exporter",
			},
			Ports: []v1.ServicePort{
				{
					Name: "event",
					Port: 8080,
				},
				{
					Name: "exporter",
					Port: 8081,
				},
			},
		},
	}
)

// KubeEventsExporter exposes information needed by the framework to interact
// with kube-events-exporter.
type KubeEventsExporter struct {
	Deployment        *appsv1.Deployment
	EventServerURL    string
	ExporterServerURL string
}

// CreateKubeEventsExporter creates kube-events-exporter deployment inside
// of the specified namespace.
func (f *Framework) CreateKubeEventsExporter(t *testing.T) *KubeEventsExporter {
	sa, err := MakeServiceAccount(serviceAccountManifest)
	if err != nil {
		t.Fatal(err)
	}
	f.CreateServiceAccount(t, sa, exporterNamespace)

	cr, err := MakeClusterRole(clusterRoleManifest)
	if err != nil {
		t.Fatal(err)
	}
	f.CreateClusterRole(t, cr)

	crb, err := MakeClusterRoleBinding(clusterRoleBindingManifest)
	if err != nil {
		t.Fatal(err)
	}
	f.CreateClusterRoleBinding(t, crb)

	f.CreateService(t, exporterService, exporterService.Namespace)

	serviceURL := fmt.Sprintf("http://localhost:8001/api/v1/namespaces/%s/services/%s", exporterService.Namespace, exporterService.ObjectMeta.Name)
	eventServerURL := fmt.Sprintf("%s:%s/proxy/", serviceURL, exporterService.Spec.Ports[0].Name)
	exporterServerURL := fmt.Sprintf("%s:%s/proxy/", serviceURL, exporterService.Spec.Ports[1].Name)

	deployment, err := MakeDeployment(deploymentManifest)
	if err != nil {
		t.Fatal(err)
	}
	if f.ExporterImage != "" {
		// Override kube-events-exporter image with the one specified.
		deployment.Spec.Template.Spec.Containers[0].Image = f.ExporterImage
	}
	deployment = f.CreateDeployment(t, deployment, exporterNamespace)

	exporter := &KubeEventsExporter{
		Deployment:        deployment,
		EventServerURL:    eventServerURL,
		ExporterServerURL: exporterServerURL,
	}

	err = f.waitUntilExporterReady(exporter)
	if err != nil {
		t.Fatal(err)
	}

	return exporter
}

func (f *Framework) waitUntilExporterReady(exporter *KubeEventsExporter) error {
	err := wait.Poll(time.Second, f.DefaultTimeout, func() (bool, error) {
		resp, err := http.Get(fmt.Sprintf("%s/healthz", exporter.EventServerURL))
		if err != nil {
			return false, err
		}
		err = resp.Body.Close()
		if err != nil {
			return false, err
		}

		return resp.StatusCode == http.StatusOK, nil
	})
	if err != nil {
		return errors.Wrapf(err, "kube-events-exporter not ready")
	}
	return nil
}

// GetEventMetricFamilies gets metrics from the event server metrics endpoint
// and converts them to Prometheus MetricFamily.
func (e *KubeEventsExporter) GetEventMetricFamilies() (map[string]*dto.MetricFamily, error) {
	return getMetricFamilies(e.EventServerURL)
}

// GetExporterMetricFamilies gets metrics from the exporter server metrics
// endpoint and converts them to Prometheus MetricFamily.
func (e *KubeEventsExporter) GetExporterMetricFamilies() (map[string]*dto.MetricFamily, error) {
	return getMetricFamilies(e.ExporterServerURL)
}

func getMetricFamilies(serverURL string) (map[string]*dto.MetricFamily, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, errors.Wrapf(err, "parse url: %s", serverURL)
	}
	u.Path = path.Join(u.Path, "metrics")

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, errors.Wrapf(err, "send GET request %s", u.String())
	}

	parser := expfmt.TextParser{}
	families, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "parse text to metric families %s", u.String())
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, errors.Wrapf(err, "close response body %s", u.String())
	}

	return families, nil
}

type metricFamiliesGetter = func() (map[string]*dto.MetricFamily, error)

// PollMetric tries to find the given metric in the metric families returned by
// the provided getter until the framework default timeout.
func (f *Framework) PollMetric(getter metricFamiliesGetter, name string, expectedMetric *dto.Metric) error {
	err := wait.Poll(time.Second, f.DefaultTimeout, func() (bool, error) {
		families, err := getter()
		if err != nil {
			return false, err
		}

		eventsTotal, found := families[name]
		if !found {
			return false, nil
		}

		for _, metric := range eventsTotal.Metric {
			if reflect.DeepEqual(metric, expectedMetric) {
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		return errors.Errorf("%s expected metric not found", name)
	}
	return nil
}
