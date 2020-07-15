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
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	eventsTotal = "kube_events_total"
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
		Reason: "test-creation",
		Type:   v1.EventTypeNormal,
	}
	event = framework.CreateEvent(t, event, event.InvolvedObject.Namespace)

	expectedMetric := &dto.Metric{
		Label: []*dto.LabelPair{
			{Name: stringPtr("involved_object_kind"), Value: &event.InvolvedObject.Kind},
			{Name: stringPtr("involved_object_namespace"), Value: &event.InvolvedObject.Namespace},
			{Name: stringPtr("reason"), Value: &event.Reason},
			{Name: stringPtr("type"), Value: &event.Type},
		},
		Counter: &dto.Counter{Value: float64Ptr(1)},
	}

	err := framework.PollMetric(exporter.GetEventMetricFamilies, eventsTotal, expectedMetric)
	if err != nil {
		t.Fatal(err)
	}
}

func TestEventUpdate(t *testing.T) {
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
		Reason: "test-update",
		Type:   v1.EventTypeNormal,
	}
	event = framework.CreateEvent(t, event, event.InvolvedObject.Namespace)

	event.Count++
	event.LastTimestamp = metav1.Now()
	event, err := framework.UpdateEvent(event, event.InvolvedObject.Namespace)
	if err != nil {
		t.Fatal(err)
	}

	expectedMetric := &dto.Metric{
		Label: []*dto.LabelPair{
			{Name: stringPtr("involved_object_kind"), Value: &event.InvolvedObject.Kind},
			{Name: stringPtr("involved_object_namespace"), Value: &event.InvolvedObject.Namespace},
			{Name: stringPtr("reason"), Value: &event.Reason},
			{Name: stringPtr("type"), Value: &event.Type},
		},
		Counter: &dto.Counter{Value: float64Ptr(2)},
	}

	err = framework.PollMetric(exporter.GetEventMetricFamilies, eventsTotal, expectedMetric)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateExistingEvent(t *testing.T) {
	event := &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		InvolvedObject: v1.ObjectReference{
			Kind:      "Pod",
			Namespace: "default",
		},
		Count:  1,
		Reason: "test-update-existing",
		Type:   v1.EventTypeNormal,
	}
	event = framework.CreateEvent(t, event, event.InvolvedObject.Namespace)
	// The exporter reconciles Events created during the same second as itself.
	// Thus, to ensure that this Event is not reconciled, we sleep one second.
	time.Sleep(time.Second)

	exporter := framework.CreateKubeEventsExporter(t)

	event.Count++
	event.LastTimestamp = metav1.Now()
	event, err := framework.UpdateEvent(event, event.Namespace)
	if err != nil {
		t.Fatal(err)
	}

	expectedMetric := &dto.Metric{
		Label: []*dto.LabelPair{
			{Name: stringPtr("involved_object_kind"), Value: &event.InvolvedObject.Kind},
			{Name: stringPtr("involved_object_namespace"), Value: &event.InvolvedObject.Namespace},
			{Name: stringPtr("reason"), Value: &event.Reason},
			{Name: stringPtr("type"), Value: &event.Type},
		},
		Counter: &dto.Counter{Value: float64Ptr(1)},
	}

	err = framework.PollMetric(exporter.GetEventMetricFamilies, eventsTotal, expectedMetric)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNotReconciling(t *testing.T) {
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
	// The exporter reconciles Events created during the same second as itself.
	// Thus, to ensure that this Event is not reconciled, we sleep one second.
	time.Sleep(time.Second)

	exporter := framework.CreateKubeEventsExporter(t)

	unexpectedMetric := &dto.Metric{
		Label: []*dto.LabelPair{
			{Name: stringPtr("involved_object_kind"), Value: &event.InvolvedObject.Kind},
			{Name: stringPtr("involved_object_namespace"), Value: &event.InvolvedObject.Namespace},
			{Name: stringPtr("reason"), Value: &event.Reason},
			{Name: stringPtr("type"), Value: &event.Type},
		},
		Counter: &dto.Counter{Value: float64Ptr(1)},
	}

	err := framework.PollMetric(exporter.GetEventMetricFamilies, eventsTotal, unexpectedMetric)
	if err == nil {
		t.Fatal("kube-events-exporter should not reconcile existing Events")
	}
}

func TestRecordEventRecorderCreate(t *testing.T) {
	exporter := framework.CreateKubeEventsExporter(t)
	recorder := framework.NewRecordEventRecorder()

	involvedObject := &v1.ObjectReference{
		Kind:      "Pod",
		Namespace: "default",
		Name:      "foo",
	}
	eventType := v1.EventTypeNormal
	reason := "test-recorder-create"

	recorder.Eventf(involvedObject, eventType, reason, "")

	expectedMetric := &dto.Metric{
		Label: []*dto.LabelPair{
			{Name: stringPtr("involved_object_kind"), Value: &involvedObject.Kind},
			{Name: stringPtr("involved_object_namespace"), Value: &involvedObject.Namespace},
			{Name: stringPtr("reason"), Value: &reason},
			{Name: stringPtr("type"), Value: &eventType},
		},
		Counter: &dto.Counter{Value: float64Ptr(1)},
	}

	err := framework.PollMetric(exporter.GetEventMetricFamilies, eventsTotal, expectedMetric)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRecordEventRecorderUpdate(t *testing.T) {
	exporter := framework.CreateKubeEventsExporter(t)
	recorder := framework.NewRecordEventRecorder()

	involvedObject := &v1.ObjectReference{
		Kind:      "Pod",
		Namespace: "default",
		Name:      "foo",
	}
	eventType := v1.EventTypeNormal
	reason := "test-recorder-update"

	recorder.Eventf(involvedObject, eventType, reason, "")
	recorder.Eventf(involvedObject, eventType, reason, "")

	expectedMetric := &dto.Metric{
		Label: []*dto.LabelPair{
			{Name: stringPtr("involved_object_kind"), Value: &involvedObject.Kind},
			{Name: stringPtr("involved_object_namespace"), Value: &involvedObject.Namespace},
			{Name: stringPtr("reason"), Value: &reason},
			{Name: stringPtr("type"), Value: &eventType},
		},
		Counter: &dto.Counter{Value: float64Ptr(2)},
	}

	err := framework.PollMetric(exporter.GetEventMetricFamilies, eventsTotal, expectedMetric)
	if err != nil {
		t.Fatal(err)
	}
}

func TestEventsEventRecorderCreate(t *testing.T) {
	exporter := framework.CreateKubeEventsExporter(t)

	stopCh := make(chan struct{})
	recorder := framework.NewEventsEventRecorder(stopCh)

	involvedObject := &v1.ObjectReference{
		Kind:      "Pod",
		Namespace: "default",
		Name:      "foo",
	}
	eventType := v1.EventTypeNormal
	reason := "test-recorder-create"

	recorder.Eventf(involvedObject, nil, eventType, reason, "action", "")

	expectedMetric := &dto.Metric{
		Label: []*dto.LabelPair{
			{Name: stringPtr("involved_object_kind"), Value: &involvedObject.Kind},
			{Name: stringPtr("involved_object_namespace"), Value: &involvedObject.Namespace},
			{Name: stringPtr("reason"), Value: &reason},
			{Name: stringPtr("type"), Value: &eventType},
		},
		Counter: &dto.Counter{Value: float64Ptr(1)},
	}

	err := framework.PollMetric(exporter.GetEventMetricFamilies, eventsTotal, expectedMetric)
	if err != nil {
		t.Fatal(err)
	}
}

func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}
