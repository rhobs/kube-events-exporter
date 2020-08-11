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
	"time"

	"github.com/rhobs/kube-events-exporter/internal/options"
	v1 "k8s.io/api/core/v1"
)

type eventFilter struct {
	creationTimestamp time.Time
	apiGroups         []string
	controllers       []string
}

func (f *eventFilter) filter(obj interface{}) bool {
	ev := obj.(*v1.Event)

	// Count only Events that were freshly emitted and not reconciled
	// during the start of the informer.
	if reconciledEvent(ev, f.creationTimestamp) {
		return false
	}

	if !includedObjectAPIGroup(ev, f.apiGroups) {
		return false
	}

	return includedController(ev, f.controllers)
}

func reconciledEvent(ev *v1.Event, t time.Time) bool {
	// Truncate timestamps to unify MicroTime and Time.
	latest := getEventLatestTimestamp(ev).Truncate(time.Second)
	t = t.Truncate(time.Second)
	return latest.Before(t)
}

func getEventLatestTimestamp(ev *v1.Event) time.Time {
	eventTimes := []time.Time{
		ev.FirstTimestamp.Time,
		ev.LastTimestamp.Time,
		ev.EventTime.Time,
	}

	if ev.Series != nil {
		eventTimes = append(eventTimes, ev.Series.LastObservedTime.Time)
	}

	latest := ev.CreationTimestamp.Time
	for _, eventTime := range eventTimes {
		if eventTime.After(latest) {
			latest = eventTime
		}
	}

	return latest
}

func includedObjectAPIGroup(ev *v1.Event, groups []string) bool {
	if groups[0] == options.APIGroupAll {
		return true
	}

	for _, group := range groups {
		if group == ev.InvolvedObject.APIVersion {
			return true
		}
	}

	return false
}

func includedController(ev *v1.Event, controllers []string) bool {
	if controllers[0] == options.ReportingControllerAll {
		return true
	}

	for _, c := range controllers {
		if c == ev.Source.Component {
			return true
		} else if c == ev.ReportingController {
			return true
		}
	}

	return false
}
