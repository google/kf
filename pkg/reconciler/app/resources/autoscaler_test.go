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
	"errors"
	"fmt"
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/pkg/ptr"
)

func ExampleAutoscalerName() {
	app := &v1alpha1.App{}
	app.Name = "my-app"

	fmt.Println("Autoscaler name:", AutoscalerName(app))

	// Output: Autoscaler name: my-app
}

func TestMakeAutoscaler(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		app     *v1alpha1.App
		space   *v1alpha1.Space
		want    *autoscalingv1.HorizontalPodAutoscaler
		wantErr error
	}{
		"disabled": {
			app: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-app",
				},
				Spec: v1alpha1.AppSpec{
					Instances: v1alpha1.AppSpecInstances{
						Autoscaling: v1alpha1.AppSpecAutoscaling{
							Enabled: false,
						},
						Replicas: ptr.Int32(30),
					},
				},
			},
			want: nil,
		},
		"app stoped": {
			app: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-app",
				},
				Spec: v1alpha1.AppSpec{
					Instances: v1alpha1.AppSpecInstances{
						Stopped: true,
						Autoscaling: v1alpha1.AppSpecAutoscaling{
							Enabled:     true,
							MaxReplicas: ptr.Int32(2),
						},
						Replicas: ptr.Int32(30),
					},
				},
			},
			want: nil,
		},
		"MaxReplicas not defined": {
			app: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-app",
				},
				Spec: v1alpha1.AppSpec{
					Instances: v1alpha1.AppSpecInstances{
						Autoscaling: v1alpha1.AppSpecAutoscaling{
							Enabled: true,
						},
						Replicas: ptr.Int32(30),
					},
				},
			},
			want: nil,
		},
		"Rules not defined": {
			app: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-app",
				},
				Spec: v1alpha1.AppSpec{
					Instances: v1alpha1.AppSpecInstances{
						Autoscaling: v1alpha1.AppSpecAutoscaling{
							Enabled:     true,
							MaxReplicas: ptr.Int32(3),
						},
						Replicas: ptr.Int32(30),
					},
				},
			},
			want: nil,
		},
		"Too many rules": {
			app: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-app",
				},
				Spec: v1alpha1.AppSpec{
					Instances: v1alpha1.AppSpecInstances{
						Autoscaling: v1alpha1.AppSpecAutoscaling{
							Enabled:     true,
							MaxReplicas: ptr.Int32(3),
							Rules: []v1alpha1.AppAutoscalingRule{
								{
									RuleType: v1alpha1.CPURuleType,
									Target:   ptr.Int32(50),
								},
								{
									RuleType: v1alpha1.CPURuleType,
									Target:   ptr.Int32(90),
								},
							},
						},
						Replicas: ptr.Int32(30),
					},
				},
			},
			wantErr: errors.New("too many autoscaling rules"),
		},
		"happy": {
			app: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-app",
				},
				Spec: v1alpha1.AppSpec{
					Instances: v1alpha1.AppSpecInstances{
						Autoscaling: v1alpha1.AppSpecAutoscaling{
							Enabled:     true,
							MinReplicas: ptr.Int32(1),
							MaxReplicas: ptr.Int32(1),
							Rules: []v1alpha1.AppAutoscalingRule{
								{
									RuleType: v1alpha1.CPURuleType,
									Target:   ptr.Int32(50),
								},
							},
						},
						Replicas: ptr.Int32(30),
					},
				},
			},
			space: &v1alpha1.Space{},
			want: &autoscalingv1.HorizontalPodAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-app",
					Labels: map[string]string{
						"app.kubernetes.io/component":  "autoscaler",
						"app.kubernetes.io/managed-by": "kf",
						"app.kubernetes.io/name":       "my-app",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         "kf.dev/v1alpha1",
							Kind:               "App",
							Name:               "my-app",
							Controller:         ptr.Bool(true),
							BlockOwnerDeletion: ptr.Bool(true),
						},
					},
				},

				Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
						APIVersion: "kf.dev/v1alpha1",
						Kind:       "App",
						Name:       "my-app",
					},
					MinReplicas:                    ptr.Int32(1),
					MaxReplicas:                    1,
					TargetCPUUtilizationPercentage: ptr.Int32(50),
				},
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// automatically fill in desired spec
			got, err := MakeHorizontalPodAutoScaler(tc.app)
			testutil.AssertEqual(t, "Autoscaler", tc.want, got)
			testutil.AssertEqual(t, "Error", tc.wantErr, err)
		})
	}
}
