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
	"github.com/gorilla/mux"
	"github.com/knative/serving/pkg/network"
	"github.com/knative/serving/pkg/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	istio "knative.dev/pkg/apis/istio/common/v1alpha1"
	networking "knative.dev/pkg/apis/istio/v1alpha3"
	"knative.dev/pkg/kmeta"
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
func MakeVirtualService(routes []*v1alpha1.Route) (*networking.VirtualService, error) {
	if len(routes) == 0 {
		return nil, errors.New("routes must not be empty")
	}

	namespace := routes[0].Namespace
	hostname := routes[0].Spec.RouteSpecFields.Hostname
	domain := routes[0].Spec.RouteSpecFields.Domain
	urlPath := routes[0].Spec.RouteSpecFields.Path
	labels := MakeVirtualServiceLabels(routes[0].Spec.RouteSpecFields)

	var (
		ownerRefs []metav1.OwnerReference
		appNames  []string
	)
	for _, route := range routes {
		// Each route will own the VirtualService. Therefore none of them can be a
		// controller.
		ownerRef := *kmeta.NewControllerRef(route)
		ownerRef.Controller = nil
		ownerRef.BlockOwnerDeletion = nil
		ownerRefs = append(ownerRefs, ownerRef)

		// AppNames
		if route.Spec.AppName != "" {
			appNames = append(appNames, route.Spec.AppName)
		}
	}

	hostDomain := domain
	if hostname != "" {
		hostDomain = hostname + "." + domain
	}

	httpRoute, err := buildHTTPRoute(namespace, urlPath, appNames)
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
			Namespace:       v1alpha1.KfNamespace,
			OwnerReferences: ownerRefs,
			Labels: resources.UnionMaps(
				labels, map[string]string{
					ManagedByLabel: "kf",
				}),
			Annotations: map[string]string{
				"domain":   domain,
				"hostname": hostname,
				"space":    namespace,
			},
		},
		Spec: networking.VirtualServiceSpec{
			Gateways: []string{KnativeIngressGateway},
			Hosts:    []string{hostDomain},
			HTTP:     httpRoute,
		},
	}, nil
}

func buildHTTPRoute(namespace, urlPath string, appNames []string) ([]networking.HTTPRoute, error) {
	var pathMatchers []networking.HTTPMatchRequest

	urlPath = path.Join("/", urlPath, "/")
	regexpPath, err := buildPathRegex(urlPath)
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

func buildPathRegex(path string) (string, error) {
	p, err := (&mux.Router{}).PathPrefix(path).GetPathRegexp()
	if err != nil {
		return "", err
	}
	return p + `(/.*)?`, nil
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
