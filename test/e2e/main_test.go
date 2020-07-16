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

package e2e

import (
	"flag"
	"log"
	"os"
	"testing"

	"github.com/rhobs/kube-events-exporter/test/framework"
)

var (
	f *framework.Framework
)

func TestMain(m *testing.M) {
	kubeconfig := flag.String(
		"kubeconfig",
		os.Getenv("KUBECONFIG"),
		"Absolute path to the kubeconfig file.",
	)
	exporterImage := flag.String(
		"exporter-image",
		"",
		"Exporter container image as specified in a deployment manifest.",
	)
	flag.Parse()

	var err error
	f, err = framework.NewFramework(*kubeconfig, *exporterImage)
	if err != nil {
		log.Fatalf("setup test framework: %v", err)
	}

	os.Exit(m.Run())
}
