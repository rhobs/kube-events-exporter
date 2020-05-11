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

	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rhobs/kube-events-exporter/internal/collector"
	"github.com/rhobs/kube-events-exporter/internal/exporter"
	exporterhttp "github.com/rhobs/kube-events-exporter/internal/http"
	"github.com/rhobs/kube-events-exporter/internal/options"
	"github.com/rhobs/kube-events-exporter/internal/version"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

func main() {
	opts := options.NewOptions()
	opts.AddFlags()

	err := opts.Parse()
	if err != nil {
		klog.Fatalf("failed to parse options: %v", err)
	}

	if opts.Version {
		fmt.Printf("%#v\n", version.GetVersion())
		os.Exit(0)
	}

	kubeConfig, err := clientcmd.BuildConfigFromFlags(opts.Apiserver, opts.Kubeconfig)
	if err != nil {
		klog.Fatalf("failed to create cluster config from flags: %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		klog.Fatalf("failed to create cluster config: %v", err)
	}

	eventRegistry := prometheus.NewRegistry()
	eventCollector := collector.NewEventCollector(kubeClient, opts)

	stopCh := make(chan struct{})
	defer close(stopCh)

	eventCollector.Run(stopCh)
	eventRegistry.MustRegister(eventCollector)

	exporterRegistry := prometheus.NewRegistry()
	exporter.RegisterExporterCollectors(exporterRegistry)

	eventMux := http.NewServeMux()
	exporterhttp.RegisterEventsMuxHandlers(eventMux, eventRegistry, exporterRegistry)
	exporterMux := http.NewServeMux()
	exporterhttp.RegisterExporterMuxHandlers(exporterMux, exporterRegistry)

	var rg run.Group
	rg.Add(listenAndServe(exporterMux, opts.ExporterHost, opts.ExporterPort))
	rg.Add(listenAndServe(eventMux, opts.Host, opts.Port))
	klog.Fatalf("metrics servers terminated: %v", rg.Run())
}

func listenAndServe(mux *http.ServeMux, host string, port int) (func() error, func(error)) {
	var listener net.Listener
	serve := func() error {
		addr := net.JoinHostPort(host, strconv.Itoa(port))
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		return http.Serve(listener, mux)
	}
	cleanup := func(error) {
		err := listener.Close()
		if err != nil {
			klog.Errorf("failed to close listener: %v", err)
		}
	}
	return serve, cleanup
}
