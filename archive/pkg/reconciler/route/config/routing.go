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

package config

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

const (
	RoutingConfigName       = "config-routing"
	KnativeServingNamespace = "knative-serving"

	IngressServiceNameKey    = "ingress.servicename"
	IngressNamespaceKey      = "ingress.namespace"
	KnativeIngressGatewayKey = "knative.ingress.gateway"

	DefaultIngressServiceName    = "cluster-local-gateway"
	DefaultIngressNamespace      = "gke-system"
	DefaultKnativeIngressGateway = "gke-system-gateway"
)

// RoutingConfig contains the networking configuration defined in the
// network config map.
type RoutingConfig struct {
	// IngressServiceName is the name of the Istio Ingress svc
	IngressServiceName string

	// IngressNamespace is the K8s namespace to search for Ingresses
	IngressNamespace string

	// KnativeIngressGateway is the name of ingress gateway in knative-serving namespace
	KnativeIngressGateway string
}

// NewRoutingConfigFromConfigMap creates a RoutingConfig from the supplied ConfigMap
func NewRoutingConfigFromConfigMap(configMap *corev1.ConfigMap) (*RoutingConfig, error) {
	rc := &RoutingConfig{}

	if ingressServiceName, ok := configMap.Data[IngressServiceNameKey]; !ok {
		rc.IngressServiceName = DefaultIngressServiceName
	} else {
		rc.IngressServiceName = ingressServiceName
	}

	if ingressNamespace, ok := configMap.Data[IngressNamespaceKey]; !ok {
		rc.IngressNamespace = DefaultIngressNamespace
	} else {
		rc.IngressNamespace = ingressNamespace
	}

	if knativeIngressGateway, ok := configMap.Data[KnativeIngressGatewayKey]; !ok {
		rc.KnativeIngressGateway = DefaultKnativeIngressGateway
	} else {
		rc.KnativeIngressGateway = knativeIngressGateway
	}

	rc.KnativeIngressGateway = fmt.Sprintf("%s/%s", KnativeServingNamespace, rc.KnativeIngressGateway)
	return rc, nil
}

// GatewayHost is the name of gateway set as route destination host in the kf virtualservice
// (usually back to the ingress gateway or cluster local gateway)
func (rc *RoutingConfig) GatewayHost() string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", rc.IngressServiceName, rc.IngressNamespace)
}
