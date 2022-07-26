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

package routes

import (
	"path"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/spf13/cobra"
	"knative.dev/pkg/ptr"
)

// RouteFlags includes commonly passed in flags to define a HTTP route.
type RouteFlags struct {
	Hostname string
	Path     string
}

// Add appends the flags to the given command
func (flags *RouteFlags) Add(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&flags.Hostname,
		"hostname",
		"",
		"Hostname for the Route.",
	)

	cmd.Flags().StringVar(
		&flags.Path,
		"path",
		"",
		"URL path for the Route.",
	)
}

// RouteSpecFields converts the flags to a RouteSpecFields instance
func (flags *RouteFlags) RouteSpecFields(domain string) v1alpha1.RouteSpecFields {
	return v1alpha1.RouteSpecFields{
		Hostname: flags.Hostname,
		Domain:   domain,
		Path:     path.Join("/", flags.Path),
	}
}

// routeBindingFlags represents the keys for a route binding
type routeBindingFlags struct {
	RouteFlags
	destinationPort int32
}

// Add appends the flags to the given command
func (flags *routeBindingFlags) Add(cmd *cobra.Command) {
	flags.RouteFlags.Add(cmd)

	cmd.Flags().Int32Var(
		&flags.destinationPort,
		"destination-port",
		0,
		"Port on the App the Route will connect to.",
	)
}

// RouteWeightBinding converts the flags to a RouteWeightBinding instance
func (flags *routeBindingFlags) RouteWeightBinding(domain string) v1alpha1.RouteWeightBinding {
	tmp := v1alpha1.RouteWeightBinding{
		RouteSpecFields: flags.RouteSpecFields(domain),
	}

	if flags.destinationPort != 0 {
		tmp.DestinationPort = ptr.Int32(flags.destinationPort)
	}

	return tmp
}
