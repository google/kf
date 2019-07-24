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

package describe_test

import (
	"bytes"
	"os"
	"testing"

	kfv1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/describe"
	"github.com/google/kf/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

func ExampleEnvVars_populated() {
	env := []corev1.EnvVar{
		{Name: "FIRST", Value: "first-value"},
		{Name: "SECOND", Value: "second-value"},
	}

	describe.EnvVars(os.Stdout, env)

	// Output: Environment:
	//   FIRST:   first-value
	//   SECOND:  second-value
}

func ExampleEnvVars_empty() {
	describe.EnvVars(os.Stdout, nil)

	// Output: Environment: <empty>
}

func ExampleTypeMeta() {
	s := &v1.Secret{}
	s.Kind = "Secret"
	s.APIVersion = "v1"

	describe.TypeMeta(os.Stdout, s.TypeMeta)

	// Output: API Version:  v1
	// Kind:         Secret
}

func TestObjectMeta(t *testing.T) {
	fiveDaysAgo := metav1.NewTime(metav1.Now().AddDate(0, 0, -5))
	deletionGracePeriodSeconds := int64(0)

	cases := map[string]struct {
		obj             metav1.ObjectMeta
		expectedStrings []string
	}{
		"static metadata": {
			obj: metav1.ObjectMeta{
				Name:       "my-object",
				Namespace:  "my-namespace",
				Generation: 42,
				UID:        "ed2ca418-531e-4b09-abfd-e18e66bd0e4a",
			},
			expectedStrings: []string{"my-object", "my-namespace", "42", "ed2ca418-531e-4b09-abfd-e18e66bd0e4a"},
		},
		"age delta": {
			// expect age to be 5 days
			obj: metav1.ObjectMeta{
				CreationTimestamp: fiveDaysAgo,
			},
			expectedStrings: []string{"5d"},
		},
		"terminating": {
			// expect age to be 5 days
			obj: metav1.ObjectMeta{
				DeletionTimestamp:          &fiveDaysAgo,
				DeletionGracePeriodSeconds: &deletionGracePeriodSeconds,
			},
			expectedStrings: []string{"5d", "Terminating Since"},
		},
		"labels": {
			// expect age to be 5 days
			obj: metav1.ObjectMeta{
				Labels: map[string]string{"labelkey": "labelval"},
			},
			expectedStrings: []string{"labelkey=labelval"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			b := &bytes.Buffer{}

			describe.ObjectMeta(b, tc.obj)

			testutil.AssertContainsAll(t, b.String(), tc.expectedStrings)
		})
	}
}

func TestDuckStatus(t *testing.T) {
	cases := map[string]struct {
		status          duckv1beta1.Status
		expectedStrings []string
	}{
		"ConditionReady unknown": {
			status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{
					{Status: v1.ConditionUnknown, Type: apis.ConditionReady, Message: "condition-ready-message"},
				},
			},
			expectedStrings: []string{"condition-ready-message", "unknown", "Ready:"},
		},
		"ConditionSucceeded unknown": {
			status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{
					{Status: v1.ConditionUnknown, Type: apis.ConditionSucceeded, Message: "condition-succeeded-message"},
				},
			},
			expectedStrings: []string{"condition-succeeded-message", "unknown", "Succeeded:"},
		},
		"Multiple Conditions": {
			status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{
					{Status: v1.ConditionUnknown, Type: apis.ConditionReady},
					{Status: v1.ConditionTrue, Type: "SecretReady"},
					{Status: v1.ConditionFalse, Type: "NamespaceReady"},
				},
			},
			expectedStrings: []string{"True", "False", "Unknown", "SecretReady", "NamespaceReady"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			b := &bytes.Buffer{}

			describe.DuckStatus(b, tc.status)

			testutil.AssertContainsAll(t, b.String(), tc.expectedStrings)
		})
	}
}

func ExampleDuckStatus_ready() {
	status := duckv1beta1.Status{
		Conditions: duckv1beta1.Conditions{
			{Status: v1.ConditionTrue, Type: "SecretReady"},
			{Status: v1.ConditionUnknown, Type: apis.ConditionReady, Reason: "NamespaceErr", Message: "problem with namespace"},
			{Status: v1.ConditionFalse, Type: "NamespaceReady", Reason: "NotOwned", Message: "couldn't create"},
		},
	}

	describe.DuckStatus(os.Stdout, status)

	// Output: Status:
	//   Ready:
	//     Ready:    Unknown
	//     Message:  problem with namespace
	//     Reason:   NamespaceErr
	//   Conditions:
	//     Type            Status  Updated    Message          Reason
	//     NamespaceReady  False   <unknown>  couldn't create  NotOwned
	//     SecretReady     True    <unknown>
}

