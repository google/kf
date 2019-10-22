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
	RoutingConfigName        = "config-routing"
	IngressServiceNameKey    = "ingress.servicename"
	IngressNamespaceKey      = "ingress.namespace"
	KnativeIngressGatewayKey = "knative.ingress.gateway"
	GatewayHostKey           = "gateway.host"

	DefaultIngressServiceName    = "istio-ingress"
	DefaultIngressNamespace      = "gke-system"
	DefaultKnativeIngressGateway = "knative-serving/gke-system-gateway"
	DefaultGatewayHost           = "cluster-local-gateway.gke-system.svc.cluster.local"
)

// RoutingConfig contains the networking configuration defined in the
// network config map.
type RoutingConfig struct {
	// Name of the Istio Ingress svc
	IngressServiceName string

	// K8s namespace to search for Ingresses
	IngressNamespace string

	KnativeIngressGateway string

	GatewayHost string
}

// NewRoutingConfigFromConfigMap creates a RoutingConfig from the supplied ConfigMap
func NewRoutingConfigFromConfigMap(configMap *corev1.ConfigMap) (*RoutingConfig, error) {
	nc := &RoutingConfig{}

	fmt.Printf("configmap looks like %#v", configMap)

	if ingressServiceName, ok := configMap.Data[IngressServiceNameKey]; !ok {
		nc.IngressServiceName = DefaultIngressServiceName
	} else {
		nc.IngressServiceName = ingressServiceName
	}

	if ingressNamespace, ok := configMap.Data[IngressNamespaceKey]; !ok {
		nc.IngressNamespace = DefaultIngressNamespace
	} else {
		nc.IngressNamespace = ingressNamespace
	}

	if knativeIngressGateway, ok := configMap.Data[IngressServiceNameKey]; !ok {
		nc.KnativeIngressGateway = DefaultKnativeIngressGateway
	} else {
		nc.KnativeIngressGateway = knativeIngressGateway
	}

	if gatewayHost, ok := configMap.Data[GatewayHostKey]; !ok {
		nc.GatewayHost = DefaultGatewayHost
	} else {
		nc.GatewayHost = gatewayHost
	}

	return nc, nil
}
