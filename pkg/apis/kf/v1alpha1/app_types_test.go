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
	"errors"
	"fmt"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	"knative.dev/pkg/ptr"
)

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

func ExampleAppComponentLabels() {
	labels := AppComponentLabels("my-app", "database")

	fmt.Println("label count:", len(labels))
	fmt.Println("name:", labels[NameLabel])
	fmt.Println("managed-by:", labels[ManagedByLabel])
	fmt.Println("component:", labels[ComponentLabel])

	// Output: label count: 3
	// name: my-app
	// managed-by: kf
	// component: database
}

func TestBuildSpec_NeedsUpdateRequestsIncrement(t *testing.T) {

	buildpackBuildSpec := &BuildSpec{
		BuildTaskRef: buildpackV3BuildTaskRef(),
	}
	dockerfileBuildSpec := &BuildSpec{
		BuildTaskRef: dockerfileBuildTaskRef(),
	}
	current := AppSpec{
		Build: AppSpecBuild{
			UpdateRequests: 33,
			Spec:           buildpackBuildSpec,
		},
	}

	cases := map[string]struct {
		old                  AppSpec
		expectNeedsIncrement bool
	}{
		"matches exactly": {
			old: AppSpec{
				Build: AppSpecBuild{
					UpdateRequests: 33,
					Spec:           buildpackBuildSpec,
				},
			},
			expectNeedsIncrement: false,
		},
		"increment already done": {
			old: AppSpec{
				Build: AppSpecBuild{
					UpdateRequests: 34,
					Spec:           dockerfileBuildSpec,
				},
			},
			expectNeedsIncrement: false,
		},
		"needs increment": {
			old: AppSpec{
				Build: AppSpecBuild{
					UpdateRequests: 33,
					Spec:           dockerfileBuildSpec,
				},
			},
			expectNeedsIncrement: true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := current.Build.NeedsUpdateRequestsIncrement(tc.old.Build)

			testutil.AssertEqual(t, "needsIncrement", tc.expectNeedsIncrement, actual)
		})
	}
}
func TestAppSpecInstances_Status(t *testing.T) {
	tests := map[string]struct {
		instances AppSpecInstances
		want      InstanceStatus
	}{
		"stopped": {
			instances: AppSpecInstances{
				Stopped: true,
			},
			want: InstanceStatus{
				Replicas:       0,
				Representation: "stopped",
			},
		},
		"stopped, with autoscaling enabled": {
			instances: AppSpecInstances{
				Stopped: true,
				Autoscaling: AppSpecAutoscaling{
					Enabled:     true,
					MinReplicas: ptr.Int32(1),
					MaxReplicas: ptr.Int32(3),
					Rules: []AppAutoscalingRule{
						{
							RuleType: CPURuleType,
							Target:   ptr.Int32(30),
						},
					},
				},
			},
			want: InstanceStatus{
				Replicas:       0,
				Representation: "stopped",
			},
		},
		"exactly": {
			instances: AppSpecInstances{
				Replicas: ptr.Int32(3),
			},
			want: InstanceStatus{
				Replicas:       3,
				Representation: "3",
			},
		},
		"missing": {
			instances: AppSpecInstances{},
			want: InstanceStatus{
				Replicas:       1,
				Representation: "1",
			},
		},
		"autoscaled": {
			instances: AppSpecInstances{
				Autoscaling: AppSpecAutoscaling{
					Enabled:     true,
					MinReplicas: ptr.Int32(1),
					MaxReplicas: ptr.Int32(3),
					Rules: []AppAutoscalingRule{
						{
							RuleType: CPURuleType,
							Target:   ptr.Int32(30),
						},
					},
				},
			},
			want: InstanceStatus{
				Replicas:       1,
				Representation: "1 (autoscaled 1 to 3)",
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			testutil.AssertEqual(t, "status", tc.want, tc.instances.Status())
		})
	}
}

func TestAppSpecInstances_DeploymentReplicas(t *testing.T) {
	tests := map[string]struct {
		instances     AppSpecInstances
		wantInstances int32
		wantErr       error
	}{
		"stopped": {
			instances: AppSpecInstances{
				Stopped: true,
			},
			wantInstances: 0,
			wantErr:       nil,
		},
		"exactly": {
			instances: AppSpecInstances{
				Replicas: ptr.Int32(3),
			},
			wantInstances: 3,
			wantErr:       nil,
		},
		"empty": {
			instances:     AppSpecInstances{},
			wantInstances: 0,
			wantErr:       errors.New("Exact scale required for deployment based setup"),
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			instances, err := tc.instances.DeploymentReplicas()
			testutil.AssertEqual(t, "count", tc.wantInstances, instances)
			testutil.AssertEqual(t, "error", tc.wantErr, err)
		})
	}
}
