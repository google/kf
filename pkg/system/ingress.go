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

package system

import (
	"errors"

	corev1 "k8s.io/api/core/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	listersv1 "k8s.io/client-go/listers/core/v1"
)

// clusterIngressLabels returns the set of labels expected to select the cluster
// ingress. This should match the lables in config/202-external-gateway.yaml
func clusterIngressLabels() labels.Set {
	return map[string]string{
		"istio": "ingressgateway",
	}
}

// ClusterIngressSelector matches the set of labels expected to select the cluster
// ingress. This matches the selector found in config/202-external-gateway.yaml.
func ClusterIngressSelector() labels.Selector {
	return labels.SelectorFromSet(clusterIngressLabels())
}

// ServiceLister is the subset of functions on v1.ServiceLister
// used by GetClusterIngress.
type ServiceLister interface {
	List(selector labels.Selector) (ret []*corev1.Service, err error)
}

// Assert that ServiceLister is indeed a subset of the real thing.
var _ (ServiceLister) = (listersv1.ServiceLister)(nil)

// GetClusterIngresses extracts LoadBalancerIngresses from the services matching
// the ClusterIngressSelector.
func GetClusterIngresses(lister ServiceLister) ([]corev1.LoadBalancerIngress, error) {
	services, err := lister.List(ClusterIngressSelector())
	if err != nil {
		return nil, err
	}

	var out []corev1.LoadBalancerIngress
	for _, service := range services {
		out = append(out, service.Status.LoadBalancer.Ingress...)
	}

	return out, nil
}

// ExtractProxyIngressFromList is a utility function that extracts a single
// routable ingress from the list.
func ExtractProxyIngressFromList(ingresses []corev1.LoadBalancerIngress) (string, error) {

	// NOTE: Currently this only supports IP ingresses, but could support domain
	// ingresses in the future if/when GKE supports those. With domains, extra
	// care and validation will be needed to ensure the domains are valid e.g.
	// don't have wildcards and will be resolved correctly.

	if len(ingresses) == 0 {
		return "", errors.New("no ingresses were found")
	}

	for _, ingress := range ingresses {
		if ingress.IP != "" {
			return ingress.IP, nil
		}
	}

	return "", errors.New("no ingresses had IP addresses listed")
}
