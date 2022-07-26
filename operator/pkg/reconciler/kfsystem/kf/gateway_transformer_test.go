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

package kf

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	. "kf-operator/pkg/apis/kfsystem/v1alpha1"
)

func TestGatewayTransform(t *testing.T) {
	tests := []struct {
		name                string
		gatewayName         string
		in                  map[string]string
		ingressGateway      IstioGatewayOverride
		clusterLocalGateway IstioGatewayOverride
		expected            map[string]string
	}{{
		name:        "UpdatesIngressGateway",
		gatewayName: "external-gateway",
		in: map[string]string{
			"istio": "old-istio",
		},
		ingressGateway: IstioGatewayOverride{
			Selector: map[string]string{
				"istio": "kf-ingress",
			},
		},
		clusterLocalGateway: IstioGatewayOverride{
			Selector: map[string]string{
				"istio": "cluster-local",
			},
		},
		expected: map[string]string{
			"istio": "kf-ingress",
		},
	}, {
		name:        "UpdatesClusterLocalGateway",
		gatewayName: "internal-gateway",
		in: map[string]string{
			"istio": "old-istio",
		},
		ingressGateway: IstioGatewayOverride{
			Selector: map[string]string{
				"istio": "kf-ingress",
			},
		},
		clusterLocalGateway: IstioGatewayOverride{
			Selector: map[string]string{
				"istio": "cluster-local",
			},
		},
		expected: map[string]string{
			"istio": "cluster-local",
		},
	}, {
		name:        "DoesNothingToOtherGateway",
		gatewayName: "not-kf-ingress-gateway",
		in: map[string]string{
			"istio": "old-istio",
		},
		ingressGateway: IstioGatewayOverride{
			Selector: map[string]string{
				"istio": "kf-ingress",
			},
		},
		clusterLocalGateway: IstioGatewayOverride{
			Selector: map[string]string{
				"istio": "cluster-local",
			},
		},
		expected: map[string]string{
			"istio": "old-istio",
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gateway := makeUnstructuredGateway(t, tt.gatewayName, tt.in)

			spec := GatewaySpec{
				IngressGateway:      tt.ingressGateway,
				ClusterLocalGateway: tt.clusterLocalGateway,
			}

			GatewayTransform(context.Background(), &spec)(gateway)

			got, ok, err := unstructured.NestedStringMap(gateway.Object, "spec", "selector")

			if err != nil {
				t.Error("Got error", err)
			}

			if ok != true {
				t.Error("Cannot find selector from Gateway object")
			}

			if !cmp.Equal(got, tt.expected) {
				t.Errorf("Got = %v, want: %v, diff:\n%s", got, tt.expected, cmp.Diff(got, tt.expected))
			}
		})
	}
}

func makeUnstructuredGateway(t *testing.T, name string, selector map[string]string) *unstructured.Unstructured {
	result := &unstructured.Unstructured{}
	result.SetAPIVersion("networking.istio.io/v1alpha3")
	result.SetKind("Gateway")
	result.SetName(name)
	unstructured.SetNestedStringMap(result.Object, selector, "spec", "selector")

	return result
}
