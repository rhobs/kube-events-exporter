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

// CreateServiceAccount creates the given service account.
func (f *Framework) CreateServiceAccount(t *testing.T, sa *v1.ServiceAccount, ns string) *v1.ServiceAccount {
	sa, err := f.KubeClient.CoreV1().ServiceAccounts(ns).Create(context.TODO(), sa, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("create service account %s: %v", sa.Name, err)
	}

	t.Cleanup(func() {
		err := f.DeleteServiceAccount(sa.Namespace, sa.Name)
		if err != nil {
			t.Fatal(err)
		}
	})

	return sa
}

// MakeServiceAccount creates a service account object from yaml manifest.
func MakeServiceAccount(manifestPath string) (*v1.ServiceAccount, error) {
	manifest, err := ioutil.ReadFile(filepath.Clean(manifestPath))
	if err != nil {
		return nil, errors.Wrapf(err, "read service account manifest %s", manifestPath)
	}

	sa := v1.ServiceAccount{}
	err = yaml.Unmarshal(manifest, &sa)
	if err != nil {
		return nil, errors.Wrapf(err, "unmarshal service account manifest %s", manifestPath)
	}

	return &sa, nil
}

// DeleteServiceAccount deletes the given service account.
func (f *Framework) DeleteServiceAccount(ns, name string) error {
	err := f.KubeClient.CoreV1().ServiceAccounts(ns).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(err, "delete service account %s", name)
	}
	return nil
}
