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

package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rhobs/kube-events-exporter/internal/collector"
	"github.com/rhobs/kube-events-exporter/internal/exporter"
	exporterhttp "github.com/rhobs/kube-events-exporter/internal/http"
	"github.com/rhobs/kube-events-exporter/internal/options"
	"github.com/rhobs/kube-events-exporter/internal/version"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

func main() {
	opts := options.NewOptions()
	opts.AddFlags()

	err := opts.Parse()
	if err != nil {
		klog.Fatalf("Error: %s", err)
	}

	if opts.Version {
		fmt.Printf("%#v\n", version.GetVersion())
		os.Exit(0)
	}

	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatalf("Could not create InCluster config: %v", err)
	}

	eventRegistry := prometheus.NewRegistry()
	eventCollector := collector.NewEventCollector()
	eventCollector.WithInformerFactory(informers.NewSharedInformerFactory(
		kubernetes.NewForConfigOrDie(kubeConfig),
		0,
	))
	stopCh := make(chan struct{})
	eventCollector.Run(stopCh)
	defer close(stopCh)
	eventRegistry.MustRegister(eventCollector)

	exporterRegistry := prometheus.NewRegistry()
	exporter.RegisterExporterCollectors(exporterRegistry)

	eventMux := http.NewServeMux()
	exporterhttp.RegisterEventsMuxHandlers(eventMux, eventRegistry, exporterRegistry)
	exporterMux := http.NewServeMux()
	exporterhttp.RegisterExporterMuxHandlers(exporterMux, exporterRegistry)

	eventListenAddr := net.JoinHostPort(opts.Host, strconv.Itoa(opts.Port))
	eventListener, err := net.Listen("tcp", eventListenAddr)
	if err != nil {
		klog.Fatalf("Could not start listening on: %v: %v", eventListenAddr, err)
	}
	exporterListenAddr := net.JoinHostPort(opts.ExporterHost, strconv.Itoa(opts.ExporterPort))
	exporterListener, err := net.Listen("tcp", exporterListenAddr)
	if err != nil {
		klog.Fatalf("Could not start listening on: %v: %v", exporterListenAddr, err)
	}

	// Serve metrics about the exporter.
	go func() {
		klog.Fatal(http.Serve(exporterListener, exporterMux))
	}()

	// Serve metrics about Kubernetes Events.
	klog.Fatal(http.Serve(eventListener, eventMux))
}
