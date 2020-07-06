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

package options

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const (
	// EventTypesAll is the argument to specify to allow all Event types.
	EventTypesAll = ""

	// APIGroupsAll is the argument to specify to allow objects from all API
	// groups.
	APIGroupsAll = ""
)

// Options are the configurable parameters for kube-events-exporter.
type Options struct {
	Apiserver    string
	Kubeconfig   string
	Host         string
	Port         int
	ExporterHost string
	ExporterPort int
	Version      bool

	EventTypes               []string
	InvolvedObjectAPIGroups  []string
	InvolvedObjectNamespaces []string

	flags *pflag.FlagSet
}

// NewOptions returns a new instance of `Options`.
func NewOptions() *Options {
	return &Options{}
}

// AddFlags populates the Options struct from the command line arguments passed.
func (o *Options) AddFlags() {
	o.flags = pflag.NewFlagSet("", pflag.ExitOnError)

	// Add klog flags.
	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)
	o.flags.AddGoFlagSet(klogFlags)

	o.flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		o.flags.PrintDefaults()
	}

	o.flags.StringVar(&o.Apiserver, "apiserver", "", "The URL of the apiserver to use as a master.")
	o.flags.StringVar(&o.Kubeconfig, "kubeconfig", os.Getenv("KUBECONFIG"), "Absolute path to the kubeconfig file.")
	o.flags.StringVar(&o.Host, "host", "0.0.0.0", "Host to expose Events metrics on.")
	o.flags.IntVar(&o.Port, "port", 8080, "Port to expose Events metrics on.")
	o.flags.StringVar(&o.ExporterHost, "exporter-host", "0.0.0.0", "Host to expose kube-events-exporter own metrics on.")
	o.flags.IntVar(&o.ExporterPort, "exporter-port", 8081, "Port to expose kube-events-exporter own metrics on.")
	o.flags.BoolVar(&o.Version, "version", false, "kube-events-exporter version information")

	o.flags.StringArrayVar(&o.EventTypes, "event-types", []string{EventTypesAll}, "List of allowed Event types. Defaults to all types.")
	o.flags.StringArrayVar(&o.InvolvedObjectAPIGroups, "involved-object-api-groups", []string{APIGroupsAll}, "List of allowed Event involved object API groups. Defaults to all API groups.")
	o.flags.StringArrayVar(&o.InvolvedObjectNamespaces, "involved-object-namespaces", []string{metav1.NamespaceAll}, "List of allowed Event involved object namespaces. Defaults to all namespaces.")
}

// Parse parses the flag definitions from the argument list.
func (o *Options) Parse() error {
	err := o.flags.Parse(os.Args)
	return err
}

// Usage is the function called when an error occurs while parsing flags.
func (o *Options) Usage() {
	o.flags.Usage()
}
