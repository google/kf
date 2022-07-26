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

package networking

import (
	"fmt"
	"log"

	semver "github.com/hashicorp/go-version"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

const (
	IstioNamespace = "istio-system"

	// IstioInjectionLabel is the label used to tell Istio to inject the
	// sidecar proxy. It is also used on the Istio deployment to show what
	// version of Istio is installed.
	IstioInjectionLabel = "istio.io/rev"

	// istioAppLabel is an istio specific label applied to istio objects.
	istioAppLabel = "app"

	// istioControlPlaneAppLabel is the "app" label value applied to istio
	// control plane objects.
	istioControlPlaneAppLabel = "istiod"

	// istioOperatorVersionLabel is the label used by Istio to label each
	// created object with the semver of Istio.
	istioOperatorVersionLabel = "operator.istio.io/version"

	// istioManagedASMConfigMap is the ConfigMap whose presence indicates a
	// managed ASM install.
	istioManagedASMConfigMap = "istio-asm-managed"

	// istioManagedASMRev is the "istio.io/rev" injection label value used for
	// managed ASM.
	istioManagedASMRev = "asm-managed"
)

var istioControlPlaneSelector labels.Selector

func init() {
	istioInjectionReq, err := labels.NewRequirement(istioAppLabel, selection.Equals, []string{istioControlPlaneAppLabel})
	if err != nil {
		log.Panicf("failed to instantiate Requirement: %v", err)
	}

	istioControlPlaneReq, err := labels.NewRequirement(IstioInjectionLabel, selection.Exists, []string{})
	if err != nil {
		log.Panicf("failed to instantiate Requirement: %v", err)
	}

	istioOperatorVersionReq, err := labels.NewRequirement(istioOperatorVersionLabel, selection.Exists, []string{})
	if err != nil {
		log.Panicf("failed to instantiate Requirement: %v", err)
	}

	istioControlPlaneSelector = labels.NewSelector().
		Add(*istioInjectionReq).
		Add(*istioControlPlaneReq).
		Add(*istioOperatorVersionReq)
}

// LookupASMRev fetches the revision label from the ASM/Istio control plane or
// sets to "asm-managed" if managed ASM is detected.
//
// If there are multiple istio control planes we use the one with the latest
// semver in the istioOperatorVersionLabel. If managed ASM is detected we
// always return the managed revision.
func LookupASMRev(
	listDeployments func(namespace string, selector labels.Selector) ([]*v1.Deployment, error),
	getConfigMap func(namespace, configMapName string) (*corev1.ConfigMap, error),
) (string, error) {
	isManaged, err := IsManagedASM(getConfigMap)
	if err == nil && isManaged {
		return istioManagedASMRev, nil
	}

	deployments, err := listDeployments(IstioNamespace, istioControlPlaneSelector)
	if err != nil {
		return "", fmt.Errorf("failed to list deployments: %v", err)
	}

	var latestDeployment *v1.Deployment
	var latestVersion *semver.Version
	for _, deployment := range deployments {
		versionLabel, ok := deployment.GetLabels()[istioOperatorVersionLabel]
		if !ok {
			continue
		}
		version, err := semver.NewVersion(versionLabel)
		if err != nil {
			continue
		}
		if latestVersion == nil || version.GreaterThan(latestVersion) {
			latestVersion = version
			latestDeployment = deployment
		}
	}

	if latestDeployment == nil {
		return "", fmt.Errorf("failed to find istio deployment with selector %q", istioControlPlaneSelector)
	}

	if revision, ok := latestDeployment.GetLabels()[IstioInjectionLabel]; !ok {
		return "", fmt.Errorf("failed to find istio deployment with selector %q", istioControlPlaneSelector)
	} else {
		return revision, nil
	}
}

// IsManagedASM checks if the cluster is using the Google-managed control plane
// installation of ASM.
//
// Managed control plane installations of ASM create a ConfigMap named
// "istio-asm-managed" in the "istio-system" namespace. The presence of that
// ConfigMap is used to determine if the cluster is using managed ASM.
func IsManagedASM(getConfigMap func(namespace, configMapName string) (*corev1.ConfigMap, error)) (bool, error) {
	configMap, err := getConfigMap(IstioNamespace, istioManagedASMConfigMap)
	switch {
	case apierrors.IsNotFound(err):
		return false, nil
	case err != nil:
		return false, err
	default:
		return configMap != nil, nil
	}
}
