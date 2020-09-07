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
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateService creates the given service.
func (f *Framework) CreateService(t *testing.T, service *v1.Service, ns string) *v1.Service {
	service, err := f.KubeClient.CoreV1().Services(ns).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("create service %s: %v", service.Name, err)
	}

	t.Cleanup(func() {
		err := f.DeleteService(service.Namespace, service.Name)
		if err != nil {
			t.Fatalf("delete service %s: %v", service.Name, err)
		}
	})

	return service
}

// MakeService creates a service object from yaml manifest.
func MakeService(manifestPath string) (*v1.Service, error) {
	manifest, err := ioutil.ReadFile(filepath.Clean(manifestPath))
	if err != nil {
		return nil, errors.Wrapf(err, "read service manifest %s", manifestPath)
	}

	service := v1.Service{}
	err = yaml.Unmarshal(manifest, &service)
	if err != nil {
		return nil, errors.Wrapf(err, "unmarshal service manifest %s", manifestPath)
	}

	return &service, nil
}

// DeleteService deletes the given service.
func (f *Framework) DeleteService(ns, name string) error {
	err := f.KubeClient.CoreV1().Services(ns).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
