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

package doctor

import (
	"context"
	"fmt"
	"strings"

	kfv1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/apis/networking"
	v1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
)

const (
	// IstioSidecarWebhookName contains the name of Istio's mutating admission
	// webhook that injects sidecars.
	IstioSidecarWebhookName = "sidecar-injector.istio.io"

	// IstioNamespaceSidecarWebhookNameSuffix contains the common name suffix of
	// Istio's mutating admission webhook that injects sidecars for namespaces.
	IstioNamespaceSidecarWebhookNameSuffix = "namespace.sidecar-injector.istio.io"

	// IstioMutatingWebhookConfigurationPrefix is the prefix of Istio MutatingWebhookConfigurations
	IstioMutatingWebhookConfigurationPrefix = "istio-sidecar-injector"
)

// NewIstioDiagnostic creates a new IstioDiagnostic to validate the Istio
// install pointed at by the client.
func NewIstioDiagnostic(kubeClient kubernetes.Interface) *IstioDiagnostic {
	return &IstioDiagnostic{
		kubeClient: kubeClient,
	}
}

// IstioDiagnostic runs tests against Istio.
type IstioDiagnostic struct {
	kubeClient kubernetes.Interface
}

var _ Diagnosable = (*IstioDiagnostic)(nil)

// Diagnose validates the installed Istio configuration.
func (u *IstioDiagnostic) Diagnose(ctx context.Context, d *Diagnostic) {
	d.Run(ctx, "Injection", func(ctx context.Context, d *Diagnostic) {
		diagnoseIstioInjectionLabel(ctx, d, u.kubeClient)
	})
}

// diagnoseIstioInjectionLabel ensures that the labels being applied by Spaces
// on Namespaces match what Istio expects. Istio uses a mutating webhook with a
// selector that matches the Namespaces it will inject sidecars onto. The
// MutatingWebhookConfiguration resource doesn't have a deterministic name, but
// it is labeled with the "istio.io/rev" revision and has one webhook listed on
// it with the value specified by IstioSidecarWebhookName or
// suffixed with IstioNamespaceSidecarWebhookNameSuffix.
func diagnoseIstioInjectionLabel(ctx context.Context, d *Diagnostic, kubeClient kubernetes.Interface) {
	selector, err := getIstioWebhookSelector(ctx, kubeClient)
	if err != nil {
		d.Fatal(err)
		return
	}

	d.Run(ctx, "Namespace", func(ctx context.Context, d *Diagnostic) {
		managedByKfSelector := labels.NewSelector()
		managedByKfSelector = managedByKfSelector.Add(kfv1alpha1.ManagedByKfRequirement())

		kfNamespaces, err := kubeClient.
			CoreV1().
			Namespaces().
			List(ctx, metav1.ListOptions{LabelSelector: managedByKfSelector.String()})
		if err != nil {
			d.Fatalf("couldn't list Kf Namespaces: %v", err)
			return
		}

		for _, namespace := range kfNamespaces.Items {
			nsLabelSet := labels.Set(namespace.Labels)

			d.Run(ctx, namespace.Name, func(ctx context.Context, d *Diagnostic) {
				if !selector.Matches(nsLabelSet) {
					d.Errorf(
						"Namespace %q labels %q don't match Istio injection selector: %s.",
						namespace.Name,
						nsLabelSet,
						selector,
					)
				}
			})
		}
	})
}

func lookupASMRev(ctx context.Context, kubernetes kubernetes.Interface) (string, error) {
	return networking.LookupASMRev(func(namespace string, selector labels.Selector) ([]*appsv1.Deployment, error) {
		deployments, err := kubernetes.
			AppsV1().
			Deployments(namespace).
			List(ctx, metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			return nil, err
		}
		var results []*appsv1.Deployment
		for _, deployment := range deployments.Items {
			dp := deployment
			results = append(results, &dp)
		}
		return results, nil
	}, func(namespace, configMapName string) (*corev1.ConfigMap, error) {
		return kubernetes.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
	})
}

func getIstioWebhookSelector(ctx context.Context, kubernetes kubernetes.Interface) (labels.Selector, error) {
	asmRev, err := lookupASMRev(ctx, kubernetes)
	if err != nil {
		return nil, err
	}
	revisionSelector := labels.NewSelector()
	req, err := labels.NewRequirement(networking.IstioInjectionLabel, selection.Equals, []string{asmRev})
	if err != nil {
		return nil, err
	}
	revisionSelector = revisionSelector.Add(*req)

	configs, err := kubernetes.
		AdmissionregistrationV1().
		MutatingWebhookConfigurations().
		List(ctx, metav1.ListOptions{LabelSelector: revisionSelector.String()})

	switch {
	case err != nil:
		return nil, err
	case len(configs.Items) == 0:
		return nil, fmt.Errorf("could not find webhook for current asm revision %q", asmRev)
	}

	var config v1.MutatingWebhookConfiguration
	for _, config = range configs.Items {
		if strings.HasPrefix(config.Name, IstioMutatingWebhookConfigurationPrefix) {
			break
		}
	}
	if len(config.Name) == 0 {
		return nil, fmt.Errorf("could not find any MutatingWebhookConfiguration that has Istio prefix")
	}

	for _, webhook := range config.Webhooks {
		if webhook.Name != IstioSidecarWebhookName && !strings.HasSuffix(webhook.Name, IstioNamespaceSidecarWebhookNameSuffix) {
			continue
		}
		if webhook.NamespaceSelector == nil {
			return nil, fmt.Errorf("webhook had nil namespace selector")
		}
		selector, err := metav1.LabelSelectorAsSelector(webhook.NamespaceSelector)
		if err != nil {
			return nil, err
		}
		return selector, nil
	}

	return nil, fmt.Errorf("could not find injection webhook in mutating webhook config %q", config.Name)
}
