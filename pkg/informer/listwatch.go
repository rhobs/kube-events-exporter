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

package informer

import (
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

// ListWatchMetrics stores the pointers of list/watch counter metrics.
type ListWatchMetrics struct {
	listTotal        prometheus.Counter
	listFailedTotal  prometheus.Counter
	watchTotal       prometheus.Counter
	watchFailedTotal prometheus.Counter
}

// NewListWatchMetrics takes in a prometheus registry and initializes and
// registers list and watch metrics.
func NewListWatchMetrics(registry *prometheus.Registry) *ListWatchMetrics {
	metrics := &ListWatchMetrics{
		listTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "kube_events_exporter_list_total",
				Help: "Number of times a list operation was performed",
			},
		),
		listFailedTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "kube_events_exporter_list_failed_total",
				Help: "Number of times a list operation failed",
			},
		),
		watchTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "kube_events_exporter_watch_total",
				Help: "Number of times a watch operation was performed",
			},
		),
		watchFailedTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "kube_events_exporter_watch_failed_total",
				Help: "Number of times a watch operation failed",
			},
		),
	}
	registry.MustRegister(
		metrics.listTotal,
		metrics.listFailedTotal,
		metrics.watchTotal,
		metrics.watchFailedTotal,
	)
	return metrics
}

// InstrumentedListerWatcher provides the list/watch metrics with a cache
// ListerWatcher obj and the related resource.
type InstrumentedListerWatcher struct {
	lw      cache.ListerWatcher
	metrics *ListWatchMetrics
}

// NewInstrumentedListerWatcher returns a new InstrumentedListerWatcher.
func NewInstrumentedListerWatcher(lw cache.ListerWatcher, metrics *ListWatchMetrics) cache.ListerWatcher {
	return &InstrumentedListerWatcher{
		lw:      lw,
		metrics: metrics,
	}
}

// List is a wrapper func around the cache.ListerWatcher.List func. It
// increases the success/error counters based on the outcome of the List
// operation it instruments.
func (i *InstrumentedListerWatcher) List(options metav1.ListOptions) (runtime.Object, error) {
	i.metrics.listTotal.Inc()

	res, err := i.lw.List(options)
	if err != nil {
		i.metrics.listFailedTotal.Inc()
		return nil, err
	}

	return res, nil
}

// Watch is a wrapper func around the cache.ListerWatcher.Watch func. It
// increases the success/error counters based on the outcome of the Watch
// operation it instruments.
func (i *InstrumentedListerWatcher) Watch(options metav1.ListOptions) (watch.Interface, error) {
	i.metrics.watchTotal.Inc()

	res, err := i.lw.Watch(options)
	if err != nil {
		i.metrics.watchFailedTotal.Inc()
		return nil, err
	}

	return res, nil
}
