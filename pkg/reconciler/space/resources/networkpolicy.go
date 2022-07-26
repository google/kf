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

package resources

import (
	"context"
	"fmt"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

func networkPolicyName(policyType string) string {
	return fmt.Sprintf("space-%s-policy", policyType)
}

// MakeAppNetworkPolicy creates a network policy targeting apps.
func MakeAppNetworkPolicy(space *v1alpha1.Space) (*networkingv1.NetworkPolicy, error) {
	policyType := v1alpha1.NetworkPolicyApp
	policy := space.Spec.NetworkConfig.AppNetworkPolicy

	return makeNetworkPolicy(space, policyType, policy)
}

// MakeBuildNetworkPolicy creates a network policy targeting builds.
func MakeBuildNetworkPolicy(space *v1alpha1.Space) (*networkingv1.NetworkPolicy, error) {
	policyType := v1alpha1.NetworkPolicyBuild
	policy := space.Spec.NetworkConfig.BuildNetworkPolicy

	return makeNetworkPolicy(space, policyType, policy)
}

func makeNetworkPolicy(space *v1alpha1.Space, policyLabelValue string, policy v1alpha1.SpaceSpecNetworkConfigPolicy) (*networkingv1.NetworkPolicy, error) {
	// Set defaults to upgrade old policies
	policy.SetDefaults(context.Background())

	ingressRule, err := makeIngressRule(policy)
	if err != nil {
		return nil, err
	}

	egressRule, err := makeEgressRule(policy)
	if err != nil {
		return nil, err
	}

	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      networkPolicyName(policyLabelValue),
			Namespace: space.Name,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(space),
			},
			Labels: map[string]string{
				managedByLabel: "kf",
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeEgress,
				networkingv1.PolicyTypeIngress,
			},
			PodSelector: *metav1.SetAsLabelSelector(map[string]string{
				v1alpha1.NetworkPolicyLabel: policyLabelValue,
			}),
			Ingress: ingressRule,
			Egress:  egressRule,
		},
	}, nil
}

func makeIngressRule(policy v1alpha1.SpaceSpecNetworkConfigPolicy) ([]networkingv1.NetworkPolicyIngressRule, error) {
	switch pt := policy.Ingress; pt {
	case v1alpha1.PermitAllNetworkPolicy:
		return []networkingv1.NetworkPolicyIngressRule{
			{}, // A single empty policy allows everything
		}, nil

	case v1alpha1.DenyAllNetworkPolicy:
		return []networkingv1.NetworkPolicyIngressRule{
			// Empty list denies all traffic
		}, nil

	default:
		return nil, fmt.Errorf("unknown ingress type: %s", pt)
	}
}

func makeEgressRule(policy v1alpha1.SpaceSpecNetworkConfigPolicy) ([]networkingv1.NetworkPolicyEgressRule, error) {
	switch pt := policy.Egress; pt {
	case v1alpha1.PermitAllNetworkPolicy:
		return []networkingv1.NetworkPolicyEgressRule{
			{}, // A single empty policy allows everything
		}, nil

	case v1alpha1.DenyAllNetworkPolicy:
		return []networkingv1.NetworkPolicyEgressRule{
			// Empty list denies all traffic
		}, nil

	default:
		return nil, fmt.Errorf("unknown egress type: %s", pt)
	}
}
