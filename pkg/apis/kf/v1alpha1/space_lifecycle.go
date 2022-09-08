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

package v1alpha1

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/kf/dynamicutils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	clusterIngressIPDomainKey = "CLUSTER_INGRESS_IP"
	spaceNameDomainKey        = "SPACE_NAME"
)

// GetGroupVersionKind returns the GroupVersionKind.
func (r *Space) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Space")
}

// PropagateNamespaceStatus copies fields from the Namespace status to Space
// and updates the readiness based on the current phase.
func (status *SpaceStatus) PropagateNamespaceStatus(ns *v1.Namespace) {
	switch ns.Status.Phase {
	case v1.NamespaceActive:
		status.manage().MarkTrue(SpaceConditionNamespaceReady)
	case v1.NamespaceTerminating:
		status.manage().MarkFalse(SpaceConditionNamespaceReady, "Terminating", "Namespace is terminating")
	default:
		status.manage().MarkUnknown(SpaceConditionNamespaceReady, "BadPhase", "Namespace entered an unknown phase: %q", ns.Status.Phase)
	}
}

// PropagateIAMPolicyStatus updates the readiness based on the current phase.
func (status *SpaceStatus) PropagateIAMPolicyStatus(ctx context.Context, u *unstructured.Unstructured) {
	switch dynamicutils.CheckCondtions(ctx, u) {
	case corev1.ConditionUnknown:
		status.manage().MarkUnknown(SpaceConditionIAMPolicyReady, "BadPhase", "IAM Policy is not yet ready")
	case corev1.ConditionTrue:
		status.manage().MarkTrue(SpaceConditionIAMPolicyReady)
	default:
		status.manage().MarkFalse(SpaceConditionIAMPolicyReady, "BadPhase", "IAM Policy is unhealthy")
	}
}

// PropagateIngressGatewayStatus copies the list of ingresses to the status
// and updates the condition. There must be at least one externally reachable
// ingress for the space to become healthy.
func (status *SpaceStatus) PropagateIngressGatewayStatus(ingresses []corev1.LoadBalancerIngress) {
	status.IngressGateways = ingresses

	if len(status.IngressGateways) == 0 {
		status.IngressGatewayCondition().MarkReconciliationPending()
	} else {
		status.IngressGatewayCondition().MarkSuccess()
	}
}

// PropagateRuntimeConfigStatus copies the application runtime settings to the
// space status.
func (status *SpaceStatus) PropagateRuntimeConfigStatus(runtimeConfig SpaceSpecRuntimeConfig, cfg *config.Config) {
	// Copy environment over wholesale because there are no values that can be set
	// cluster-wide.
	status.RuntimeConfig = SpaceStatusRuntimeConfig{} // clear out anything old
	status.RuntimeConfig.Env = runtimeConfig.Env

	// Copy from the config if possible.
	defaultsConfig, err := cfg.Defaults()
	if err != nil {
		status.RuntimeConfigCondition().MarkReconciliationError("NilConfig", err)
		return
	}

	status.RuntimeConfig.AppCPUMin = defaultsConfig.AppCPUMin
	status.RuntimeConfig.AppCPUPerGBOfRAM = defaultsConfig.AppCPUPerGBOfRAM
	status.RuntimeConfig.ProgressDeadlineSeconds = defaultsConfig.ProgressDeadlineSeconds

	status.RuntimeConfigCondition().MarkSuccess()
}

// PropagateNetworkConfigStatus copies the application networking settings to the
// space status.
func (status *SpaceStatus) PropagateNetworkConfigStatus(networkConfig SpaceSpecNetworkConfig, cfg *config.Config, spaceName string) {
	status.NetworkConfig = SpaceStatusNetworkConfig{} // clear out anything old
	if cfg == nil {
		status.NetworkConfigCondition().MarkReconciliationError("NilConfig", errors.New("the Kf defaults configmap couldn't be found"))
		return
	}

	defaultsConfig, err := cfg.Defaults()
	if err != nil {
		status.NetworkConfigCondition().MarkReconciliationError("NilConfig", err)
		return
	}

	// Join domains in the order space -> custom then deduplicate to ensure
	// space config takes precedence over global configuration
	{
		var domains []SpaceDomain

		// Space configured domains take precedence over cluster-wide ones.
		for _, domain := range networkConfig.Domains {
			// Deep copy to prevent accidentally modifying a cached value
			domains = append(domains, *domain.DeepCopy())
		}

		for _, defaultDomain := range defaultsConfig.SpaceClusterDomains {
			domains = append(domains, SpaceDomain{
				Domain:      defaultDomain.Domain,
				GatewayName: defaultDomain.GatewayName,
			})
		}

		// Replace variables
		replacements := make(map[string]string)
		replacements[spaceNameDomainKey] = spaceName

		if ipPtr := status.FindIngressIP(); ipPtr != nil {
			replacements[clusterIngressIPDomainKey] = *ipPtr
		}

		for i := range domains {
			domains[i].Domain = applyDomainReplacements(domains[i].Domain, replacements)
		}

		status.NetworkConfig.Domains = StableDeduplicateSpaceDomainList(domains)
	}

	// Ensure all domains have a gatewayName
	{
		var domains []SpaceDomain
		for _, d := range status.NetworkConfig.Domains {
			if d.GatewayName == "" {
				domains = append(domains, SpaceDomain{
					Domain:      d.Domain,
					GatewayName: KfExternalIngressGateway,
				})
			} else {
				domains = append(domains, d)
			}
		}
		status.NetworkConfig.Domains = domains
	}

	status.NetworkConfigCondition().MarkSuccess()
}

func applyDomainReplacements(input string, replacements map[string]string) string {
	var oldnew []string
	for k, v := range replacements {
		oldnew = append(oldnew, fmt.Sprintf("$(%s)", k), v)
	}

	replacer := strings.NewReplacer(oldnew...)
	return replacer.Replace(input)
}

// PropagateBuildConfigStatus copies the application build settings to the
// space status.
func (status *SpaceStatus) PropagateBuildConfigStatus(spaceSpec SpaceSpec, cfg *config.Config) {
	status.BuildConfig = SpaceStatusBuildConfig{} // clear out anything old
	if cfg == nil {
		status.BuildConfigCondition().MarkReconciliationError("NilConfig", errors.New("the kf defaults configmap couldn't be found"))
		return
	}

	configDefaults, err := cfg.Defaults()
	if err != nil {
		status.BuildConfigCondition().MarkReconciliationError("NilConfig", err)
		return
	}

	status.BuildConfig.BuildpacksV2 = configDefaults.SpaceBuildpacksV2.WithoutDisabled()
	status.BuildConfig.StacksV2 = configDefaults.SpaceStacksV2
	status.BuildConfig.StacksV3 = configDefaults.SpaceStacksV3
	status.BuildConfig.Env = spaceSpec.BuildConfig.Env
	status.BuildConfig.ContainerRegistry = spaceSpec.BuildConfig.ContainerRegistry
	status.BuildConfig.ServiceAccount = spaceSpec.BuildConfig.ServiceAccount

	if spaceSpec.BuildConfig.DefaultToV3Stack != nil {
		status.BuildConfig.DefaultToV3Stack = *spaceSpec.BuildConfig.DefaultToV3Stack
	} else {
		status.BuildConfig.DefaultToV3Stack = configDefaults.SpaceDefaultToV3Stack
	}

	status.BuildConfigCondition().MarkSuccess()
}
