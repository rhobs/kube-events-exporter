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

package http

import (
	"net"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rhobs/kube-events-exporter/internal/collectors"
	"k8s.io/klog"
)

const (
	metricsPath = "/metrics"
)

// ServeMetrics registers collectors and start serving metrics on metricPath.
func ServeMetrics(host string, port int) {
	registry := prometheus.NewRegistry()

	// Add exporter version collector to the registry.
	collectors.RegisterVersionCollector(registry)

	// Address to listen on for web interface and telemetry.
	listenAddress := net.JoinHostPort(host, strconv.Itoa(port))

	klog.Infof("Starting exporter metrics server: %s", listenAddress)
	mux := http.NewServeMux()

	// Add instrumented metricsPath.
	metricsHandler := collectors.InstrumentMetricHandler(registry,
		promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	)
	mux.Handle(metricsPath, metricsHandler)

	klog.Fatal(http.ListenAndServe(listenAddress, mux))
}
