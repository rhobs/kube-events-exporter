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
	"context"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateService creates the given service.
func (f *Framework) CreateService(t *testing.T, service *v1.Service, ns string) *v1.Service {
	service, err := f.KubeClient.CoreV1().Services(ns).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Could not create service %s: %v\n", service.Name, err)
	}

	t.Cleanup(func() {
		err := f.DeleteService(service.Namespace, service.Name)
		if err != nil {
			t.Fatalf("Could not delete service %s: %v\n", service.Name, err)
		}
	})

	return service
}

// DeleteService deletes the given service.
func (f *Framework) DeleteService(ns, name string) error {
	err := f.KubeClient.CoreV1().Services(ns).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
