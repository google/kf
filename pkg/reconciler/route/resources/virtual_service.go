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

package resources

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"sort"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/knative/serving/pkg/network"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	istio "knative.dev/pkg/apis/istio/common/v1alpha1"
	networking "knative.dev/pkg/apis/istio/v1alpha3"
)

const (
	ManagedByLabel        = "app.kubernetes.io/managed-by"
	KnativeIngressGateway = "knative-ingress-gateway.knative-serving.svc.cluster.local"
	GatewayHost           = "istio-ingressgateway.istio-system.svc.cluster.local"
)

// MakeVirtualServiceLabels creates Labels that can be used to tie a
// VirtualService to a Route.
func MakeVirtualServiceLabels(spec v1alpha1.RouteSpecFields) map[string]string {
	return map[string]string{
		v1alpha1.ManagedByLabel: "kf",
		v1alpha1.ComponentLabel: "virtualservice",
		v1alpha1.RouteHostname:  spec.Hostname,
		v1alpha1.RouteDomain:    spec.Domain,
	}
}

// MakeVirtualService creates a VirtualService from a Route object.
func MakeVirtualService(claims []*v1alpha1.RouteClaim, routes []*v1alpha1.Route) (*networking.VirtualService, error) {
	if len(claims) == 0 {
		return nil, errors.New("claims must not be empty")
	}

	namespace := claims[0].Namespace
	hostname := claims[0].Spec.RouteSpecFields.Hostname
	domain := claims[0].Spec.RouteSpecFields.Domain
	labels := MakeVirtualServiceLabels(claims[0].Spec.RouteSpecFields)

	hostDomain := domain
	if hostname != "" {
		hostDomain = hostname + "." + domain
	}

	// Build map of paths to set of bound apps
	pathApps := buildPathApps(claims, routes)
	httpRoutes, err := buildHTTPRoutes(hostDomain, pathApps, namespace)
	if err != nil {
		return nil, err
	}

	return &networking.VirtualService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "VirtualService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: v1alpha1.GenerateName(
				hostname,
				domain,
			),
			Namespace: v1alpha1.KfNamespace,
			Labels:    labels,
			Annotations: map[string]string{
				"domain":   domain,
				"hostname": hostname,
				"space":    namespace,
			},
		},
		Spec: networking.VirtualServiceSpec{
			Gateways: []string{KnativeIngressGateway},
			Hosts:    []string{hostDomain},
			HTTP:     httpRoutes,
		},
	}, nil
}

// Create HTTP routes for all paths with the same host + domain.
// Paths that do not have an app bound to them will return a 503 when a request is sent to that path.
func buildHTTPRoutes(hostDomain string, pathApps map[string]sets.String, namespace string) ([]networking.HTTPRoute, error) {
	var httpRoutes []networking.HTTPRoute

	for path, apps := range pathApps {
		var httpRoute networking.HTTPRoute

		pathMatchers, err := buildPathMatchers(path)
		if err != nil {
			return nil, err
		}

		if apps.Len() == 0 {
			// no apps bound to this path, return http route with fault for path
			httpRoute = networking.HTTPRoute{
				Match: pathMatchers,
				Fault: &networking.HTTPFaultInjection{
					Abort: &networking.InjectAbort{
						Percent:    100,
						HTTPStatus: http.StatusServiceUnavailable,
					},
				},
				Route: buildDefaultRouteDestination(),
			}
		} else {
			// create HTTP route for path with app(s) bound
			httpRoute = networking.HTTPRoute{
				Match:   pathMatchers,
				Route:   buildRouteDestinations(apps.List(), namespace),
				Headers: buildForwardingHeaders(hostDomain),
			}
		}
		httpRoutes = append(httpRoutes, httpRoute)
	}

	// Sort by reverse to defer to the longest matchers.
	// Routing rules are evaluated in order from first to last, where the first rule is given highest priority.
	sort.Sort(sort.Reverse(v1alpha1.HTTPRoutes(httpRoutes)))

	return httpRoutes, nil
}

