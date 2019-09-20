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
	"github.com/google/kf/pkg/kf/algorithms"
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

	namespace := claims[0].Namespace
	hostname := claims[0].Spec.RouteSpecFields.Hostname
	domain := claims[0].Spec.RouteSpecFields.Domain
	labels := MakeVirtualServiceLabels(claims[0].Spec.RouteSpecFields)

	hostDomain := domain
	if hostname != "" {
		hostDomain = hostname + "." + domain
	}

	var (
		httpRoutes []networking.HTTPRoute
	)

	// Build up HTTP Routes
	// We'll do claims first so when we merge the Routes in (which have apps
	// associated), they will replace the claims.

	for _, route := range claims {
		urlPath := route.Spec.RouteSpecFields.Path

		httpRoute, err := buildHTTPRoute(hostDomain, namespace, urlPath, nil)
		if err != nil {
			return nil, err
		}

		httpRoutes = algorithms.Merge(
			v1alpha1.HTTPRoutes(httpRoutes),
			v1alpha1.HTTPRoutes(httpRoute),
		).(v1alpha1.HTTPRoutes)
	}

	for _, route := range routes {
		var appNames []string
		urlPath := route.Spec.RouteSpecFields.Path

		// AppNames
		if route.Spec.AppName != "" {
			appNames = append(appNames, route.Spec.AppName)
		}

		httpRoute, err := buildHTTPRoute(hostDomain, namespace, urlPath, appNames)
		if err != nil {
			return nil, err
		}

		httpRoutes = algorithms.Merge(
			v1alpha1.HTTPRoutes(httpRoutes),
			v1alpha1.HTTPRoutes(httpRoute),
		).(v1alpha1.HTTPRoutes)
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

func buildHTTPRoute(hostDomain, namespace, urlPath string, appNames []string) ([]networking.HTTPRoute, error) {
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

	var httpRoutes []networking.HTTPRoute

	for _, appName := range appNames {
		httpRoutes = append(httpRoutes, networking.HTTPRoute{
			Match: pathMatchers,
			Route: buildRouteDestination(),
			Rewrite: &networking.HTTPRewrite{
				Authority: network.GetServiceHostname(appName, namespace),
			},
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
		})
	}

	// If there aren't any services bound to the route, we just want to
	// serve a 503.
	if len(httpRoutes) == 0 {
		return []networking.HTTPRoute{
			{
				Match: pathMatchers,
				Route: buildRouteDestination(),
				Fault: &networking.HTTPFaultInjection{
					Abort: &networking.InjectAbort{
						Percent:    100,
						HTTPStatus: http.StatusServiceUnavailable,
					},
				},
			},
		}, nil
	}

	return httpRoutes, nil
}

func buildRouteDestination() []networking.HTTPRouteDestination {
	return []networking.HTTPRouteDestination{
		{
			Destination: networking.Destination{
				Host: GatewayHost,
			},
			Weight: 100,
		},
	}
}
