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

package collector

import (
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBeforeLatestEvent(t *testing.T) {
	now := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	older := now.Add(-time.Minute)
	newer := now.Add(time.Minute)

	newerSimpleEvent := &v1.Event{}
	newerSimpleEvent.CreationTimestamp = metav1.NewTime(newer)

	olderSimpleEvent := &v1.Event{}
	olderSimpleEvent.CreationTimestamp = metav1.NewTime(older)

	scenarios := []struct {
		Desc     string
		Event    *v1.Event
		Expected bool
	}{
		{
			Desc:     "api:newer",
			Event:    newerSimpleEvent,
			Expected: true,
		},
		{
			Desc:     "api:older",
			Event:    olderSimpleEvent,
			Expected: false,
		},
		{
			Desc: "core.EventRecorder:newer",
			Event: &v1.Event{
				FirstTimestamp: metav1.NewTime(older),
				LastTimestamp:  metav1.NewTime(newer),
			},
			Expected: true,
		},
		{
			Desc: "core.EventRecorder:older",
			Event: &v1.Event{
				FirstTimestamp: metav1.NewTime(older),
				LastTimestamp:  metav1.NewTime(older),
			},
			Expected: false,
		},
		{
			Desc: "events.EventRecorder:newer",
			Event: &v1.Event{
				EventTime: metav1.NewMicroTime(newer),
			},
			Expected: true,
		},
		{
			Desc: "events.EventRecorder:older",
			Event: &v1.Event{
				EventTime: metav1.NewMicroTime(older),
			},
			Expected: false,
		},
		{
			Desc: "events.EventRecorder:newer_serie",
			Event: &v1.Event{
				Series: &v1.EventSeries{LastObservedTime: metav1.NewMicroTime(newer)},
			},
			Expected: true,
		},
		{
			Desc: "events.EventRecorder:older_serie",
			Event: &v1.Event{
				Series: &v1.EventSeries{LastObservedTime: metav1.NewMicroTime(older)},
			},
			Expected: false,
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.Desc, func(t *testing.T) {
			t.Parallel()
			result := beforeLatestEvent(now, scenario.Event)
			if result != scenario.Expected {
				t.Fatalf("expected %t", scenario.Expected)
			}
		})
	}
}

func TestUpdatedEventNb(t *testing.T) {
	scenarios := []struct {
		Desc     string
		OldEvent v1.Event
		NewEvent v1.Event
		Expected int32
	}{
		{
			Desc:     "0_updated",
			OldEvent: v1.Event{Count: 1},
			NewEvent: v1.Event{Count: 1},
			Expected: 0,
		},
		{
			Desc:     "10_updated",
			OldEvent: v1.Event{Count: 1},
			NewEvent: v1.Event{Count: 11},
			Expected: 10,
		},
		{
			Desc:     "0_series_updated",
			OldEvent: v1.Event{Series: &v1.EventSeries{Count: 1}},
			NewEvent: v1.Event{Series: &v1.EventSeries{Count: 1}},
			Expected: 0,
		},
		{
			Desc:     "10_series_updated",
			OldEvent: v1.Event{Series: &v1.EventSeries{Count: 1}},
			NewEvent: v1.Event{Series: &v1.EventSeries{Count: 11}},
			Expected: 10,
		},
		{
			Desc:     "new_serie",
			OldEvent: v1.Event{},
			NewEvent: v1.Event{Series: &v1.EventSeries{Count: 1}},
			Expected: 1,
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.Desc, func(t *testing.T) {
			t.Parallel()
			nbNew := updatedEventNb(&scenario.OldEvent, &scenario.NewEvent)
			if nbNew != scenario.Expected {
				t.Errorf("expected %d updated Events, got %d", scenario.Expected, nbNew)
			}
		})
	}
}
