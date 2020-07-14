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

func TestEmittedEvent(t *testing.T) {
	now := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	older := now.Add(-time.Minute)
	newer := now.Add(time.Minute)

	testCases := []struct {
		desc     string
		time     time.Time
		event    *v1.Event
		expected bool
	}{
		{
			desc:     "Older",
			time:     now,
			event:    &v1.Event{EventTime: metav1.NewMicroTime(older)},
			expected: false,
		},
		{
			desc:     "Equal",
			time:     now,
			event:    &v1.Event{EventTime: metav1.NewMicroTime(now)},
			expected: true,
		},
		{
			desc:     "Newer",
			time:     now,
			event:    &v1.Event{EventTime: metav1.NewMicroTime(newer)},
			expected: true,
		},
		{
			desc:     "TruncateTime",
			time:     now.Add(100 * time.Millisecond),
			event:    &v1.Event{EventTime: metav1.NewMicroTime(now)},
			expected: true,
		},
		{
			desc:     "TruncateEvent",
			time:     now.Add(200 * time.Millisecond),
			event:    &v1.Event{EventTime: metav1.NewMicroTime(now.Add(100 * time.Millisecond))},
			expected: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			got := emittedEvent(tc.event, tc.time)
			if got != tc.expected {
				t.Fatalf("expected %t, got %t", tc.expected, got)
			}
		})
	}
}

func TestGetEventLatestTimestamp(t *testing.T) {
	now := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	older := now.Add(-time.Minute)
	newer := now.Add(time.Minute)

	testCases := []struct {
		desc       string
		tweakEvent func(*v1.Event)
		expected   time.Time
	}{
		{
			desc:       "OlderEvent",
			tweakEvent: func(*v1.Event) {},
			expected:   older,
		},
		{
			desc:       "CreationTimestamp",
			tweakEvent: func(ev *v1.Event) { ev.CreationTimestamp = metav1.NewTime(newer) },
			expected:   newer,
		},
		{
			desc:       "FirstTimestamp",
			tweakEvent: func(ev *v1.Event) { ev.FirstTimestamp = metav1.NewTime(newer) },
			expected:   newer,
		},
		{
			desc:       "LastTimestamp",
			tweakEvent: func(ev *v1.Event) { ev.LastTimestamp = metav1.NewTime(newer) },
			expected:   newer,
		},
		{
			desc:       "EventTime",
			tweakEvent: func(ev *v1.Event) { ev.EventTime = metav1.NewMicroTime(newer) },
			expected:   newer,
		},
		{
			desc:       "creationTimestamp",
			tweakEvent: func(ev *v1.Event) { ev.Series.LastObservedTime = metav1.NewMicroTime(newer) },
			expected:   newer,
		},
	}

	olderEvent := &v1.Event{
		FirstTimestamp: metav1.NewTime(older),
		LastTimestamp:  metav1.NewTime(older),
		EventTime:      metav1.NewMicroTime(older),
		Series:         &v1.EventSeries{LastObservedTime: metav1.NewMicroTime(older)},
	}
	olderEvent.CreationTimestamp = metav1.NewTime(older)

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			ev := olderEvent.DeepCopy()
			tc.tweakEvent(ev)
			got := getEventLatestTimestamp(ev)
			if got != tc.expected {
				t.Fatalf("expected %s as latest timestamp", tc.expected)
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
