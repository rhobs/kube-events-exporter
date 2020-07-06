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
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rhobs/kube-events-exporter/internal/options"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// EventCollector is a prometeus.Collector that bundles all the metrics related
// to Kubernetes Events.
type EventCollector struct {
	eventsTotal *prometheus.CounterVec

	lock              sync.Mutex
	informerFactories []informers.SharedInformerFactory
	filter            eventFilter
}

// NewEventCollector returns a prometheus.Collector collecting metrics about
// Kubernetes Events.
func NewEventCollector(kubeClient kubernetes.Interface, opts *options.Options) *EventCollector {
	var factories []informers.SharedInformerFactory
	for _, ns := range opts.InvolvedObjectNamespaces {
		for _, eventType := range opts.EventTypes {
			factories = append(factories, newFilteredInformerFactory(kubeClient, ns, eventType))
		}
	}

	collector := &EventCollector{
		eventsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "kube_events_total",
			Help: "Count of all Kubernetes Events",
		}, []string{"type", "involved_object_namespace", "involved_object_kind", "reason"}),

		lock:              sync.Mutex{},
		informerFactories: factories,
		filter: eventFilter{
			creationTimestamp: time.Now(),
			apiGroups:         opts.InvolvedObjectAPIGroups,
		},
	}

	collector.initInformers()

	return collector
}

func newFilteredInformerFactory(kubeClient kubernetes.Interface, ns, eventType string) informers.SharedInformerFactory {
	return informers.NewFilteredSharedInformerFactory(
		kubeClient,
		0,
		metav1.NamespaceAll,
		func(list *metav1.ListOptions) {
			filterInvolvedObjectNs(list, ns)
			filterEventType(list, eventType)
		},
	)
}

// Describe implements the prometheus.Collector interface.
func (collector *EventCollector) Describe(ch chan<- *prometheus.Desc) {
	collector.eventsTotal.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (collector *EventCollector) Collect(ch chan<- prometheus.Metric) {
	collector.eventsTotal.Collect(ch)
}

// Run starts updating EventCollector metrics.
func (collector *EventCollector) Run(stopCh <-chan struct{}) {
	for _, factory := range collector.informerFactories {
		go factory.Start(stopCh)
	}
}

func (collector *EventCollector) initInformers() {
	for _, factory := range collector.informerFactories {
		eventsTotalInformer := factory.Core().V1().Events().Informer()
		eventsTotalInformer.AddEventHandler(collector.eventsTotalHandler())
	}
}

func (collector *EventCollector) eventsTotalHandler() cache.ResourceEventHandler {
	return cache.FilteringResourceEventHandler{
		FilterFunc: collector.filter.filter,
		Handler: &cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				ev := obj.(*v1.Event)
				collector.increaseEventsTotal(ev, 1)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				oldEv := oldObj.(*v1.Event)
				newEv := newObj.(*v1.Event)
				nbNew := updatedEventNb(oldEv, newEv)
				collector.increaseEventsTotal(newEv, float64(nbNew))
			},
		},
	}
}

func (collector *EventCollector) increaseEventsTotal(event *v1.Event, nbNew float64) {
	collector.lock.Lock()
	collector.eventsTotal.With(prometheus.Labels{
		"type":                      event.Type,
		"involved_object_namespace": event.InvolvedObject.Namespace,
		"involved_object_kind":      event.InvolvedObject.Kind,
		"reason":                    event.Reason,
	}).Add(nbNew)
	collector.lock.Unlock()
}

func updatedEventNb(oldEv, newEv *v1.Event) int32 {
	if newEv.Series != nil {
		if oldEv.Series != nil {
			return newEv.Series.Count - oldEv.Series.Count
		}
		// When event is emitted for the first time it's written to the API
		// server without series field set.
		return newEv.Series.Count
	}

	return newEv.Count - oldEv.Count
}

func filterInvolvedObjectNs(list *metav1.ListOptions, ns string) {
	if ns != metav1.NamespaceAll {
		list.FieldSelector += ",involvedObject.namespace=" + ns
	}
}

func filterEventType(list *metav1.ListOptions, eventType string) {
	if eventType != options.EventTypesAll {
		list.FieldSelector += ",type=" + eventType
	}
}
