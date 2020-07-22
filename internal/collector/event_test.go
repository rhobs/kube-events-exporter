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

	v1 "k8s.io/api/core/v1"
)

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
