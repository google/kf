// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"fmt"
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
	"github.com/google/kf/third_party/knative-serving/pkg/apis/autoscaling"
)

func intPtr(val int) *int {
	return &val
}

func TestAppSpecInstances_MinAnnotationValue(t *testing.T) {
	cases := map[string]struct {
		instances AppSpecInstances
		expected  string
	}{
		"stopped": {
			instances: AppSpecInstances{Stopped: true},
			expected:  "0",
		},
		"min defined": {
			instances: AppSpecInstances{Min: intPtr(3)},
			expected:  "3",
		},
		"exactly defined": {
			instances: AppSpecInstances{Exactly: intPtr(33)},
			expected:  "33",
		},
		"empty": {
			instances: AppSpecInstances{},
			expected:  "",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := tc.instances.MinAnnotationValue()

			testutil.AssertEqual(t, "annotation", tc.expected, actual)
		})
	}
}

func TestAppSpecInstances_MaxAnnotationValue(t *testing.T) {
	cases := map[string]struct {
		instances AppSpecInstances
		expected  string
	}{
		"stopped": {
			instances: AppSpecInstances{Stopped: true},
			expected:  "0",
		},
		"max defined": {
			instances: AppSpecInstances{Max: intPtr(3)},
			expected:  "3",
		},
		"exactly defined": {
			instances: AppSpecInstances{Exactly: intPtr(33)},
			expected:  "33",
		},
		"empty": {
			instances: AppSpecInstances{},
			expected:  "",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := tc.instances.MaxAnnotationValue()

			testutil.AssertEqual(t, "annotation", tc.expected, actual)
		})
	}
}

func TestAppSpecInstances_ScalingAnnotations(t *testing.T) {
	cases := map[string]struct {
		instances AppSpecInstances
		expected  map[string]string
	}{
		"stopped": {
			instances: AppSpecInstances{Stopped: true, Exactly: intPtr(10)},
			expected: map[string]string{
				autoscaling.MinScaleAnnotationKey: "0",
				autoscaling.MaxScaleAnnotationKey: "0",
			},
		},
		"max defined": {
			instances: AppSpecInstances{Max: intPtr(30)},
			expected: map[string]string{
				autoscaling.MaxScaleAnnotationKey: "30",
			},
		},
		"min defined": {
			instances: AppSpecInstances{Min: intPtr(3)},
			expected: map[string]string{
				autoscaling.MinScaleAnnotationKey: "3",
			},
		},
		"range defined": {
			instances: AppSpecInstances{Min: intPtr(3), Max: intPtr(5)},
			expected: map[string]string{
				autoscaling.MinScaleAnnotationKey: "3",
				autoscaling.MaxScaleAnnotationKey: "5",
			},
		},
		"exactly defined": {
			instances: AppSpecInstances{Exactly: intPtr(4)},
			expected: map[string]string{
				autoscaling.MinScaleAnnotationKey: "4",
				autoscaling.MaxScaleAnnotationKey: "4",
			},
		},
		"empty": {
			instances: AppSpecInstances{},
			expected:  map[string]string{},
		},
		"exactly takes precidence": {
			// If the webhook fails somehow and exactly gets defined alongside min and
			// max, then exactly takes precedence.
			instances: AppSpecInstances{Exactly: intPtr(4), Min: intPtr(3), Max: intPtr(5)},
			expected: map[string]string{
				autoscaling.MinScaleAnnotationKey: "4",
				autoscaling.MaxScaleAnnotationKey: "4",
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := tc.instances.ScalingAnnotations()

			testutil.AssertEqual(t, "annotations", tc.expected, actual)
		})
	}
}

func ExampleApp_ComponentLabels() {
	app := App{}
	app.Name = "my-app"

	labels := app.ComponentLabels("database")

	fmt.Println("label count:", len(labels))
	fmt.Println("name:", labels[NameLabel])
	fmt.Println("managed-by:", labels[ManagedByLabel])
	fmt.Println("component:", labels[ComponentLabel])

	// Output: label count: 3
	// name: my-app
	// managed-by: kf
	// component: database
}

func TestAppSpecInstances_Status(t *testing.T) {
	tests := map[string]struct {
		instances AppSpecInstances
		want      InstanceStatus
	}{
		"stopped": {
			instances: AppSpecInstances{
				Min:     intPtr(3),
				Max:     intPtr(5),
				Stopped: true,
			},
			want: InstanceStatus{
				EffectiveMin:   "0",
				EffectiveMax:   "0",
				Representation: "stopped",
			},
		},
		"ranged": {
			instances: AppSpecInstances{
				Min: intPtr(3),
				Max: intPtr(5),
			},
			want: InstanceStatus{
				EffectiveMin:   "3",
				EffectiveMax:   "5",
				Representation: "3-5",
			},
		},
		"no min": {
			instances: AppSpecInstances{
				Max: intPtr(5),
			},
			want: InstanceStatus{
				EffectiveMin:   "",
				EffectiveMax:   "5",
				Representation: "0-5",
			},
		},
		"no max": {
			instances: AppSpecInstances{
				Min: intPtr(5),
			},
			want: InstanceStatus{
				EffectiveMin:   "5",
				EffectiveMax:   "",
				Representation: "5-∞",
			},
		},
		"exactly": {
			instances: AppSpecInstances{
				Exactly: intPtr(3),
			},
			want: InstanceStatus{
				EffectiveMin:   "3",
				EffectiveMax:   "3",
				Representation: "3",
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			testutil.AssertEqual(t, "status", tc.want, tc.instances.Status())
		})
	}
}
