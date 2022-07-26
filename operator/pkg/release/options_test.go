// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"os"
	"strconv"
	"testing"
)

func buildOpts(devMode bool, gcr string, gcs string, version string) options {
	o := options{}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	o.addFlags()
	flag.Set("development", strconv.FormatBool(devMode))
	flag.Set("gcrPath", gcr)
	flag.Set("gcsPath", gcs)
	flag.Set("version", version)
	return o
}

func TestMustProcessFlags(t *testing.T) {
	processFlagTests := []struct {
		name        string
		devMode     bool
		gcrString   string
		gcsString   string
		verString   string
		expectError bool
	}{
		{
			name:        "Successful",
			devMode:     false,
			gcrString:   "gcr.io/project123/registry1",
			gcsString:   "project123/storage1",
			verString:   "1.2.3-gke.8",
			expectError: false,
		},
		{
			name:        "Commit hash works in dev mode",
			devMode:     true,
			gcrString:   "gcr.io/project123/registry1",
			gcsString:   "project123/storage1",
			verString:   "deadbeef",
			expectError: false,
		},
		{
			name:        "semver works in dev mode",
			devMode:     true,
			gcrString:   "gcr.io/project123/registry1",
			gcsString:   "project123/storage1",
			verString:   "1.2.3-gke.8",
			expectError: false,
		},
		{
			name:        "bad commit hash in dev mode fails",
			devMode:     true,
			gcrString:   "gcr.io/project123/registry1",
			gcsString:   "project123/storage1",
			verString:   "livebeef",
			expectError: true,
		},
		{
			name:        "Empty gcr path fails",
			devMode:     false,
			gcrString:   "",
			gcsString:   "project123/storage1",
			verString:   "1.2.3-gke.8",
			expectError: true,
		},
		{
			name:        "Empty gcs path fails",
			devMode:     false,
			gcrString:   "gcr.io/project123/registry1",
			gcsString:   "",
			verString:   "1.2.3-gke.8",
			expectError: true,
		},
		{
			name:        "malformed gcr path fails",
			devMode:     false,
			gcrString:   "malformed/project123/registry1",
			gcsString:   "project123/storage1",
			verString:   "1.2.3-gke.8",
			expectError: true,
		},
		{
			name:        "Bad version fails",
			devMode:     false,
			gcrString:   "gcr.io/project123/registry1",
			gcsString:   "project123/storage1",
			verString:   "1.2.3",
			expectError: true,
		},
	}
	for _, tt := range processFlagTests {
		o := buildOpts(tt.devMode, tt.gcrString, tt.gcsString, tt.verString)
		actual := o.mustProcessFlags()
		gotError := actual != nil
		if gotError != tt.expectError {
			t.Errorf("Return value from o.mustProcessFlags() didn't match when invocation was %v", tt)
		}
	}
}
