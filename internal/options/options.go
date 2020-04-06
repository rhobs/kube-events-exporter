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
	"k8s.io/klog"
)

// Options are the configurable parameters for kube-events-exporter.
type Options struct {
	ExporterHost string
	ExporterPort int
	Version      bool

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

	o.flags.StringVar(&o.ExporterHost, "exporter-host", "0.0.0.0", `Host to expose kube-events-exporter self metrics on.`)
	o.flags.IntVar(&o.ExporterPort, "exporter-port", 8081, `Port to expose kube-events-exporter self metrics on.`)
	o.flags.BoolVar(&o.Version, "version", false, "kube-events-exporter version information")
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
