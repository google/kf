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
	"strings"

	corev1 "k8s.io/api/core/v1"
)

const (
	RoutingConfigName = "config-routing"
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

// NewConfigFromConfigMap creates a RoutingConfig from the supplied ConfigMap
func NewConfigFromConfigMap(configMap *corev1.ConfigMap) (*RoutingConfig, error) {
	nc := &RoutingConfig{}
	
}

// From network.Config

// NewConfigFromConfigMap creates a RoutingConfig from the supplied ConfigMap
// func NewConfigFromConfigMap(configMap *corev1.ConfigMap) (*RoutingConfig, error) {
// 	nc := &Config{}
// 	if ipr, ok := configMap.Data[IstioOutboundIPRangesKey]; !ok {
// 		// It is OK for this to be absent, we will elide the annotation.
// 		nc.IstioOutboundIPRanges = "*"
// 	} else if normalizedIpr, err := validateAndNormalizeOutboundIPRanges(ipr); err != nil {
// 		return nil, err
// 	} else {
// 		nc.IstioOutboundIPRanges = normalizedIpr
// 	}

// 	if ingressClass, ok := configMap.Data[DefaultClusterIngressClassKey]; !ok {
// 		nc.DefaultClusterIngressClass = IstioIngressClassName
// 	} else {
// 		nc.DefaultClusterIngressClass = ingressClass
// 	}

// 	nc.DefaultCertificateClass = CertManagerCertificateClassName
// 	if certClass, ok := configMap.Data[DefaultCertificateClassKey]; ok {
// 		nc.DefaultCertificateClass = certClass
// 	}

// 	// Blank DomainTemplate makes no sense so use our default
// 	if dt, ok := configMap.Data[DomainTemplateKey]; !ok {
// 		nc.DomainTemplate = DefaultDomainTemplate
// 	} else {
// 		t, err := template.New("domain-template").Parse(dt)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if err := checkDomainTemplate(t); err != nil {
// 			return nil, err
// 		}

// 		nc.DomainTemplate = dt
// 	}

// 	// Blank TagTemplate makes no sense so use our default
// 	if tt, ok := configMap.Data[TagTemplateKey]; !ok {
// 		nc.TagTemplate = DefaultTagTemplate
// 	} else {
// 		t, err := template.New("tag-template").Parse(tt)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if err := checkTagTemplate(t); err != nil {
// 			return nil, err
// 		}

// 		nc.TagTemplate = tt
// 	}

// 	nc.AutoTLS = strings.ToLower(configMap.Data[AutoTLSKey]) == "enabled"

// 	switch strings.ToLower(configMap.Data[HTTPProtocolKey]) {
// 	case string(HTTPEnabled):
// 		nc.HTTPProtocol = HTTPEnabled
// 	case "":
// 		// If HTTPProtocol is not set in the config-network, we set the default value
// 		// to HTTPEnabled.
// 		nc.HTTPProtocol = HTTPEnabled
// 	case string(HTTPDisabled):
// 		nc.HTTPProtocol = HTTPDisabled
// 	case string(HTTPRedirected):
// 		nc.HTTPProtocol = HTTPRedirected
// 	default:
// 		return nil, fmt.Errorf("httpProtocol %s in config-network ConfigMap is not supported", configMap.Data[HTTPProtocolKey])
// 	}
// 	return nc, nil
// }

// from Domain

// // NewDomainFromConfigMap creates a Domain from the supplied ConfigMap
// func NewDomainFromConfigMap(configMap *corev1.ConfigMap) (*Domain, error) {
// 	c := Domain{Domains: map[string]*LabelSelector{}}
// 	hasDefault := false
// 	for k, v := range configMap.Data {
// 		if k == configmap.ExampleKey {
// 			continue
// 		}
// 		labelSelector := LabelSelector{}
// 		err := yaml.Unmarshal([]byte(v), &labelSelector)
// 		if err != nil {
// 			return nil, err
// 		}
// 		c.Domains[k] = &labelSelector
// 		if len(labelSelector.Selector) == 0 {
// 			hasDefault = true
// 		}
// 	}
// 	if !hasDefault {
// 		c.Domains[DefaultDomain] = &LabelSelector{}
// 	}
// 	return &c, nil
// }
