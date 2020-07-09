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

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateEvent creates the given Event.
func (f *Framework) CreateEvent(t *testing.T, event *v1.Event, ns string) *v1.Event {
	event, err := f.KubeClient.CoreV1().Events(ns).Create(context.TODO(), event, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("create event %s: %v", event.Name, err)
	}

	t.Cleanup(func() {
		err := f.DeleteEvent(event.Namespace, event.Name)
		if err != nil {
			t.Fatal(err)
		}
	})

	return event
}

// UpdateEvent updates the given Event.
func (f *Framework) UpdateEvent(event *v1.Event, ns string) (*v1.Event, error) {
	event, err := f.KubeClient.CoreV1().Events(ns).Update(context.TODO(), event, metav1.UpdateOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "update event %s", event.Name)
	}
	return event, nil
}

// DeleteEvent deletes the given Event.
func (f *Framework) DeleteEvent(ns, name string) error {
	err := f.KubeClient.CoreV1().Events(ns).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(err, "delete event %s", name)
	}
	return nil
}
