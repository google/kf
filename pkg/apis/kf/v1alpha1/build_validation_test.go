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
	"strings"
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/kf/testutil"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	logtesting "knative.dev/pkg/logging/testing"
)

func TestBuild_Validate(t *testing.T) {
	goodSpec := BuildSpec{
		BuildTaskRef: buildpackV3BuildTaskRef(),
	}

	badMeta := metav1.ObjectMeta{
		Name: strings.Repeat("A", 64), // Too long
	}

	cases := map[string]struct {
		spec Build
		want *apis.FieldError
	}{
		"valid spec": {
			spec: Build{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: goodSpec,
			},
		},
		"invalid ObjectMeta": {
			spec: Build{
				ObjectMeta: badMeta,
				Spec:       goodSpec,
			},
			want: apis.ValidateObjectMetadata(badMeta.GetObjectMeta()).ViaField("metadata"),
		},
		"invalid kind": {
			spec: Build{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: BuildSpec{
					BuildTaskRef: BuildTaskRef{
						Name: "buildpack",
						Kind: "a-terrible-kind",
					},
				},
			},
			want: ErrInvalidEnumValue("a-terrible-kind", "spec.kind", []string{
				string(tektonv1beta1.NamespacedTaskKind),
				BuiltinTaskKind,
				string(tektonv1beta1.ClusterTaskKind),
			}),
		},
		"missing name": {
			spec: Build{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: BuildSpec{
					BuildTaskRef: BuildTaskRef{
						Name: "",
						Kind: "ClusterTask",
					},
				},
			},
			want: apis.ErrMissingField("spec.name"),
		},
		"missing kind": {
			spec: Build{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: BuildSpec{
					BuildTaskRef: BuildTaskRef{
						Name: "a-name",
					},
				},
			},
			want: ErrInvalidEnumValue("", "spec.kind", []string{
				string(tektonv1beta1.NamespacedTaskKind),
				BuiltinTaskKind,
				string(tektonv1beta1.ClusterTaskKind),
			}),
		},
		"has both SOURCE_IMAGE and SourcePackage": {
			spec: Build{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: BuildSpec{
					SourcePackage: corev1.LocalObjectReference{
						Name: "some-package-name",
					},
					BuildTaskRef: builtinTaskRef("some-name"),
					Params: []BuildParam{
						{Name: SourceImageParamName, Value: "some-image"},
					},
				},
			},
			want: apis.ErrInvalidArrayValue("some-image", "spec.params", 0),
		},
	}

	store := config.NewDefaultConfigStore(logtesting.TestLogger(t))

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctx := store.ToContext(context.Background())
			got := tc.spec.Validate(ctx)

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}

