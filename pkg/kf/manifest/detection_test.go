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

package manifest

import (
	"errors"
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
)

func defaultV2Stack() config.StackV2Definition {
	return config.StackV2Definition{
		Name:  "default_v2_stack",
		Image: "default/v2/stack:latest",
	}
}

func altV2Stack() config.StackV2Definition {
	return config.StackV2Definition{
		Name:  "alt_v2_stack",
		Image: "alt/v2/stack:latest",
	}
}

func altV3Stack() config.StackV3Definition {
	return config.StackV3Definition{
		Name:       "alt_v3_stack",
		BuildImage: "alt/v3/stack:build",
		RunImage:   "alt/v3/stack:run",
	}
}

func defaultV3Stack() config.StackV3Definition {
	return config.StackV3Definition{
		Name:       "default_v3_stack",
		BuildImage: "default/v3/stack:build",
		RunImage:   "default/v3/stack:run",
	}
}

func builtinV2Buildpack() config.BuildpackV2Definition {
	return config.BuildpackV2Definition{
		Name: "java_buildpack",
		URL:  "quic://java/buildpack",
	}
}

func buildConfig() v1alpha1.SpaceStatusBuildConfig {
	return v1alpha1.SpaceStatusBuildConfig{
		BuildpacksV2: config.BuildpackV2List{
			builtinV2Buildpack(),
		},
		StacksV2: config.StackV2List{
			defaultV2Stack(),
			altV2Stack(),
		},
		StacksV3: config.StackV3List{
			defaultV3Stack(),
			altV3Stack(),
		},
		DefaultToV3Stack: false,
	}
}

func bldPtr(build v1alpha1.BuildSpec) *v1alpha1.BuildSpec {
	return &build
}

