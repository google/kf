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
	"bytes"
	"context"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func makeConfigMap(name string, namespace string) runtime.Object {
	cm := &corev1.ConfigMap{}
	cm.APIVersion = "v1"
	cm.Kind = "ConfigMap"
	cm.Name = name
	cm.Namespace = namespace

	return cm
}

func makeNamespace(name string, labels map[string]string) runtime.Object {
	ns := &corev1.Namespace{}
	ns.APIVersion = "v1"
	ns.Kind = "Namespace"
	ns.Name = name
	ns.Labels = labels

	return ns
}

func makeWebhook(name, webhookName, rev string) runtime.Object {
	wh := &admissionregistrationv1.MutatingWebhookConfiguration{}
	wh.APIVersion = "admissionregistration.k8s.io/v1"
	wh.Kind = "MutatingWebhookConfiguration"
	wh.Name = name
	wh.Labels = map[string]string{
		"istio.io/rev": rev,
	}
	wh.Webhooks = []admissionregistrationv1.MutatingWebhook{
		{
			Name: webhookName,
			NamespaceSelector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{Key: "istio-injection", Operator: metav1.LabelSelectorOpDoesNotExist},
					{Key: "istio.io/rev", Operator: metav1.LabelSelectorOpIn, Values: []string{rev}},
				},
			},
		},
	}

	return wh
}

func makeDeployment(rev, revSemver string) runtime.Object {
	d := &appsv1.Deployment{}
	d.Name = "istiod-" + rev
	d.Namespace = "istio-system"
	d.Labels = map[string]string{
		"app":                       "istiod",
		"istio.io/rev":              rev,
		"operator.istio.io/version": revSemver,
	}
	return d
}

func makeKfManagedSpace(rev string) runtime.Object {
	return makeNamespace("asm-ns"+rev, map[string]string{
		"app.kubernetes.io/managed-by": "kf",
		"istio.io/rev":                 rev,
	})
}

func makeIstioWebhook(rev string) runtime.Object {
	return makeWebhook(
		"istio-sidecar-injector-"+rev,
		"sidecar-injector.istio.io",
		rev,
	)
}

func TestDiagnoseIstioInjectionLabel(t *testing.T) {
	// The following objects are adapted from real Kf clusters; strings are
	// hard-coded to ensure this test always matches real objects.
	nonKfNs := makeNamespace("gke-system", nil)

	managedASMIstioWebhook := makeWebhook(
		"istiod-asm-managed",
		"namespace.sidecar-injector.istio.io",
		"asm-managed",
	)

	revisionedIstioWebhook := makeWebhook(
		"istiod-sidecar-injector-asm-1102-2",
		"rev.namespace.sidecar-injector.istio.io",
		"asm-1102-2",
	)

	managedASMConfigMap := makeConfigMap("istio-asm-managed", "istio-system")

	cases := map[string]struct {
		Env        []runtime.Object
		WantFailed bool
	}{
		"no istio control plane fails": {
			Env: []runtime.Object{
				nonKfNs,
			},
			WantFailed: true,
		},
		"no istio webhook fails": {
			Env: []runtime.Object{
				nonKfNs,
				makeDeployment("asm-195-2", "1.9.5-asm.2"),
			},
			WantFailed: true,
		},
		"no failure on no space": {
			Env: []runtime.Object{
				nonKfNs,
				makeDeployment("asm-173-6", "1.7.3-asm.6"),
				makeIstioWebhook("asm-173-6"),
			},
			WantFailed: false,
		},
		"no failure on correctly labeled space": {
			Env: []runtime.Object{
				nonKfNs,
				makeDeployment("asm-173-6", "1.7.3-asm.6"),
				makeIstioWebhook("asm-173-6"),
				makeKfManagedSpace("asm-173-6"),
			},
			WantFailed: false,
		},
		"space with incorrect label fails": {
			Env: []runtime.Object{
				nonKfNs,
				makeDeployment("asm-173-6", "1.7.3-asm.6"),
				makeIstioWebhook("asm-173-6"),
				makeKfManagedSpace("asm-195-2"),
			},
			WantFailed: true,
		},
		"multiple webhooks is ok": {
			Env: []runtime.Object{
				nonKfNs,
				makeDeployment("asm-195-2", "1.9.5-asm.2"),
				makeIstioWebhook("asm-173-6"),
				makeIstioWebhook("asm-195-2"),
				makeKfManagedSpace("asm-195-2"),
			},
			WantFailed: false,
		},
		"selects latest revision": {
			Env: []runtime.Object{
				nonKfNs,
				makeDeployment("asm-173-6", "1.7.3-asm.6"),
				makeDeployment("asm-1100-2", "1.10.0-asm.2"),
				makeDeployment("asm-195-2", "1.9.5-asm.2"),
				makeIstioWebhook("asm-1100-2"),
				makeKfManagedSpace("asm-1100-2"),
			},
			WantFailed: false,
		},
		"managed ASM": {
			Env: []runtime.Object{
				managedASMConfigMap,
				makeKfManagedSpace("asm-managed"),
				managedASMIstioWebhook,
			},
			WantFailed: false,
		},
		"revisioned webhook": {
			Env: []runtime.Object{
				nonKfNs,
				makeDeployment("asm-1102-2", "1.10.2-asm.2"),
				revisionedIstioWebhook,
				makeKfManagedSpace("asm-1102-2"),
			},
			WantFailed: false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			client := k8sfake.NewSimpleClientset(tc.Env...)
			buf := &bytes.Buffer{}
			diagnostic := NewDiagnostic(tn, buf)

			diagnostic.Run(context.Background(), "test", func(ctx context.Context, d *Diagnostic) {
				diagnoseIstioInjectionLabel(context.Background(), d, client)
			})

			t.Log(buf.String())
			testutil.AssertEqual(t, "failed", tc.WantFailed, diagnostic.Failed())
		})
	}
}
