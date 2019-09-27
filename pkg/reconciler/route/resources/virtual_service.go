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
	"fmt"
	"net/http"
	"path"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/knative/serving/pkg/network"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
func MakeVirtualService(fields v1alpha1.RouteSpecFields, routes []*v1alpha1.Route) (*networking.VirtualService, error) {
	hostname := fields.Hostname
	domain := fields.Domain
	urlPath := fields.Path
	labels := MakeVirtualServiceLabels(fields)

	hostDomain := domain
	if hostname != "" {
		hostDomain = hostname + "." + domain
	}

	httpRoutes, err := buildHTTPRoute(hostDomain, urlPath, routes)
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
			},
		},
		Spec: networking.VirtualServiceSpec{
			Gateways: []string{KnativeIngressGateway},
			Hosts:    []string{hostDomain},
			HTTP:     httpRoutes,
		},
	}, nil
}

func buildHTTPRoute(hostDomain, urlPath string, routes []*v1alpha1.Route) ([]networking.HTTPRoute, error) {
	var pathMatchers []networking.HTTPMatchRequest
	urlPath = path.Join("/", urlPath, "/")
	regexpPath, err := v1alpha1.BuildPathRegexp(urlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to convert path to regexp: %s", err)
	}

	if urlPath != "" {
		pathMatchers = append(pathMatchers, networking.HTTPMatchRequest{
			URI: &istio.StringMatch{
				Regex: regexpPath,
			},
		})
	}

	if len(routes) != 0 {
		return []networking.HTTPRoute{
			{
				Match: pathMatchers,
				Route: buildRouteDestinations(routes),
				Headers: &networking.Headers{
					Request: &networking.HeaderOperations{
						Add: map[string]string{
							// Set forwarding headers so the app gets the real hostname it's serving
							// at rather than the internal one:
							// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Forwarded1
							"X-Forwarded-Host": hostDomain,
							"Forwarded":        fmt.Sprintf("host=%s", hostDomain),
						},
					},
				},
			},
		}, nil
	}

	// If there are no apps mapped to this hostdomain+path combo, return HTTP route with fault
	return []networking.HTTPRoute{
		{
			Match: pathMatchers,
			Fault: &networking.HTTPFaultInjection{
				Abort: &networking.InjectAbort{
					Percent:    100,
					HTTPStatus: http.StatusServiceUnavailable,
				},
			},
			Route: buildDefaultRouteDestination(),
		},
	}, nil
}

func buildRouteDestinations(routes []*v1alpha1.Route) []networking.HTTPRouteDestination {

	routeWeights := getRouteWeights(len(routes))
	routeDestinations := []networking.HTTPRouteDestination{}
	for i, route := range routes {
		namespace := route.GetNamespace()
		appName := route.Spec.AppName
		routeDestination := networking.HTTPRouteDestination{
			Destination: networking.Destination{
				Host: GatewayHost,
			},
			Headers: &networking.Headers{
				Request: &networking.HeaderOperations{
					Set: map[string]string{
						"Host": network.GetServiceHostname(appName, namespace),
					},
				},
			},
			Weight: routeWeights[i],
		}
		routeDestinations = append(routeDestinations, routeDestination)
	}

	return routeDestinations
}

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
func getRouteWeights(numRoutes int) []int {
	var weights []int
	uniformRouteWeight := 100 / numRoutes // round down

	for i := 0; i < numRoutes; i++ {
		weights[i] = uniformRouteWeight
	}

	remainder := 100 - (uniformRouteWeight * numRoutes)

	routeIndex := 0
	for remainder > 0 {
		weights[routeIndex]++
		remainder--
		routeIndex++
	}

	return weights
}
