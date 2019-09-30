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
func MakeVirtualService(claims []*v1alpha1.RouteClaim, routes []*v1alpha1.Route) (*networking.VirtualService, error) {
	if len(claims) == 0 {
		return nil, errors.New("claims must not be empty")
	}
	namespace := claims[0].GetNamespace()
	hostname := claims[0].Spec.RouteSpecFields.Hostname
	domain := claims[0].Spec.RouteSpecFields.Domain
	labels := MakeVirtualServiceLabels(claims[0].Spec.RouteSpecFields)

	hostDomain := domain
	if hostname != "" {
		hostDomain = hostname + "." + domain
	}

	// Get route claim paths that don't have a corresponding mapped route path
	// TODO: optimize this
	var unmappedPaths []string
	for _, claim := range claims {
		path := claim.Spec.RouteSpecFields.Path
		matchingPath := false
		for _, route := range routes {
			routePath := route.Spec.RouteSpecFields.Path
			if path == routePath {
				matchingPath = true
			}
		}
		if matchingPath == false {
			unmappedPaths = append(unmappedPaths, path)
		}
	}

	httpRoutes, err := buildHTTPRoutes(hostDomain, unmappedPaths, routes)
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

func buildHTTPRoutes(hostDomain string, unmappedPaths []string, routes []*v1alpha1.Route) ([]networking.HTTPRoute, error) {
	var httpRoutes []networking.HTTPRoute

	for _, route := range routes {
		routePath := route.Spec.RouteSpecFields.Path
		routePathMatchers, err := buildPathMatchers(routePath)
		if err != nil {
			return nil, err
		}
		newHttpRoute := networking.HTTPRoute{
			Match:   routePathMatchers,
			Route:   buildRouteDestinations(routes),
			Headers: buildForwardingHeaders(hostDomain),
		}
		httpRoutes = append(httpRoutes, newHttpRoute)
	}

	// If there are routeclaims with a path not mapped to an app, return HTTP route with fault for that path
	for _, path := range unmappedPaths {
		unmappedPathMatchers, err := buildPathMatchers(path)
		if err != nil {
			return nil, err
		}

		faultHttpRoute := networking.HTTPRoute{
			Match: unmappedPathMatchers,
			Fault: &networking.HTTPFaultInjection{
				Abort: &networking.InjectAbort{
					Percent:    100,
					HTTPStatus: http.StatusServiceUnavailable,
				},
			},
			Route: buildDefaultRouteDestination(),
		}
		httpRoutes = append(httpRoutes, faultHttpRoute)
	}

	return httpRoutes, nil
}

func buildRouteDestinations(routes []*v1alpha1.Route) []networking.HTTPRouteDestination {

	routeWeights := getRouteWeights(len(routes))
	routeDestinations := []networking.HTTPRouteDestination{}

	for i, route := range routes {
		routeDestination := networking.HTTPRouteDestination{
			Destination: networking.Destination{
				Host: GatewayHost,
			},
			Headers: buildHostHeader(route.Spec.AppName, route.GetNamespace()),
			Weight:  routeWeights[i],
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
	weights := make([]int, numRoutes)
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

func buildPathMatchers(urlPath string) ([]networking.HTTPMatchRequest, error) {
	var pathMatchers []networking.HTTPMatchRequest
	urlPath = path.Join("/", urlPath, "/")
	regexpPath, err := v1alpha1.BuildPathRegexp(urlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to convert path to regexp: %s", err)
	}

	pathMatchers = append(pathMatchers, networking.HTTPMatchRequest{
		URI: &istio.StringMatch{
			Regex: regexpPath,
		},
	})

	return pathMatchers, nil
}

func buildForwardingHeaders(hostDomain string) *networking.Headers {
	return &networking.Headers{
		Request: &networking.HeaderOperations{
			Add: map[string]string{
				// Set forwarding headers so the app gets the real hostname it's serving
				// at rather than the internal one:
				// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Forwarded1
				"X-Forwarded-Host": hostDomain,
				"Forwarded":        fmt.Sprintf("host=%s", hostDomain),
			},
		},
	}
}

func buildHostHeader(appName, namespace string) *networking.Headers {
	return &networking.Headers{
		Request: &networking.HeaderOperations{
			Set: map[string]string{
				"Host": network.GetServiceHostname(appName, namespace),
			},
		},
	}
}
