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
	"github.com/gorilla/mux"
	"github.com/knative/serving/pkg/network"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	istio "knative.dev/pkg/apis/istio/common/v1alpha1"
	networking "knative.dev/pkg/apis/istio/v1alpha3"
	"knative.dev/pkg/kmeta"
)

const (
	KnativeIngressGateway = "knative-ingress-gateway.knative-serving.svc.cluster.local"
	GatewayHost           = "istio-ingressgateway.istio-system.svc.cluster.local"
)

// MakeVirtualService creates a VirtualService from a Route object.
func MakeVirtualService(route *v1alpha1.Route) (*networking.VirtualService, error) {
	hostDomain := route.Spec.Domain
	if route.Spec.Hostname != "" {
		hostDomain = route.Spec.Hostname + "." + route.Spec.Domain
	}

	httpRoute, err := buildHTTPRoute(route)
	if err != nil {
		return nil, err
	}

	// Each route will own the VirtualService. Therefore none of them can be a
	// controller.
	ownerRef := *kmeta.NewControllerRef(route)
	ownerRef.Controller = nil
	ownerRef.BlockOwnerDeletion = nil

	return &networking.VirtualService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "VirtualService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      v1alpha1.GenerateName(route.Spec.Hostname, route.Spec.Domain),
			Namespace: v1alpha1.KfNamespace,
			OwnerReferences: []metav1.OwnerReference{
				ownerRef,
			},
			Labels: route.GetLabels(),
			Annotations: map[string]string{
				"domain":   route.Spec.Domain,
				"hostname": route.Spec.Hostname,
				"space":    route.GetNamespace(),
			},
		},
		Spec: networking.VirtualServiceSpec{
			Gateways: []string{KnativeIngressGateway},
			Hosts:    []string{hostDomain},
			HTTP:     httpRoute,
		},
	}, nil
}

func buildHTTPRoute(route *v1alpha1.Route) ([]networking.HTTPRoute, error) {
	var pathMatchers []networking.HTTPMatchRequest

	urlPath := path.Join("/", route.Spec.Path, "/")
	regexpPath, err := buildPathRegex(urlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to convert path to regexp: %s", err)
	}

	if route.Spec.Path != "" {
		pathMatchers = append(pathMatchers, networking.HTTPMatchRequest{
			URI: &istio.StringMatch{
				Regex: regexpPath,
			},
		})
	}

	var httpRoutes []networking.HTTPRoute

	for _, ksvcName := range route.Spec.KnativeServiceNames {
		httpRoutes = append(httpRoutes, networking.HTTPRoute{
			Match: pathMatchers,
			Route: buildRouteDestination(),
			Rewrite: &networking.HTTPRewrite{
				Authority: network.GetServiceHostname(ksvcName, route.GetNamespace()),
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
