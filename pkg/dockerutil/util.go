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

package dockerutil

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/google/kf/v2/pkg/kf/describe"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// ReadConfig is a utility function to read a docker config from a path.
// if the path is blank the default config is used.
func ReadConfig(configPath string) (*configfile.ConfigFile, error) {
	return config.Load(configPath)
}

// DescribeConfig creates a function that can write information about a
// configuration file.
func DescribeConfig(w io.Writer, cfg *configfile.ConfigFile) {
	describe.SectionWriter(w, "Docker config", func(w io.Writer) {
		describe.SectionWriter(w, "Auth", func(w io.Writer) {
			if len(cfg.AuthConfigs) == 0 {
				fmt.Fprintln(w, "<none>")
				return
			}

			var registries []string
			for registry := range cfg.AuthConfigs {
				registries = append(registries, registry)
			}
			sort.Strings(registries)

			fmt.Fprintln(w, "Registry\tUsername\tEmail")
			for _, registry := range registries {
				authConfig := cfg.AuthConfigs[registry]

				fmt.Fprintf(w, "%s\t%s\t%s\n",
					registry,
					authConfig.Username,
					authConfig.Email,
				)
			}
		})

		describe.SectionWriter(w, "Credential helpers", func(w io.Writer) {
			if len(cfg.CredentialHelpers) == 0 {
				fmt.Fprintln(w, "<none>")
				return
			}

			var registries []string
			for registry := range cfg.CredentialHelpers {
				registries = append(registries, registry)
			}
			sort.Strings(registries)

			fmt.Fprintln(w, "Registry\tHelper")
			for _, registry := range registries {
				fmt.Fprintf(w, "%s\t%s\n", registry, cfg.CredentialHelpers[registry])
			}
		})
	})
}

// DescribeDefaultConfig writes debug info about the default docker
// configuration to the given writer.
func DescribeDefaultConfig(w io.Writer) {
	cfg, err := ReadConfig("")
	if err != nil {
		fmt.Fprintf(w, "couldn't read default docker config: %v\n", err)
	} else {
		DescribeConfig(w, cfg)
	}
}

// DescribeWorkloadIdentity writes debug info about the identity
// that can be accessed via the GCE Metadata server.
//
// If the server can't be reached, a message gets printed that the server
// couldn't be found.
func DescribeWorkloadIdentity(w io.Writer) {
	// XXX: As of 2020-05-15 the GCE Metadata API doesn't support mocking of any
	// kind for tests.

	// WI on the cluster provides a limited set of metadata:
	// https://cloud.google.com/kubernetes-engine/docs/concepts/workload-identity#metadata_server
	const defaultSA = "default"

	// Create a custom client so the function can fail fast.
	// In GKE the WI server is listening on each node in a DaemonSet so the dial
	// timeout can be very short.
	client := metadata.NewClient(&http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   500 * time.Millisecond,
				KeepAlive: 5 * time.Second,
			}).Dial,
		},
	})

	reports := []struct {
		name string
		exec func() (interface{}, error)
	}{
		{
			name: "Account",
			exec: func() (interface{}, error) { return client.Email(defaultSA) },
		},
		{
			name: "Scopes",
			exec: func() (interface{}, error) { return client.Scopes(defaultSA) },
		},
		{
			name: "Zone",
			exec: func() (interface{}, error) { return client.Zone() },
		},
		{
			name: "Cluster",
			exec: func() (interface{}, error) { return client.InstanceAttributeValue("cluster-name") },
		},
	}

	describe.SectionWriter(w, "Workload Identity", func(w io.Writer) {
		for idx, report := range reports {
			val, err := report.exec()

			// First see if we got a timeout error indicating the WI server isn't
			// listening and we haven't reported anything yet.
			if e, ok := err.(net.Error); ok && idx == 0 && e.Timeout() {
				fmt.Fprintln(w, "Server not found")
				break
			}

			if err != nil {
				fmt.Fprintf(w, "%s:\t%s\n", report.name, err.Error())
			} else {
				fmt.Fprintf(w, "%s:\t%v\n", report.name, val)
			}
		}
	})
}

// GetAuthKeyChain constructs the default auth key chain to authenticate
// with GCP services, e.g. GCR.
func GetAuthKeyChain() remote.Option {
	return remote.WithAuthFromKeychain(authn.NewMultiKeychain(
		authn.DefaultKeychain,
		google.Keychain,
	))
}
