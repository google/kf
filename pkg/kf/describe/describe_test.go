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

	kfv1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/describe"
	"github.com/google/kf/v2/pkg/kf/testutil"
	v1 "k8s.io/api/core/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	"knative.dev/pkg/ptr"
)

func TestLabels(t *testing.T) {
	cases := map[string]struct {
		labels map[string]string
	}{
		"Empty": {},
		"Sorted": {
			labels: map[string]string{
				"abc": "123",
				"def": "456",
				"ghi": "789",
			},
		},
		"Unsorted": {
			labels: map[string]string{
				"ghi": "789",
				"def": "456",
				"abc": "123",
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			b := &bytes.Buffer{}
			describe.Labels(b, tc.labels)
			testutil.AssertGolden(t, "labels", b.Bytes())
		})
	}
}

func TestDuckStatus(t *testing.T) {
	cases := map[string]struct {
		status duckv1beta1.Status
	}{
		"ConditionReady unknown": {
			status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{
					{Status: v1.ConditionUnknown, Type: apis.ConditionReady, Message: "condition-ready-message"},
				},
			},
		},
		"ConditionSucceeded unknown": {
			status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{
					{Status: v1.ConditionUnknown, Type: apis.ConditionSucceeded, Message: "condition-succeeded-message"},
				},
			},
		},
		"Multiple Conditions": {
			status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{
					{Status: v1.ConditionUnknown, Type: apis.ConditionReady},
					{Status: v1.ConditionTrue, Type: "SecretReady"},
					{Status: v1.ConditionFalse, Type: "NamespaceReady"},
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			b := &bytes.Buffer{}

			describe.DuckStatus(b, tc.status)
			testutil.AssertGolden(t, "status", b.Bytes())
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
	instances := kfv1alpha1.AppSpecInstances{}
	instances.Stopped = true
	instances.Replicas = ptr.Int32(3)

	describe.AppSpecInstances(os.Stdout, instances)

	// Output: Scale:
	//   Stopped?:  true
	//   Replicas:  3
}

func ExampleAppSpecAutoscaling() {
	autoscalingSpec := &kfv1alpha1.AppSpecAutoscaling{
		Enabled:     true,
		MinReplicas: ptr.Int32(1),
		MaxReplicas: ptr.Int32(3),

		Rules: []kfv1alpha1.AppAutoscalingRule{
			{
				RuleType: kfv1alpha1.CPURuleType,
				Target:   ptr.Int32(80),
			},
		},
	}

	describe.AppSpecAutoscaling(os.Stdout, autoscalingSpec)

	// Output: Autoscaling:
	//   Enabled?:     true
	//   MaxReplicas:  3
	//   MinReplicas:  1
	//   Rules:
	//     RuleType  Target
	//     CPU       80
}

func ExampleMetaV1Beta1Table() {
	describe.MetaV1Beta1Table(os.Stdout, &metav1beta1.Table{
		ColumnDefinitions: []metav1beta1.TableColumnDefinition{
			{Name: "Name"},
			{Name: "Age"},
			{Name: "Instances"},
		},

		Rows: []metav1beta1.TableRow{
			{Cells: []interface{}{"First", "12d", 12}},
			{Cells: []interface{}{"Second", "3h", 1}},
			{Cells: []interface{}{"Third", "9s", 0}},
		},
	})

	// Output: Name    Age  Instances
	// First   12d  12
	// Second  3h   1
	// Third   9s   0
}
