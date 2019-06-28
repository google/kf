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
	"encoding/base64"
	"hash/crc64"
	"net/http"
	"path"
	"strings"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	istio "knative.dev/pkg/apis/istio/common/v1alpha1"
	networking "knative.dev/pkg/apis/istio/v1alpha3"
	"knative.dev/pkg/kmeta"
)

const (
	KnativeIngressGateway = "knative-ingress-gateway.knative-serving.svc.cluster.local"
	GatewayHost           = "istio-ingressgateway.istio-system.svc.cluster.local"
)

// VirtualServiceName gets the name of a VirtualService given the route.
func VirtualServiceName(hostname, domain, urlPath string) string {
	hasher := crc64.New(crc64.MakeTable(crc64.ECMA))
	return hostname + "-" +
		strings.ToLower(
			base64.RawURLEncoding.EncodeToString(
				hasher.Sum([]byte(hostname+domain+path.Join("/", urlPath))),
			),
		)
}

// MakeVirtualService creates a VirtualService from a Route object.
func MakeVirtualService(route *v1alpha1.Route) (*networking.VirtualService, error) {
	hostDomain := route.Spec.Domain
	if route.Spec.Hostname != "" {
		hostDomain = route.Spec.Hostname + "." + route.Spec.Domain
	}

	var pathMatchers []networking.HTTPMatchRequest

	urlPath := path.Join("/", route.Spec.Path)
	if route.Spec.Path != "" {
		pathMatchers = append(pathMatchers, networking.HTTPMatchRequest{
			URI: &istio.StringMatch{
				Prefix: urlPath,
			},
		})
	}

	return &networking.VirtualService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "VirtualService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      VirtualServiceName(route.Spec.Hostname, route.Spec.Domain, route.Spec.Path),
			Namespace: route.GetNamespace(),
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(route),
			},
			Labels: route.GetLabels(),
			Annotations: map[string]string{
				"domain":   route.Spec.Domain,
				"hostname": route.Spec.Hostname,
				"path":     urlPath,
			},
		},
		Spec: networking.VirtualServiceSpec{
			Gateways: []string{KnativeIngressGateway},
			Hosts:    []string{hostDomain},
			HTTP: []networking.HTTPRoute{
				{
					Match: pathMatchers,
					Route: []networking.HTTPRouteDestination{
						{
							Destination: networking.Destination{
								Host: GatewayHost,

								// XXX: If this is not included, then
								// we get an error back from the
								// server suggesting we have to have a
								// port set. It doesn't seem to hurt
								// anything as we just return a fault.
								Port: networking.PortSelector{
									Number: 80,
								},
							},
							Weight: 100,
						},
					},
					Fault: &networking.HTTPFaultInjection{
						Abort: &networking.InjectAbort{
							Percent:    100,
							HTTPStatus: http.StatusServiceUnavailable,
						},
					},
				},
			},
		},
	}, nil
}
