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
	"context"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/kf/testutil"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	autoscaling "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	logtesting "knative.dev/pkg/logging/testing"
	"knative.dev/pkg/ptr"
)

func TestApp_Validate(t *testing.T) {
	goodInstances := AppSpecInstances{Stopped: true}
	goodTemplate := AppSpecTemplate{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{}},
		},
	}
	goodBuild := AppSpecBuild{
		Spec: &BuildSpec{
			BuildTaskRef: buildpackV3BuildTaskRef(),
		},
	}
	badMeta := metav1.ObjectMeta{
		Name: strings.Repeat("A", 64), // Too long
	}

	goodRoute := RouteWeightBinding{
		Weight: ptr.Int32(1),
		RouteSpecFields: RouteSpecFields{
			Hostname: "some-host",
			Domain:   "example.com",
		},
	}

	badRouteWeight := *goodRoute.DeepCopy()
	badRouteWeight.Weight = ptr.Int32(-1)

	routeWithoutDomain := *goodRoute.DeepCopy()
	routeWithoutDomain.Domain = ""

	goodRouteWithDestinationPort := *goodRoute.DeepCopy()
	goodRouteWithDestinationPort.DestinationPort = ptr.Int32(8080)

	cases := map[string]struct {
		spec App
		want *apis.FieldError
	}{
		"valid": {
			spec: App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: AppSpec{
					Template:  goodTemplate,
					Instances: goodInstances,
					Build:     goodBuild,
					Routes: []RouteWeightBinding{
						goodRoute,
						routeWithoutDomain,
						goodRouteWithDestinationPort,
					},
				},
			},
		},
		"valid with buildRef": {
			spec: App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: AppSpec{
					Template:  goodTemplate,
					Instances: goodInstances,
					Build: AppSpecBuild{
						BuildRef: &corev1.LocalObjectReference{Name: "some-name"},
					},
					Routes: []RouteWeightBinding{
						goodRoute,
						routeWithoutDomain,
						goodRouteWithDestinationPort,
					},
				},
			},
		},
		"invalid ObjectMeta": {
			spec: App{
				ObjectMeta: badMeta,
				Spec: AppSpec{
					Template:  goodTemplate,
					Instances: goodInstances,
					Build:     goodBuild,
				},
			},
			want: apis.ValidateObjectMetadata(badMeta.GetObjectMeta()).ViaField("metadata"),
		},
		"invalid instances": {
			spec: App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: AppSpec{
					Template:  goodTemplate,
					Instances: AppSpecInstances{Replicas: ptr.Int32(-1)},
					Build:     goodBuild,
				},
			},
			want: apis.ErrInvalidValue(-1, "spec.instances.replicas"),
		},
		"invalid template": {
			spec: App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: AppSpec{
					Template:  AppSpecTemplate{},
					Instances: goodInstances,
					Build:     goodBuild,
				},
			},
			want: apis.ErrMissingField("spec.template.spec.containers"),
		},
		"invalid build fields": {
			spec: App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: AppSpec{
					Template:  goodTemplate,
					Instances: goodInstances,
					Build: AppSpecBuild{
						Spec: &BuildSpec{
							BuildTaskRef: BuildTaskRef{
								Name: "a-name",
								Kind: "a-terrible-kind",
							},
						},
					},
				},
			},
			want: ErrInvalidEnumValue("a-terrible-kind", "spec.build.spec.kind", []string{"Task", BuiltinTaskKind}),
		},
		"build image and buildRef": {
			spec: App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: AppSpec{
					Template:  goodTemplate,
					Instances: goodInstances,
					Build: AppSpecBuild{
						Image: ptr.String("some-image"),
						BuildRef: &corev1.LocalObjectReference{
							Name: "some-name",
						},
					},
				},
			},
			want: apis.ErrMultipleOneOf("spec.build.spec", "spec.build.image", "spec.build.buildRef"),
		},
		"build fields and buildRef": {
			spec: App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: AppSpec{
					Template:  goodTemplate,
					Instances: goodInstances,
					Build: AppSpecBuild{
						Spec: &BuildSpec{
							BuildTaskRef: buildpackV3BuildTaskRef(),
						},
						BuildRef: &corev1.LocalObjectReference{
							Name: "some-name",
						},
					},
				},
			},
			want: apis.ErrMultipleOneOf("spec.build.spec", "spec.build.image", "spec.build.buildRef"),
		},
		"invalid route weight": {
			spec: App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: AppSpec{
					Template:  goodTemplate,
					Instances: goodInstances,
					Build:     goodBuild,
					Routes: []RouteWeightBinding{
						badRouteWeight,
					},
				},
			},
			want: apis.ErrInvalidValue(-1, "spec.routes[0].weight"),
		},
		"route without domain": {
			spec: App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: AppSpec{
					Template:  goodTemplate,
					Instances: goodInstances,
					Build:     goodBuild,
					Routes: []RouteWeightBinding{
						routeWithoutDomain,
					},
				},
			},
		},
		"non-nil buildRef with empty name": {
			spec: App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: AppSpec{
					Template:  goodTemplate,
					Instances: goodInstances,
					Build: AppSpecBuild{
						BuildRef: &corev1.LocalObjectReference{
							Name: "",
						},
					},
				},
			},
			want: apis.ErrMissingField("spec.build.buildRef.name"),
		},
	}

	store := config.NewDefaultConfigStore(logtesting.TestLogger(t))

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := tc.spec.Validate(store.ToContext(context.Background()))

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}

