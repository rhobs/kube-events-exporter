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
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateClusterRoleBinding creates the given cluster role binding.
func (f *Framework) CreateClusterRoleBinding(t *testing.T, crb *v1.ClusterRoleBinding) *v1.ClusterRoleBinding {
	crb, err := f.KubeClient.RbacV1().ClusterRoleBindings().Create(context.TODO(), crb, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("create cluster role binding %s: %v", crb.Name, err)
	}

	t.Cleanup(func() {
		err := f.DeleteClusterRoleBinding(crb.Name)
		if err != nil {
			t.Fatal(err)
		}
	})

	return crb
}

// MakeClusterRoleBinding creates a cluster role binding object from yaml
// manifest.
func MakeClusterRoleBinding(manifestPath string) (*v1.ClusterRoleBinding, error) {
	manifest, err := ioutil.ReadFile(filepath.Clean(manifestPath))
	if err != nil {
		return nil, errors.Wrapf(err, "read cluster role binding manifest %s", manifestPath)
	}

	crb := v1.ClusterRoleBinding{}
	err = yaml.Unmarshal(manifest, &crb)
	if err != nil {
		return nil, errors.Wrapf(err, "unmarshal cluster role binding manifest %s", manifestPath)
	}

	return &crb, nil
}

// DeleteClusterRoleBinding deletes the given cluster role binding.
func (f *Framework) DeleteClusterRoleBinding(name string) error {
	err := f.KubeClient.RbacV1().ClusterRoleBindings().Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(err, "delete cluster role binding %s", name)
	}
	return nil
}
