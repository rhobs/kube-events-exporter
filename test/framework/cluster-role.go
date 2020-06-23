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

// CreateClusterRole creates the given cluster role.
func (f *Framework) CreateClusterRole(t *testing.T, cr *v1.ClusterRole) *v1.ClusterRole {
	cr, err := f.KubeClient.RbacV1().ClusterRoles().Create(context.TODO(), cr, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("create cluster role %s: %v", cr.Name, err)
	}

	t.Cleanup(func() {
		err := f.DeleteClusterRole(cr.Name)
		if err != nil {
			t.Fatal(err)
		}
	})

	return cr
}

// MakeClusterRole creates a cluster role object from yaml manifest.
func MakeClusterRole(manifestPath string) (*v1.ClusterRole, error) {
	manifest, err := ioutil.ReadFile(filepath.Clean(manifestPath))
	if err != nil {
		return nil, errors.Wrapf(err, "read cluster role manifest %s", manifestPath)
	}

	cr := v1.ClusterRole{}
	err = yaml.Unmarshal(manifest, &cr)
	if err != nil {
		return nil, errors.Wrapf(err, "unmarshal cluster role manifest %s", manifestPath)
	}

	return &cr, nil
}

// DeleteClusterRole deletes the given cluster role.
func (f *Framework) DeleteClusterRole(name string) error {
	err := f.KubeClient.RbacV1().ClusterRoles().Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(err, "delete cluster role %s", name)
	}
	return nil
}
