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

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ()

var (
	kubeEventsExporterService = &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app.kubernetes.io/component": "exporter",
				"app.kubernetes.io/name":      "kube-events-exporter",
			},
			Name: "kube-events-exporter",
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app.kubernetes.io/component": "exporter",
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

	// EventServerURL contains the URL to access the event server from outside.
	EventServerURL string
	// ExporterServerURL contains the URL to access the exporter server from
	// outside.
	ExporterServerURL string
)

// CreateKubeEventsExporter creates kube-events-exporter deployment inside
// of the specified namespace.
func (f *Framework) CreateKubeEventsExporter(ns, exporterImage string) ([]finalizerFn, error) {
	var finalizers []finalizerFn

	service, err := f.CreateService(kubeEventsExporterService, ns)
	if err != nil {
		return nil, errors.Wrap(err, "create kube-events-exporter service")
	}
	finalizers = append(finalizers, func() error { return f.DeleteService(service.Namespace, service.Name) })

	EventServerURL = fmt.Sprintf("http://localhost:8001/api/v1/namespaces/%s/services/kube-events-exporter:event/proxy/", ns)
	ExporterServerURL = fmt.Sprintf("http://localhost:8001/api/v1/namespaces/%s/services/kube-events-exporter:exporter/proxy/", ns)

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

	deployment, err = f.CreateDeployment(deployment, ns)
	if err != nil {
		return nil, errors.Wrap(err, "create kube-events-exporter deployment")
	}
	finalizers = append(finalizers, func() error { return f.DeleteDeployment(deployment.Namespace, deployment.Name) })

	err = f.WaitUntilDeploymentReady(deployment.Namespace, deployment.Name)
	if err != nil {
		return nil, errors.Wrap(err, "kube-events-exporter not ready")
	}

	return finalizers, nil
}
