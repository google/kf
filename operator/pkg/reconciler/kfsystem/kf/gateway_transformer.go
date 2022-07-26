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

	mf "github.com/manifestival/manifestival"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"knative.dev/pkg/logging"

	"kf-operator/pkg/apis/kfsystem/v1alpha1"
)

// GatewayTransform transforms built-in gateways into user provided ones
func GatewayTransform(ctx context.Context, gateway *v1alpha1.GatewaySpec) mf.Transformer {
	return func(u *unstructured.Unstructured) error {
		if gateway == nil {
			return nil
		}

		logger := logging.FromContext(ctx)
		// Update the deployment with the new registry and tag
		if u.GetAPIVersion() == "networking.istio.io/v1alpha3" && u.GetKind() == "Gateway" {
			if u.GetName() == "external-gateway" {
				return updateIngressGateway(gateway.IngressGateway, u, logger)
			}
			if u.GetName() == "internal-gateway" {
				return updateIngressGateway(gateway.ClusterLocalGateway, u, logger)
			}
		}
		return nil
	}
}

func updateIngressGateway(gatewayOverrides v1alpha1.IstioGatewayOverride, u *unstructured.Unstructured, log *zap.SugaredLogger) error {
	if len(gatewayOverrides.Selector) == 0 {
		return nil
	}

	log.Debugw("Updating Gateway", "name", u.GetName(), "gatewayOverrides", gatewayOverrides)
	if err := unstructured.SetNestedStringMap(u.Object, gatewayOverrides.Selector, "spec", "selector"); err != nil {
		return err
	}

	log.Debugw("Finished conversion", "name", u.GetName(), "unstructured", u.Object)
	return nil
}