func TestBuildWithDisableCustomBuildsFlag_Validate(t *testing.T) {
	cases := map[string]struct {
		spec                Build
		inCreate            bool
		disableCustomBuilds bool
		want                *apis.FieldError
	}{
		"custom builds disabled, Cluster Task": {
			spec:                basicBuild(string(tektonv1beta1.ClusterTaskKind)),
			inCreate:            true,
			disableCustomBuilds: true,
			want: apis.ErrGeneric(
				fmt.Sprintf("Custom Builds are disabled, kind must be %q but was %q",
					BuiltinTaskKind,
					string(tektonv1beta1.ClusterTaskKind)),
				"spec.kind"),
		},
		"custom builds disabled, Namespaced Task": {
			spec:                basicBuild(string(tektonv1beta1.NamespacedTaskKind)),
			inCreate:            true,
			disableCustomBuilds: true,
			want: apis.ErrGeneric(
				fmt.Sprintf("Custom Builds are disabled, kind must be %q but was %q",
					BuiltinTaskKind,
					string(tektonv1beta1.NamespacedTaskKind)),
				"spec.kind"),
		},
		"custom builds disabled, Builtin Task": {
			spec:                basicBuild(BuiltinTaskKind),
			inCreate:            true,
			disableCustomBuilds: true,
		},
		"custom builds enabled, ClusterTaskKind": {
			spec:                basicBuild(string(tektonv1beta1.ClusterTaskKind)),
			inCreate:            true,
			disableCustomBuilds: false,
		},
		"custom builds disabled, not in create, NamespacedTaskKind": {
			spec:                basicBuild(string(tektonv1beta1.NamespacedTaskKind)),
			inCreate:            false,
			disableCustomBuilds: true,
		},
	}

	store := config.NewDefaultConfigStore(logtesting.TestLogger(t))
	for tn, tc := range cases {
		ctx := store.ToContext(context.Background())
		cfg, err := config.FromContext(ctx).Defaults()
		testutil.AssertNil(t, "err", err)
		cfg.FeatureFlags = config.FeatureFlagToggles{}

		t.Run(tn, func(t *testing.T) {
			if tc.inCreate {
				ctx = apis.WithinCreate(ctx)
			}
			cfg, err = config.FromContext(ctx).Defaults()
			testutil.AssertNil(t, "err", err)
			cfg.FeatureFlags.SetDisableCustomBuilds(tc.disableCustomBuilds)
			got := tc.spec.Validate(ctx)
			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}

func TestBuildWithDockerfileBuildsFlag_Validate(t *testing.T) {
	cases := map[string]struct {
		spec                    BuildSpec
		inCreate                bool
		inUpdate                bool
		dockerfileBuildsEnabled bool
		want                    *apis.FieldError
	}{
		"dockerfile builds disabled, in create, Docker Build": {
			spec:                    DockerfileBuild("some/source/image", "path/to/dockerfile"),
			inCreate:                true,
			dockerfileBuildsEnabled: false,
			want: apis.ErrGeneric(fmt.Sprintf(
				"Dockerfile Builds are disabled, but BuildTaskRef name was %q",
				DockerfileBuildTaskName),
				"name"),
		},
		"dockerfile builds disabled, not in create, Docker Build": {
			spec:                    DockerfileBuild("some/source/image", "path/to/dockerfile"),
			inUpdate:                true,
			dockerfileBuildsEnabled: false,
			want: apis.ErrGeneric(fmt.Sprintf(
				"Dockerfile Builds are disabled, but BuildTaskRef name was %q",
				DockerfileBuildTaskName),
				"name"),
		},
		"dockerfile builds disabled, in create, Buildpack V3": {
			spec: BuildpackV3Build(
				"some/source/image",
				config.StackV3Definition{
					Name:       "stack-name",
					BuildImage: "build/image:latest",
					RunImage:   "run/image:latest",
				},
				[]string{"buildpack"},
			),
			inCreate:                true,
			dockerfileBuildsEnabled: false,
		},
		"docker builds disabled, in create, Buildpack V2": {
			spec: BuildpackV2Build(
				"some/source/image",
				config.StackV2Definition{
					Name:  "stack-name",
					Image: "google/base:latest",
				},
				[]string{"buildpack"},
				true,
			),
			inCreate:                true,
			dockerfileBuildsEnabled: false,
		},
		"docker builds disabled, Cluster Task": {
			spec:                    basicBuildSpec("name", string(tektonv1beta1.ClusterTaskKind)),
			inCreate:                true,
			dockerfileBuildsEnabled: false,
		},
		"docker builds enabled, Buildpack V3": {
			spec: BuildpackV3Build(
				"some/source/image",
				config.StackV3Definition{
					Name:       "stack-name",
					BuildImage: "build/image:latest",
					RunImage:   "run/image:latest",
				},
				[]string{"buildpack"},
			),
			inCreate:                true,
			dockerfileBuildsEnabled: true,
		},
		"docker builds enabled, Buildpack V2": {
			spec: BuildpackV2Build(
				"some/source/image",
				config.StackV2Definition{
					Name:  "stack-name",
					Image: "google/base:latest",
				},
				[]string{"buildpack"},
				true,
			),
			inCreate:                true,
			dockerfileBuildsEnabled: true,
		},
		"docker builds enabled, Cluster Task": {
			spec:                    basicBuildSpec("name", string(tektonv1beta1.ClusterTaskKind)),
			inCreate:                true,
			dockerfileBuildsEnabled: true,
		},
	}

	store := config.NewDefaultConfigStore(logtesting.TestLogger(t))
	for tn, tc := range cases {
		ctx := store.ToContext(context.Background())
		cfg, err := config.FromContext(ctx).Defaults()
		testutil.AssertNil(t, "err", err)
		cfg.FeatureFlags = config.FeatureFlagToggles{}

		t.Run(tn, func(t *testing.T) {
			if tc.inCreate {
				ctx = apis.WithinCreate(ctx)
			}
			cfg, err = config.FromContext(ctx).Defaults()
			testutil.AssertNil(t, "err", err)
			cfg.FeatureFlags.SetDockerfileBuilds(tc.dockerfileBuildsEnabled)
			got := tc.spec.Validate(ctx)
			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}

func TestBuildWithCustomBuildpacksAndStacksFlag_Validate(t *testing.T) {
	store := config.NewDefaultConfigStore(logtesting.TestLogger(t))
	ctx := store.ToContext(context.Background())
	cfg, err := config.FromContext(ctx).Defaults()
	testutil.AssertNil(t, "err", err)
	stackV2 := basicStackV2("test")
	stackV3 := basicStackV3("test")
	cfg.SpaceStacksV2 = config.StackV2List{stackV2}
	cfg.SpaceStacksV3 = config.StackV3List{stackV3}
	cfg.SpaceBuildpacksV2 = config.BuildpackV2List{basicBuildpackV2("test")}

	v2Build := BuildpackV2Build("some/source/image", stackV2, []string{}, false)
	v2BuildCustomBuildpack := BuildpackV2Build("some/source/image", stackV2, []string{"buildpack"}, false)
	v2BuildCustomStack := BuildpackV2Build("some/source/image", basicStackV2("stack"), []string{}, false)
	v2BuildInvalidParams := basicBuildSpec(BuildpackV2BuildTaskName, BuiltinTaskKind)
	v3Build := BuildpackV3Build("some/source/image", stackV3, []string{})
	v3BuildCustomBuildpack := BuildpackV3Build("some/source/image", stackV3, []string{"buildpack"})
	v3BuildCustomStack := BuildpackV3Build("some/source/image", basicStackV3("stack"), []string{})
	v3BuildInvalidParams := basicBuildSpec(BuildpackV3BuildTaskName, BuiltinTaskKind)
	docker := DockerfileBuild("some/source/image", "path/to/dockerfile")
	cluster := basicBuildSpec("name", string(tektonv1beta1.ClusterTaskKind))

	cases := map[string]struct {
		spec                    BuildSpec
		customBuildpacksEnabled bool
		customStacksEnabled     bool
		want                    *apis.FieldError
	}{
		"V2 Build, nothing custom, default": {
			spec:                    v2Build,
			customBuildpacksEnabled: true,
			customStacksEnabled:     true,
		},
		"V2 Build, nothing custom, custom Buildpacks disabled": {
			spec:                    v2Build,
			customBuildpacksEnabled: false,
			customStacksEnabled:     true,
		},
		"V2 Build, nothing custom, custom Stacks disabled": {
			spec:                    v2Build,
			customBuildpacksEnabled: true,
			customStacksEnabled:     false,
		},
		"V2 Build, custom buildpacks, default": {
			spec:                    v2BuildCustomBuildpack,
			customBuildpacksEnabled: true,
			customStacksEnabled:     true,
		},
		"V2 Build, custom Buildpacks, custom Buildpacks disabled": {
			spec:                    v2BuildCustomBuildpack,
			customBuildpacksEnabled: false,
			customStacksEnabled:     true,
			want: apis.ErrGeneric(
				"Builds are restricted to configured Buildpacks, but Build depends on other Buildpacks", "name"),
		},
		"V2 Build, custom Buildpacks, custom Stacks disabled": {
			spec:                    v2BuildCustomBuildpack,
			customBuildpacksEnabled: true,
			customStacksEnabled:     false,
		},
		"V2 Build, custom Stacks, default": {
			spec:                    v2BuildCustomStack,
			customBuildpacksEnabled: true,
			customStacksEnabled:     true,
		},
		"V2 Build, custom Stacks, custom Buildpacks disabled": {
			spec:                    v2BuildCustomStack,
			customBuildpacksEnabled: false,
			customStacksEnabled:     true,
		},
		"V2 Build, custom Stacks, custom Stacks disabled": {
			spec:                    v2BuildCustomStack,
			customBuildpacksEnabled: true,
			customStacksEnabled:     false,
			want: apis.ErrGeneric(
				"Builds are restricted to configured Stacks, but Build depends on other Stacks", "name"),
		},
		"V2 Build, invalid spec, default": {
			spec:                    v2BuildInvalidParams,
			customBuildpacksEnabled: true,
			customStacksEnabled:     true,
		},
		"V2 Build, invalid spec, custom Buildpacks disabled": {
			spec:                    v2BuildInvalidParams,
			customBuildpacksEnabled: false,
			customStacksEnabled:     true,
			want: apis.ErrGeneric(
				"Builds are restricted to configured Buildpacks, but Build depends on other Buildpacks", "name"),
		},
		"V2 Build, invalid spec, custom Stacks disabled": {
			spec:                    v2BuildInvalidParams,
			customBuildpacksEnabled: true,
			customStacksEnabled:     false,
			want: apis.ErrGeneric(
				"Builds are restricted to configured Stacks, but Build depends on other Stacks", "name"),
		},
		"V3 Build, nothing custom, default": {
			spec:                    v3Build,
			customBuildpacksEnabled: true,
			customStacksEnabled:     true,
		},
		"V3 Build, nothing custom, custom Buildpacks disabled": {
			spec:                    v3Build,
			customBuildpacksEnabled: false,
			customStacksEnabled:     true,
		},
		"V3 Build, nothing custom, custom Stacks disabled": {
			spec:                    v3Build,
			customBuildpacksEnabled: true,
			customStacksEnabled:     false,
		},
		"V3 Build, custom Buildpacks, default": {
			spec:                    v3BuildCustomBuildpack,
			customBuildpacksEnabled: true,
			customStacksEnabled:     true,
		},
		"V3 Build, custom Buildpacks, custom Buildpacks disabled": {
			spec:                    v3BuildCustomBuildpack,
			customBuildpacksEnabled: false,
			customStacksEnabled:     true,
			// A corresponding V3 stack is required to use custom V3 Buildpacks, so this case is not checked on purpose.
		},
		"V3 Build, custom Buildpacks, custom Stacks disabled": {
			spec:                    v3BuildCustomBuildpack,
			customBuildpacksEnabled: true,
			customStacksEnabled:     false,
		},
		"V3 Build, custom Stacks, default": {
			spec:                    v3BuildCustomStack,
			customBuildpacksEnabled: true,
			customStacksEnabled:     true,
		},
		"V3 Build, custom Stacks, custom Buildpacks disabled": {
			spec:                    v3BuildCustomStack,
			customBuildpacksEnabled: false,
			customStacksEnabled:     true,
			want: apis.ErrGeneric(
				"Builds are restricted to configured Buildpacks, but Build potentially depends on other Buildpacks through a custom stack", "name"),
		},
		"V3 Build, custom Stacks, custom Stacks disabled": {
			spec:                    v3BuildCustomStack,
			customBuildpacksEnabled: true,
			customStacksEnabled:     false,
			want: apis.ErrGeneric(
				"Builds are restricted to configured Stacks, but Build depends on other Stacks", "name"),
		},
		"V3 Build, invalid spec, default": {
			spec:                    v3BuildInvalidParams,
			customBuildpacksEnabled: true,
			customStacksEnabled:     true,
		},
		"V3 Build, invalid spec, custom Buildpacks disabled": {
			spec:                    v3BuildInvalidParams,
			customBuildpacksEnabled: false,
			customStacksEnabled:     true,
			want: apis.ErrGeneric(
				"Builds are restricted to configured Buildpacks, but Build potentially depends on other Buildpacks through a custom stack", "name"),
		},
		"V3 Build, invalid spec, custom Stacks disabled": {
			spec:                    v2BuildInvalidParams,
			customBuildpacksEnabled: true,
			customStacksEnabled:     false,
			want: apis.ErrGeneric(
				"Builds are restricted to configured Stacks, but Build depends on other Stacks", "name"),
		},
		"Dockerfile Build, default": {
			spec:                    docker,
			customBuildpacksEnabled: true,
			customStacksEnabled:     true,
		},
		"Dockerfile Build, custom Buildpacks disabled": {
			spec:                    docker,
			customBuildpacksEnabled: false,
			customStacksEnabled:     true,
		},
		"Dockerfile Build, custom Stacks disabled": {
			spec:                    docker,
			customBuildpacksEnabled: true,
			customStacksEnabled:     false,
		},
		"Cluster Task, default": {
			spec:                    cluster,
			customBuildpacksEnabled: false,
			customStacksEnabled:     true,
		},
		"Cluster Task, custom Buildpacks disabled": {
			spec:                    cluster,
			customBuildpacksEnabled: false,
			customStacksEnabled:     true,
		},
		"Cluster Task, custom Stacks disabled": {
			spec:                    cluster,
			customBuildpacksEnabled: true,
			customStacksEnabled:     false,
		},
	}

	for tn, tc := range cases {
		ctx := store.ToContext(context.Background())
		cfg, err := config.FromContext(ctx).Defaults()
		testutil.AssertNil(t, "err", err)
		cfg.FeatureFlags = config.FeatureFlagToggles{}

		t.Run(tn, func(t *testing.T) {
			cfg, err = config.FromContext(ctx).Defaults()
			testutil.AssertNil(t, "err", err)
			cfg.FeatureFlags.SetCustomBuildpacks(tc.customBuildpacksEnabled)
			cfg.FeatureFlags.SetCustomStacks(tc.customStacksEnabled)
			got := tc.spec.Validate(ctx)
			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}

func basicBuild(kind string) Build {
	return Build{
		ObjectMeta: metav1.ObjectMeta{
			Name: "valid",
		},
		Spec: basicBuildSpec("a-name", kind),
	}
}

func basicBuildSpec(name, kind string) BuildSpec {
	return BuildSpec{
		BuildTaskRef: BuildTaskRef{
			Kind: kind,
			Name: name,
		},
		Params: []BuildParam{},
	}
}

func basicBuildpackV2(image string) config.BuildpackV2Definition {
	return config.BuildpackV2Definition{
		Name:     image,
		URL:      image,
		Disabled: false,
	}
}

func basicStackV2(image string) config.StackV2Definition {
	return config.StackV2Definition{
		Name:        image,
		Description: image,
		Image:       image,
	}
}

func basicStackV3(image string) config.StackV3Definition {
	return config.StackV3Definition{
		Name:        image,
		Description: image,
		BuildImage:  image + "/build",
		RunImage:    image + "/run",
	}
}
