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

package collectors

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rhobs/kube-events-exporter/internal/version"
)

// InstrumentMetricHandler is a middleware that wraps the provided http.Handler
// to observe requests sent to the exporter.
func InstrumentMetricHandler(registry *prometheus.Registry, handler http.Handler) http.Handler {
	requestsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kube_events_exporter_requests_total",
		Help: "Total number of scrapes.",
	}, []string{"code"})

	requestsInFlight := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "kube_events_exporter_requests_in_flight",
		Help: "Current number of scrapes being served.",
	})

	requestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "kube_events_exporter_request_duration_seconds",
		Help: "Duration of all scrapes.",
	}, []string{"code"})

	registry.MustRegister(
		requestsTotal,
		requestsInFlight,
		requestDuration,
	)

	return promhttp.InstrumentHandlerDuration(
		requestDuration.MustCurryWith(prometheus.Labels{"code": "200"}),
		promhttp.InstrumentHandlerInFlight(requestsInFlight,
			promhttp.InstrumentHandlerCounter(requestsTotal, handler),
		),
	)
}

// RegisterVersionCollector registers a Gauge metric describing the exporter
// version.
func RegisterVersionCollector(registry *prometheus.Registry) {
	exporterVersion := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "kube_events_exporter_version",
		Help: "Version of the exporter.",
		ConstLabels: map[string]string{
			"version": version.GetVersion(),
		},
	})
	exporterVersion.Set(1)

	registry.MustRegister(exporterVersion)
}
