// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"context"
	"path"
)

var defaultRouteWeight int32 = 1

const (
	// RouteHostname is the hostname of a route.
	RouteHostname = "route.kf.dev/hostname"
	// RouteDomain is the domain of a route.
	RouteDomain = "route.kf.dev/domain"
	// RoutePath is the URL path of a route.
	RoutePath = "route.kf.dev/path"
	// RouteAppName is the App's name that owns the Route.
	RouteAppName = "route.kf.dev/appname"

	// DefaultRouteDestinationPort holds the default port route traffic is sent to.
	DefaultRouteDestinationPort = 80
)

type routeDefaultDomain struct{}

var routeDefaultDomainKey = routeDefaultDomain{}

// WithRouteDefaultDomain sets the default domain for a route.
func WithRouteDefaultDomain(ctx context.Context, defaultDomain string) context.Context {
	return context.WithValue(ctx, routeDefaultDomainKey, defaultDomain)
}

func getRouteDefaultDomain(ctx context.Context) *string {
	if typed, ok := ctx.Value(routeDefaultDomainKey).(string); ok {
		return &typed
	}

	return nil
}

type defaultDestPort struct{}

var defaultDestPortKey = defaultDestPort{}

// WithRouteDefaultDestinationPort sets the default destination port for the context
func WithRouteDefaultDestinationPort(ctx context.Context, port int32) context.Context {
	return context.WithValue(ctx, defaultDestPortKey, port)
}

func getRouteDefaultDestinationPort(ctx context.Context) *int32 {
	if typed, ok := ctx.Value(defaultDestPortKey).(int32); ok {
		return &typed
	}

	return nil
}

// GenerateRouteName creates the deterministic name for a Route.
func GenerateRouteName(hostname, domain, urlPath string) string {
	return GenerateName(hostname, domain, path.Join("/", urlPath), "")
}

// SetDefaults implements apis.Defaultable
func (k *RouteWeightBinding) SetDefaults(ctx context.Context) {
	if k.Weight == nil {
		k.Weight = &defaultRouteWeight
	}

	if defaultPort := getRouteDefaultDestinationPort(ctx); k.DestinationPort == nil && defaultPort != nil {
		k.DestinationPort = defaultPort
	}

	k.RouteSpecFields.SetDefaults(ctx)
}

// SetDefaults implements apis.Defaultable
func (k *RouteSpecFields) SetDefaults(ctx context.Context) {
	k.Path = path.Join("/", k.Path)

	if defaultDomain := getRouteDefaultDomain(ctx); k.Domain == "" && defaultDomain != nil {
		k.Domain = *defaultDomain
	}
}

func defaultRouteLabels() map[string]string {
	return map[string]string{
		ManagedByLabel: "kf",
		ComponentLabel: "route",
	}
}

// SetDefaults sets the defaults for a Route.
func (k *Route) SetDefaults(ctx context.Context) {
	k.Spec.SetDefaults(ctx)
	k.Labels = UnionMaps(k.Labels, defaultRouteLabels())
}

// SetDefaults implements apis.Defaultable
func (k *RouteSpec) SetDefaults(ctx context.Context) {
	k.RouteSpecFields.SetDefaults(ctx)
}