func TestAppSpec_ValidateBuildSpec(t *testing.T) {
	dockerBuild := DockerfileBuild("some/source/image", "path/to/dockerfile")
	dockerBuild2 := DockerfileBuild("other/source/image", "path/to/dockerfile")
	cases := map[string]struct {
		old                      *AppSpec
		current                  AppSpec
		customBuildsDisabled     bool
		dockerfileBuildsDisabled bool
		want                     *apis.FieldError
	}{
		"invalid build fields": {
			current: AppSpec{
				Build: AppSpecBuild{
					Spec: &BuildSpec{
						BuildTaskRef: BuildTaskRef{
							Name: "a-name",
							Kind: "a-terrible-kind",
						},
					},
				},
			},
			want: ErrInvalidEnumValue("a-terrible-kind", "spec.kind", []string{"Task", BuiltinTaskKind}),
		},
		"build changed incorrectly": {
			old: &AppSpec{
				Build: AppSpecBuild{
					Spec: &BuildSpec{
						BuildTaskRef: buildpackV3BuildTaskRef(),
					},
				},
			},
			current: AppSpec{
				Build: AppSpecBuild{
					Spec: &BuildSpec{
						BuildTaskRef: dockerfileBuildTaskRef(),
					},
				},
			},
			want: &apis.FieldError{Message: "must increment UpdateRequests with change to build", Paths: []string{"UpdateRequests"}},
		},
		"build UpdateRequests less than last": {
			old: &AppSpec{
				Build: AppSpecBuild{
					UpdateRequests: 42,
					Spec: &BuildSpec{
						BuildTaskRef: buildpackV3BuildTaskRef(),
					},
				},
			},
			current: AppSpec{
				Build: AppSpecBuild{
					UpdateRequests: 5,
					Spec: &BuildSpec{
						BuildTaskRef: dockerfileBuildTaskRef(),
					},
				},
			},
			want: &apis.FieldError{Message: "UpdateRequests must be nondecreasing, previous value: 42 new value: 5", Paths: []string{"UpdateRequests"}},
		},
		"build changed with increment": {
			old: &AppSpec{
				Build: AppSpecBuild{
					Spec: &BuildSpec{
						BuildTaskRef: buildpackV3BuildTaskRef(),
					},
				},
			},
			current: AppSpec{
				Build: AppSpecBuild{
					UpdateRequests: 2,
					Spec: &BuildSpec{
						BuildTaskRef: dockerfileBuildTaskRef(),
					},
				},
			},
		},
		"Custom Builds Disabled, in create": {
			current: AppSpec{
				Build: AppSpecBuild{
					Spec: &BuildSpec{
						BuildTaskRef: BuildTaskRef{
							Name: "a-name",
							Kind: string(tektonv1beta1.NamespacedTaskKind),
						},
					},
				},
			},
			customBuildsDisabled: true,
			want: apis.ErrGeneric(
				fmt.Sprintf("Custom Builds are disabled, kind must be %q but was %q",
					BuiltinTaskKind,
					string(tektonv1beta1.NamespacedTaskKind)),
				"spec.kind"),
		},
		"Docker Builds Disabled, in update": {
			old: &AppSpec{
				Build: AppSpecBuild{
					UpdateRequests: 4,
					Spec:           &dockerBuild,
				},
			},
			current: AppSpec{
				Build: AppSpecBuild{
					UpdateRequests: 5,
					Spec:           &dockerBuild2,
				},
			},
			dockerfileBuildsDisabled: true,
			want: apis.ErrGeneric(fmt.Sprintf(
				"Dockerfile Builds are disabled, but BuildTaskRef name was %q",
				DockerfileBuildTaskName),
				"spec.name"),
		},
		"Dockerfile Builds Disabled, in create": {
			current: AppSpec{
				Build: AppSpecBuild{
					Spec: &dockerBuild,
				},
			},
			dockerfileBuildsDisabled: true,
			want: apis.ErrGeneric(fmt.Sprintf(
				"Dockerfile Builds are disabled, but BuildTaskRef name was %q",
				DockerfileBuildTaskName),
				"spec.name"),
		},
	}

	store := config.NewDefaultConfigStore(logtesting.TestLogger(t))

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctx := store.ToContext(context.Background())

			if tc.old != nil {
				ctx = apis.WithinUpdate(ctx, &App{
					Spec: *tc.old,
				})
			} else {
				ctx = apis.WithinCreate(ctx)
			}

			cfg, err := config.FromContext(ctx).Defaults()
			testutil.AssertNil(t, "err", err)
			cfg.FeatureFlags = config.FeatureFlagToggles{}
			cfg.FeatureFlags.SetDisableCustomBuilds(tc.customBuildsDisabled)
			cfg.FeatureFlags.SetDockerfileBuilds(!tc.dockerfileBuildsDisabled)

			got := tc.current.Build.Validate(store.ToContext(ctx))

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}

