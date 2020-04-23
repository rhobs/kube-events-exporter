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

package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

// RegisterExporterCollectors register the following collectors in the given
// prometheus.Registry:
//   - prometheus.NewProcessCollector
//   - prometheus.NewGoCollector
//   - collectors.NewExporterVersionCollector
// This is intended to be used to expose metrics about the exporter.
func RegisterExporterCollectors(registry *prometheus.Registry) {
	registry.MustRegister(
		// Add the standard process and Go metrics to the registry.
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		prometheus.NewGoCollector(),
		// Add exporter version collector to the registry.
		NewExporterVersionCollector(),
	)
}