func ExampleDuckStatus_succeeded() {
	status := duckv1beta1.Status{
		Conditions: duckv1beta1.Conditions{
			{Status: v1.ConditionUnknown, Type: apis.ConditionSucceeded, Reason: "NamespaceErr", Message: "problem with namespace"},
			{Status: v1.ConditionFalse, Type: "NamespaceReady", Reason: "NotOwned", Message: "couldn't create"},
		},
	}

	describe.DuckStatus(os.Stdout, status)

	// Output: Status:
	//   Succeeded:
	//     Ready:    Unknown
	//     Message:  problem with namespace
	//     Reason:   NamespaceErr
	//   Conditions:
	//     Type            Status  Updated    Message          Reason
	//     NamespaceReady  False   <unknown>  couldn't create  NotOwned
}

func ExampleAppSpecInstances_exactly() {
	exactly := 3
	instances := kfv1alpha1.AppSpecInstances{}
	instances.Stopped = true
	instances.Exactly = &exactly

	describe.AppSpecInstances(os.Stdout, instances)

	// Output: Scale:
	//   Stopped?:  true
	//   Exactly:   3
}

func ExampleAppSpecInstances_minOnly() {
	min := 3
	instances := kfv1alpha1.AppSpecInstances{}
	instances.Min = &min

	describe.AppSpecInstances(os.Stdout, instances)

	// Output: Scale:
	//   Stopped?:  false
	//   Min:       3
	//   Max:       ∞
}

func ExampleAppSpecInstances_minMax() {
	min := 3
	max := 5
	instances := kfv1alpha1.AppSpecInstances{}
	instances.Min = &min
	instances.Max = &max

	describe.AppSpecInstances(os.Stdout, instances)

	// Output: Scale:
	//   Stopped?:  false
	//   Min:       3
	//   Max:       5
}

func ExampleSourceSpec_buildpack() {
	spec := kfv1alpha1.SourceSpec{
		ServiceAccount: "builder-account",
		BuildpackBuild: kfv1alpha1.SourceSpecBuildpackBuild{
			Source:           "gcr.io/my-registry/src-mysource",
			Stack:            "cflinuxfs3",
			BuildpackBuilder: "gcr.io/my-registry/my-builder:latest",
			Registry:         "gcr.io/my-registry",
		},
	}

	describe.SourceSpec(os.Stdout, spec)

	// Output: Source:
	//   Build Type:       buildpack
	//   Service Account:  builder-account
	//   Buildpack Build:
	//     Source:    gcr.io/my-registry/src-mysource
	//     Stack:     cflinuxfs3
	//     Bulider:   gcr.io/my-registry/my-builder:latest
	//     Registry:  gcr.io/my-registry
	//     Environment: <empty>
}

func ExampleSourceSpec_docker() {
	spec := kfv1alpha1.SourceSpec{
		ServiceAccount: "builder-account",
		ContainerImage: kfv1alpha1.SourceSpecContainerImage{
			Image: "mysql/mysql",
		},
	}

	describe.SourceSpec(os.Stdout, spec)

	// Output: Source:
	//   Build Type:       container
	//   Service Account:  builder-account
	//   Container Image:
	//     Image:  mysql/mysql
}

func ExampleHealthCheck_nil() {
	describe.HealthCheck(os.Stdout, nil)

	// Output: Health Check: <empty>
}

func ExampleHealthCheck_http() {
	describe.HealthCheck(os.Stdout, &corev1.Probe{
		TimeoutSeconds: 42,
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{Path: "/healthz"},
		},
	})

	// Output: Health Check:
	//   Timeout:   42s
	//   Type:      http
	//   Endpoint:  /healthz
}

func ExampleHealthCheck_tcp() {
	describe.HealthCheck(os.Stdout, &corev1.Probe{
		TimeoutSeconds: 42,
		Handler: corev1.Handler{
			TCPSocket: &corev1.TCPSocketAction{},
		},
	})

	// Output: Health Check:
	//   Timeout:  42s
	//   Type:     port (tcp)
}