func TestAppSpecInstances_Validate(t *testing.T) {
	// These test cases are broken out separately because they're
	// too extenstive to copy the whole service struct for.

	cases := map[string]struct {
		spec AppSpecInstances
		want *apis.FieldError
	}{
		"blank": {
			spec: AppSpecInstances{},
		},
		"stopped": {
			spec: AppSpecInstances{Stopped: true},
		},
		"valid replicas": {
			spec: AppSpecInstances{Replicas: ptr.Int32(3)},
		},
		"replicas lt 0": {
			spec: AppSpecInstances{Replicas: ptr.Int32(-1)},
			want: apis.ErrInvalidValue(-1, "replicas"),
		},
		"replicas eq 0": {
			spec: AppSpecInstances{Replicas: ptr.Int32(0)},
			want: apis.ErrInvalidValue(0, "replicas"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := tc.spec.Validate(context.Background())

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}

func TestAutoscalingSpec_Validate(t *testing.T) {
	// These test cases are broken out separately because they're
	// too extenstive to copy the whole service struct for.

	cases := map[string]struct {
		spec AppSpecAutoscaling
		want *apis.FieldError
	}{
		"blank, valid": {
			spec: AppSpecAutoscaling{},
		},
		"multiple rules": {
			spec: AppSpecAutoscaling{
				MaxReplicas: ptr.Int32(3),
				MinReplicas: ptr.Int32(1),
				Rules: []AppAutoscalingRule{
					{
						RuleType: CPURuleType,
						Target:   ptr.Int32(80),
					},
					{
						RuleType: CPURuleType,
						Target:   ptr.Int32(50),
					},
				},
			},
			want: apis.ErrMultipleOneOf("rules"),
		},
		"max replicas is nil, min replicas is not nil": {
			spec: AppSpecAutoscaling{
				MinReplicas: ptr.Int32(3),
			},
			want: apis.ErrMissingField("maxReplicas"),
		},
		"max replicas not a positive integer": {
			spec: AppSpecAutoscaling{MaxReplicas: ptr.Int32(0)},
			want: apis.ErrOutOfBoundsValue(0, 1, math.MaxInt32, "maxReplicas"),
		},
		"min replicas not a positive integer": {
			spec: AppSpecAutoscaling{
				MaxReplicas: ptr.Int32(3),
				MinReplicas: ptr.Int32(0),
			},
			want: apis.ErrOutOfBoundsValue(0, 1, 3, "minReplicas"),
		},
		"max replicas lt min replicas": {
			spec: AppSpecAutoscaling{
				MaxReplicas: ptr.Int32(3),
				MinReplicas: ptr.Int32(4),
			},
			want: apis.ErrOutOfBoundsValue(4, 1, 3, "minReplicas"),
		},
		"min replicas nil, max replicas valid, invalid": {
			spec: AppSpecAutoscaling{
				MaxReplicas: ptr.Int32(3),
			},
			want: apis.ErrMissingField("minReplicas"),
		},
		"min replicas valid, max replicas valid, valid": {
			spec: AppSpecAutoscaling{
				MaxReplicas: ptr.Int32(3),
				MinReplicas: ptr.Int32(1),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := tc.spec.Validate(context.Background())
			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}

func TestAppAutoscalingRules_Validate(t *testing.T) {
	// These test cases are broken out separately because they're
	// too extenstive to copy the whole service struct for.

	cases := map[string]struct {
		spec AppAutoscalingRule
		want *apis.FieldError
	}{
		"blank": {
			spec: AppAutoscalingRule{},
			want: apis.ErrInvalidValue("", "ruleType"),
		},
		"ruleType not CPU": {
			spec: AppAutoscalingRule{
				RuleType: "type",
			},
			want: apis.ErrInvalidValue("type", "ruleType"),
		},
		"target is nil": {
			spec: AppAutoscalingRule{
				RuleType: CPURuleType,
			},
			want: apis.ErrMissingField("target"),
		},
		"invalid target, too small": {
			spec: AppAutoscalingRule{
				RuleType: CPURuleType,
				Target:   ptr.Int32(0),
			},
			want: apis.ErrOutOfBoundsValue(0, 1, 100, "target"),
		},
		"invalid target, too large": {
			spec: AppAutoscalingRule{
				RuleType: CPURuleType,
				Target:   ptr.Int32(101),
			},
			want: apis.ErrOutOfBoundsValue(101, 1, 100, "target"),
		},
		"valid": {
			spec: AppAutoscalingRule{
				RuleType: CPURuleType,
				Target:   ptr.Int32(80),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := tc.spec.Validate(context.Background())
			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}

func TestScale_Validate(t *testing.T) {
	// These test cases are broken out separately because they're
	// too extenstive to copy the whole service struct for.

	cases := map[string]struct {
		spec autoscaling.ScaleSpec
		want *apis.FieldError
	}{
		"valid replicas": {
			spec: autoscaling.ScaleSpec{Replicas: 3},
		},
		"replicas lt 0": {
			spec: autoscaling.ScaleSpec{Replicas: -1},
			want: apis.ErrInvalidValue(-1, "replicas"),
		},
		"replicas eq 0": {
			spec: autoscaling.ScaleSpec{Replicas: 0},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			s := Scale{
				Scale: autoscaling.Scale{Spec: tc.spec},
			}
			got := s.Validate(context.Background())

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}
