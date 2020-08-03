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
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers/internalinterfaces"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// NewInstrumentedEventInformer constructs a new informer for Event type with
// instrumented list watch.
func NewInstrumentedEventInformer(client kubernetes.Interface, namespace string, metrics *ListWatchMetrics, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		NewInstrumentedListerWatcher(
			NewEventListerWatcher(client, namespace, tweakListOptions),
			metrics,
		),
		&v1.Event{},
		resyncPeriod,
		indexers,
	)
}

// NewEventListerWatcher constructs a new lister watcher for Event type.
func NewEventListerWatcher(client kubernetes.Interface, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.ListerWatcher {
	return &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			if tweakListOptions != nil {
				tweakListOptions(&options)
			}
			return client.CoreV1().Events(namespace).List(context.TODO(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			if tweakListOptions != nil {
				tweakListOptions(&options)
			}
			return client.CoreV1().Events(namespace).Watch(context.TODO(), options)
		},
	}
}
