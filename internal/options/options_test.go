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
	"os"
	"testing"
)

func TestOptionsParse(t *testing.T) {
	tests := []struct {
		Desc string
		Args []string
	}{
		{
			Desc: "version command line argument",
			Args: []string{"./kube-events-exporter", "--version"},
		},
		{
			Desc: "exporter port and host command line argument",
			Args: []string{"./kube-events-exporter",
				"--host=127.0.0.1",
				"--port=8080",
				"--exporter-host=127.0.0.1",
				"--exporter-port=8081",
			},
		},
	}

	for _, test := range tests {
		opts := NewOptions()
		opts.AddFlags()

		os.Args = test.Args

		err := opts.Parse()
		if err != nil {
			t.Errorf("Test error for Desc: %s.", test.Desc)
		}
	}
}
