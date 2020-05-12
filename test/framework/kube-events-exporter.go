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

	"github.com/pkg/errors"
	dto "github.com/prometheus/client_model/go"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ()

var (
	kubeEventsExporterService = &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app.kubernetes.io/component": "events-exporter",
				"app.kubernetes.io/name":      "kube-events-exporter",
			},
			Name: "kube-events-exporter",
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
	Finalizers        []finalizerFn
}

// CreateKubeEventsExporter creates kube-events-exporter deployment inside
// of the specified namespace.
func (f *Framework) CreateKubeEventsExporter(ns, exporterImage string) (*KubeEventsExporter, error) {
	var finalizers []finalizerFn

	service, err := f.CreateService(kubeEventsExporterService, ns)
	if err != nil {
		return nil, errors.Wrap(err, "create kube-events-exporter service")
	}
	finalizers = append(finalizers, func() error { return f.DeleteService(service.Namespace, service.Name) })

	serviceURL := fmt.Sprintf("http://localhost:8001/api/v1/namespaces/%s/services/%s", ns, service.ObjectMeta.Name)
	eventServerURL := fmt.Sprintf("%s:%s/proxy/", serviceURL, service.Spec.Ports[0].Name)
	exporterServerURL := fmt.Sprintf("%s:%s/proxy/", serviceURL, service.Spec.Ports[1].Name)

	deployment, err := MakeDeployment("../../manifests/kube-events-exporter-deployment.yaml")
	if err != nil {
		return nil, errors.Wrap(err, "make kube-events-exporter deployment")
	}

	if exporterImage != "" {
		// Override kube-events-exporter image with the one specified.
		deployment.Spec.Template.Spec.Containers[0].Image = exporterImage
	}

	// TODO: create rbac configuration
	deployment.Spec.Template.Spec.ServiceAccountName = ""

	deployment.Namespace = ns
	err = f.CreateDeployment(deployment)
	if err != nil {
		return nil, errors.Wrap(err, "create kube-events-exporter deployment")
	}
	finalizers = append(finalizers, func() error { return f.DeleteDeployment(deployment.Namespace, deployment.Name) })

	err = f.WaitUntilDeploymentReady(deployment.Namespace, deployment.Name)
	if err != nil {
		return nil, errors.Wrap(err, "kube-events-exporter not ready")
	}

	exporter := &KubeEventsExporter{
		Deployment:        deployment,
		EventServerURL:    eventServerURL,
		ExporterServerURL: exporterServerURL,
		Finalizers:        finalizers,
	}

	return exporter, nil
}

// ResetExporterMetrics resets the exporter metrics by recreating
// kube-events-exporter pod.
func (f *Framework) ResetExporterMetrics() error {
	// Delete kube-events-exporter pod.
	err := f.UpdateDeploymentReplicas(f.Exporter.Deployment, 0)
	if err != nil {
		return errors.Wrapf(err, "update deployment %s replicas to 0", f.Exporter.Deployment.Name)
	}

	// Recreate kube-events-exporter pod.
	err = f.UpdateDeploymentReplicas(f.Exporter.Deployment, 1)
	if err != nil {
		return errors.Wrapf(err, "update deployment %s replicas to 1", f.Exporter.Deployment.Name)
	}

	return nil
}

func (f *Framework) getMetricFamilies(serverURL string) (map[string]*dto.MetricFamily, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, errors.Wrapf(err, "parse url: %s", serverURL)
	}
	u.Path = path.Join(u.Path, "metrics")

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, errors.Wrapf(err, "send GET request %s", u.String())
	}

	families, err := f.MetricsParser.TextToMetricFamilies(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "parse text to metric families %s", u.String())
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, errors.Wrapf(err, "close response body %s", u.String())
	}

	return families, nil
}

// GetEventMetricFamilies gets metrics from the event server metrics endpoint
// and converts them to Prometheus MetricFamily.
func (f *Framework) GetEventMetricFamilies() (map[string]*dto.MetricFamily, error) {
	return f.getMetricFamilies(f.Exporter.EventServerURL)
}

// GetExporterMetricFamilies gets metrics from the exporter server metrics
// endpoint and converts them to Prometheus MetricFamily.
func (f *Framework) GetExporterMetricFamilies() (map[string]*dto.MetricFamily, error) {
	return f.getMetricFamilies(f.Exporter.ExporterServerURL)
}
