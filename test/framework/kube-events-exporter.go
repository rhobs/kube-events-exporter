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
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
)

// CreateKubeEventsExporter creates kube-events-exporter deployment inside
// of the specified namespace.
func (f *Framework) CreateKubeEventsExporter(ns string) (*appsv1.Deployment, error) {
	deployment, err := MakeDeployment("../../manifests/deployment.yaml")
	if err != nil {
		return nil, errors.Wrap(err, "make kube-events-exporter deployment")
	}

	deployment, err = f.CreateDeployment(deployment, ns)
	if err != nil {
		return nil, errors.Wrap(err, "create kube-events-exporter deployment")
	}

	err = f.WaitUntilDeploymentReady(deployment.Namespace, deployment.Name)
	if err != nil {
		return nil, errors.Wrap(err, "kube-events-exporter not ready")
	}

	return deployment, nil
}
