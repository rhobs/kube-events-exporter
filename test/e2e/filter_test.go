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
	"fmt"
	"testing"

	"github.com/rhobs/kube-events-exporter/test/framework"
)

func TestNamespaceFilter(t *testing.T) {
	namespace := framework.NewBasicEvent().InvolvedObject.Namespace

	testCases := []struct {
		name      string
		namespace string
		assert    framework.AssertEventsTotalFunc
	}{
		{
			name:      "Included",
			namespace: namespace,
			assert:    f.AssertEventsTotalPresent,
		},
		{
			name:      "Excluded",
			namespace: fmt.Sprintf("not-%s", namespace),
			assert:    f.AssertEventsTotalAbsent,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			f := *f
			f.ExporterArgs = []string{fmt.Sprintf("--involved-object-namespaces=%s", tc.namespace)}
			exporter := f.CreateKubeEventsExporter(t)
			event := f.CreateBasicEvent(t)
			tc.assert(t, exporter, event, 1)
		})
	}
}

func TestEventTypeFilter(t *testing.T) {
	eventType := framework.NewBasicEvent().Type

	testCases := []struct {
		name      string
		eventType string
		assert    framework.AssertEventsTotalFunc
	}{
		{
			name:      "Included",
			eventType: eventType,
			assert:    f.AssertEventsTotalPresent,
		},
		{
			name:      "Excluded",
			eventType: fmt.Sprintf("not-%s", eventType),
			assert:    f.AssertEventsTotalAbsent,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			f := *f
			f.ExporterArgs = []string{fmt.Sprintf("--event-types=%s", tc.eventType)}
			exporter := f.CreateKubeEventsExporter(t)
			event := f.CreateBasicEvent(t)
			tc.assert(t, exporter, event, 1)
		})
	}
}