func TestDetectBuildType(t *testing.T) {
	t.Parallel()

	const skipDetection = true

	cases := map[string]struct {
		app                      Application
		sourceImage              string
		buildConfig              v1alpha1.SpaceStatusBuildConfig
		expectedBuildSpec        *v1alpha1.BuildSpec
		expectedShouldPushSource bool
		expectedErr              error
	}{
		"default no config": {
			app: Application{
				Name: "app",
			},
			buildConfig: v1alpha1.SpaceStatusBuildConfig{},
			sourceImage: "source-image",
			expectedErr: errors.New("can't detect the build type from the manifest"),
		},
		"default v2": {
			app: Application{
				Name: "app",
			},
			buildConfig: buildConfig(),
			sourceImage: "source-image",
			expectedBuildSpec: bldPtr(v1alpha1.BuildpackV2Build(
				"source-image",
				defaultV2Stack(),
				[]string{"quic://java/buildpack"},
				!skipDetection,
			)),
			expectedShouldPushSource: true,
		},
		"default v3": {
			app: Application{
				Name: "app",
			},
			buildConfig: (func() v1alpha1.SpaceStatusBuildConfig {
				cfg := buildConfig()
				cfg.DefaultToV3Stack = true
				return cfg
			}()),
			sourceImage:              "source-image",
			expectedBuildSpec:        bldPtr(v1alpha1.BuildpackV3Build("source-image", defaultV3Stack(), nil)),
			expectedShouldPushSource: true,
		},
		"default v2 only have v3": {
			app: Application{
				Name: "app",
			},
			buildConfig: (func() v1alpha1.SpaceStatusBuildConfig {
				cfg := buildConfig()
				cfg.DefaultToV3Stack = false
				cfg.StacksV2 = nil
				return cfg
			}()),
			sourceImage:              "source-image",
			expectedBuildSpec:        bldPtr(v1alpha1.BuildpackV3Build("source-image", defaultV3Stack(), nil)),
			expectedShouldPushSource: true,
		},
		"v2 buildpack": {
			app: Application{
				Name:  "app",
				Stack: altV2Stack().Name,
				Buildpacks: []string{
					"https://github.com/cloudfoundry/java-buildpack",
				},
			},
			buildConfig: buildConfig(),
			sourceImage: "source-image",
			expectedBuildSpec: bldPtr(v1alpha1.BuildpackV2Build(
				"source-image",
				altV2Stack(),
				[]string{"https://github.com/cloudfoundry/java-buildpack"},
				skipDetection,
			)),
			expectedShouldPushSource: true,
		},
		"v3 buildpack": {
			app: Application{
				Name:  "app",
				Stack: altV3Stack().Name,
				Buildpacks: []string{
					"bp.1",
					"bp.2",
				},
			},
			buildConfig:              buildConfig(),
			sourceImage:              "source-image",
			expectedBuildSpec:        bldPtr(v1alpha1.BuildpackV3Build("source-image", altV3Stack(), []string{"bp.1", "bp.2"})),
			expectedShouldPushSource: true,
		},

		"cf custom buildpack": {
			app: Application{
				Name: "app",
				Buildpacks: []string{
					"https://git.mycompany.com/java_buildpack.git#v3.11.2",
				},
			},
			buildConfig: buildConfig(),
			sourceImage: "source-image",
			expectedBuildSpec: bldPtr(v1alpha1.BuildpackV2Build(
				"source-image",
				defaultV2Stack(),
				[]string{"https://git.mycompany.com/java_buildpack.git#v3.11.2"},
				skipDetection,
			)),
			expectedShouldPushSource: true,
		},
		"cf buildpack expansion": {
			app: Application{
				Name:  "app",
				Stack: defaultV2Stack().Name,
				Buildpacks: []string{
					"java_buildpack",
				},
			},
			buildConfig: buildConfig(),
			sourceImage: "source-image",
			expectedBuildSpec: bldPtr(v1alpha1.BuildpackV2Build(
				"source-image",
				defaultV2Stack(),
				[]string{builtinV2Buildpack().URL},
				skipDetection,
			)),
			expectedShouldPushSource: true,
		},
		"manifest build": {
			app: Application{
				Name: "app",
				KfApplicationExtension: KfApplicationExtension{
					Build: &v1alpha1.BuildSpec{
						BuildTaskRef: v1alpha1.BuildTaskRef{
							Name:       "custom-task",
							Kind:       "ClusterTask",
							APIVersion: "tekton.dev/v1beta1",
						},
					},
				},
			},
			buildConfig: buildConfig(),
			sourceImage: "source-image",
			expectedBuildSpec: &v1alpha1.BuildSpec{
				BuildTaskRef: v1alpha1.BuildTaskRef{
					Name:       "custom-task",
					Kind:       "ClusterTask",
					APIVersion: "tekton.dev/v1beta1",
				},
				Params: []v1alpha1.BuildParam{
					v1alpha1.StringParam(v1alpha1.SourceImageParamName, "source-image"),
				},
			},
			expectedShouldPushSource: true,
		},
		"docker image": {
			app: Application{
				Name: "app",
				Docker: AppDockerImage{
					Image: "redis",
				},
			},
			buildConfig:              buildConfig(),
			sourceImage:              "source-image",
			expectedBuildSpec:        nil,
			expectedShouldPushSource: false,
		},
		"multiple buildpacks error": {
			app: Application{
				Name:  "app",
				Stack: "alt_v2_stack",
				Docker: AppDockerImage{
					Image: "redis",
				},
			},
			buildConfig:              buildConfig(),
			sourceImage:              "source-image",
			expectedBuildSpec:        nil,
			expectedShouldPushSource: false,
			expectedErr:              errMultipleBuilds("Buildpack V2", "Docker Image"),
		},
		"dockerfile build": {
			app: Application{
				Name: "app",
				KfApplicationExtension: KfApplicationExtension{
					Dockerfile: Dockerfile{
						Path: "path/to/Dockerfile",
					},
				},
			},
			buildConfig:              buildConfig(),
			sourceImage:              "source-image",
			expectedBuildSpec:        bldPtr(v1alpha1.DockerfileBuild("source-image", "path/to/Dockerfile")),
			expectedShouldPushSource: true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			builder, actualShouldPushSource, actualErr := tc.app.DetectBuildType(tc.buildConfig)
			testutil.AssertEqual(t, "error", tc.expectedErr, actualErr)

			if actualErr != nil {
				// An error implies the rest of the test is not necessary
				return
			}

			actualBuildSpec, err := builder(tc.sourceImage)
			testutil.AssertNil(t, "error", err)

			testutil.AssertEqual(t, "buildSpec", tc.expectedBuildSpec, actualBuildSpec)
			testutil.AssertEqual(t, "shouldPushSource", tc.expectedShouldPushSource, actualShouldPushSource)
		})
	}
}
