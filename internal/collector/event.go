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
	"github.com/prometheus/client_golang/prometheus"
)

// EventCollector is a prometeus.Collector that bundles all the metrics related
// to Kubernetes Events.
type EventCollector struct {
	eventsTotal *prometheus.CounterVec
}

// NewEventCollector returns a prometheus.Collector collecting metrics about
// Kubernetes Events.
func NewEventCollector() *EventCollector {
	return &EventCollector{
		eventsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "kube_events_total",
			Help: "Count of all Kubernetes Events",
		}, []string{"type", "involved_object_namespace", "involved_object_kind", "reason"}),
	}
}

// Describe implements the prometheus.Collector interface.
func (collector *EventCollector) Describe(ch chan<- *prometheus.Desc) {
	collector.eventsTotal.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (collector *EventCollector) Collect(ch chan<- prometheus.Metric) {
	collector.eventsTotal.Collect(ch)
}
