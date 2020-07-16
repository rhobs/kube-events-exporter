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

	"github.com/rhobs/kube-events-exporter/test/framework"
)

func TestEventCreation(t *testing.T) {
	exporter := f.CreateKubeEventsExporter(t)
	event := f.CreateBasicEvent(t)

	f.AssertEventsTotalPresent(t, exporter, event, 1)
}

func TestEventUpdate(t *testing.T) {
	exporter := f.CreateKubeEventsExporter(t)
	event := f.CreateBasicEvent(t)

	event, err := f.UpdateEvent(event, event.InvolvedObject.Namespace)
	if err != nil {
		t.Fatal(err)
	}

	f.AssertEventsTotalPresent(t, exporter, event, 2)
}

func TestUpdateExistingEvent(t *testing.T) {
	event := f.CreateBasicEvent(t)
	// The exporter reconcile Events created during the same second as itself.
	// Thus, to ensure that this Event is not reconciled, we sleep one second.
	time.Sleep(time.Second)

	exporter := f.CreateKubeEventsExporter(t)

	event, err := f.UpdateEvent(event, event.Namespace)
	if err != nil {
		t.Fatal(err)
	}

	f.AssertEventsTotalPresent(t, exporter, event, 1)
}

func TestNotReconciling(t *testing.T) {
	event := f.CreateBasicEvent(t)
	// The exporter reconcile Events created during the same second as itself.
	// Thus, to ensure that this Event is not reconciled, we sleep one second.
	time.Sleep(time.Second)

	exporter := f.CreateKubeEventsExporter(t)

	f.AssertEventsTotalAbsent(t, exporter, event, 1)
}

func TestRecordEventRecorderCreate(t *testing.T) {
	exporter := f.CreateKubeEventsExporter(t)
	recorder := f.NewRecordEventRecorder()

	event := framework.NewBasicEvent()
	recorder.Eventf(&event.InvolvedObject, event.Type, event.Reason, event.Message)

	f.AssertEventsTotalPresent(t, exporter, event, 1)
}

func TestRecordEventRecorderUpdate(t *testing.T) {
	exporter := f.CreateKubeEventsExporter(t)
	recorder := f.NewRecordEventRecorder()

	event := framework.NewBasicEvent()
	recorder.Eventf(&event.InvolvedObject, event.Type, event.Reason, event.Message)
	recorder.Eventf(&event.InvolvedObject, event.Type, event.Reason, event.Message)

	f.AssertEventsTotalPresent(t, exporter, event, 2)
}

func TestEventsEventRecorderCreate(t *testing.T) {
	exporter := f.CreateKubeEventsExporter(t)

	stopCh := make(chan struct{})
	recorder := f.NewEventsEventRecorder(stopCh)

	event := framework.NewBasicEvent()
	recorder.Eventf(&event.InvolvedObject, nil, event.Type, event.Reason, event.Action, event.Message)

	f.AssertEventsTotalPresent(t, exporter, event, 1)
}