// Hostname + domain + path combos with bound app(s) have a custom route destination for each path.
// The request is sent back to the istio ingress gateway with the host set as the app's internal host name.
// If there are multiple apps bound to a route, the traffic is split uniformly across the apps.
func buildRouteDestinations(appNames []string, namespace string) []networking.HTTPRouteDestination {
	routeWeights := getRouteWeights(len(appNames))
	routeDestinations := []networking.HTTPRouteDestination{}

	for i, app := range appNames {
		routeDestination := networking.HTTPRouteDestination{
			Destination: networking.Destination{
				Host: GatewayHost,
			},
			Headers: buildHostHeader(app, namespace),
			Weight:  routeWeights[i],
		}
		routeDestinations = append(routeDestinations, routeDestination)
	}

	return routeDestinations
}

// Hostname + domain + path combos without an app are given the default route destination,
// which simply redirects the request back to the istio ingress gateway
func buildDefaultRouteDestination() []networking.HTTPRouteDestination {
	return []networking.HTTPRouteDestination{
		{
			Destination: networking.Destination{
				Host: GatewayHost,
			},
			Weight: 100,
		},
	}
}

// getRouteWeights generates a list of integer percentages for route weights that sum to 100.
// If the number of routes does not evenly divide 100, the weights are calculated as follows:
// Round all the weights down, find the difference between that sum and 100, then distribute
// the difference among the weights.
//
// e.g. if numRoutes = 6, then 100/6 = 16.666, which rounds down to 16, with a remainder of 100 % 6 = 4.
// The final percentages would be [17, 17, 17, 17, 16, 16].
func getRouteWeights(numRoutes int) []int {
	weights := make([]int, numRoutes)
	uniformRouteWeight := 100 / numRoutes // round down
	remainder := 100 % numRoutes

	for i := 0; i < numRoutes; i++ {
		weights[i] = uniformRouteWeight
		if i < remainder {
			weights[i]++
		}
	}

	return weights
}

// buildPathMatchers creates regex matchers for a given route path.
// These matchers are used in the virtual service to determine which path a request was sent to
func buildPathMatchers(urlPath string) ([]networking.HTTPMatchRequest, error) {
	urlPath = path.Join("/", urlPath, "/")
	regexpPath, err := v1alpha1.BuildPathRegexp(urlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to convert path to regexp: %s", err)
	}

	return []networking.HTTPMatchRequest{
		{
			URI: &istio.StringMatch{
				Regex: regexpPath,
			},
		},
	}, nil
}

// buildForwardingHeaders sets forwarding headers so the app gets the real hostname it's serving
// at rather than the internal one (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Forwarded1)
func buildForwardingHeaders(hostDomain string) *networking.Headers {
	return &networking.Headers{
		Request: &networking.HeaderOperations{
			Add: map[string]string{
				"X-Forwarded-Host": hostDomain,
				"Forwarded":        fmt.Sprintf("host=%s", hostDomain),
			},
		},
	}
}

// buildHostHeader sets the host of the request (e.g. myapp.namespace.svc.cluster.local) to the app's internal host name
func buildHostHeader(appName, namespace string) *networking.Headers {
	return &networking.Headers{
		Request: &networking.HeaderOperations{
			Set: map[string]string{
				"Host": network.GetServiceHostname(appName, namespace),
			},
		},
	}
}

// buildPathApps creates a map of route paths to the apps bound to those paths (represented as a set).
func buildPathApps(claims []*v1alpha1.RouteClaim, routes []*v1alpha1.Route) map[string]sets.String {
	pathApps := make(map[string]sets.String)

	for _, claim := range claims {
		pathApps[claim.Spec.RouteSpecFields.Path] = sets.NewString()
	}

	for _, route := range routes {
		path := route.Spec.RouteSpecFields.Path
		// only add apps to route if route claim exists
		if _, exists := pathApps[path]; exists {
			pathApps[path].Insert(route.Spec.AppName)
		}
	}

	return pathApps
}
