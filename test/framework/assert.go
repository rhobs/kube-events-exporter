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
	"testing"

	dto "github.com/prometheus/client_model/go"

	v1 "k8s.io/api/core/v1"
)

const (
	eventsTotal = "kube_events_total"
)

// AssertEventsTotalFunc is a function that assert kube_events_total metrics.
type AssertEventsTotalFunc func(t *testing.T, exporter *KubeEventsExporter, event *v1.Event, count float64)

// AssertEventsTotalPresent asserts that the kube_events_total metric related
// to the given event and count is present on the exporter.
func (f *Framework) AssertEventsTotalPresent(t *testing.T, exporter *KubeEventsExporter, event *v1.Event, count float64) {
	err := pollEventsTotalMetric(f, exporter, event, count)
	if err != nil {
		t.Fatal(err)
	}
}

// AssertEventsTotalAbsent asserts that the kube_events_total metric related to
// the given event and count is absent on the exporter.
func (f *Framework) AssertEventsTotalAbsent(t *testing.T, exporter *KubeEventsExporter, event *v1.Event, count float64) {
	err := pollEventsTotalMetric(f, exporter, event, count)
	if err == nil {
		t.Fatal("found unexpected kube_events_total metric")
	}
}

func pollEventsTotalMetric(f *Framework, exporter *KubeEventsExporter, event *v1.Event, count float64) error {
	metric := expectedEventsTotal(event, count)
	return f.PollMetric(exporter.GetEventMetricFamilies, eventsTotal, metric)
}

func expectedEventsTotal(ev *v1.Event, count float64) *dto.Metric {
	return &dto.Metric{
		Label: []*dto.LabelPair{
			{Name: stringPtr("involved_object_kind"), Value: &ev.InvolvedObject.Kind},
			{Name: stringPtr("involved_object_namespace"), Value: &ev.InvolvedObject.Namespace},
			{Name: stringPtr("reason"), Value: &ev.Reason},
			{Name: stringPtr("type"), Value: &ev.Type},
		},
		Counter: &dto.Counter{Value: &count},
	}
}

func stringPtr(s string) *string {
	return &s
}
