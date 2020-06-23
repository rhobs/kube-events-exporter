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

package e2e

import (
	"reflect"
	"testing"

	dto "github.com/prometheus/client_model/go"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEventCreation(t *testing.T) {
	exporter := framework.CreateKubeEventsExporter(t)

	event := &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		InvolvedObject: v1.ObjectReference{
			Kind:      "Pod",
			Namespace: "default",
		},
		Count:  1,
		Reason: "test",
		Type:   v1.EventTypeNormal,
	}
	event = framework.CreateEvent(t, event, event.InvolvedObject.Namespace)
	err := framework.WaitUntilEventReady(event.Namespace, event.Name)
	if err != nil {
		t.Fatal(err)
	}

	families, err := exporter.GetEventMetricFamilies()
	if err != nil {
		t.Fatal(err)
	}

	eventsTotal, found := families["kube_events_total"]
	if !found {
		t.Fatal("kube_events_total metric not found")
	}

	expectedMetric := dto.Metric{
		Label: []*dto.LabelPair{
			{Name: stringPtr("involved_object_kind"), Value: &event.InvolvedObject.Kind},
			{Name: stringPtr("involved_object_namespace"), Value: &event.InvolvedObject.Namespace},
			{Name: stringPtr("reason"), Value: &event.Reason},
			{Name: stringPtr("type"), Value: &event.Type},
		},
		Counter: &dto.Counter{Value: float64Ptr(1)},
	}

	for _, metric := range eventsTotal.Metric {
		if reflect.DeepEqual(metric.Label, expectedMetric.Label) {
			if !reflect.DeepEqual(metric.Counter, expectedMetric.Counter) {
				t.Fatalf("kube_events_total value is %v instead of %v", metric.Counter.GetValue(), expectedMetric.Counter.GetValue())
			}
			return
		}
	}
	t.Fatal("kube_events_total metric not found")
}

func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}
