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

package spaces

import (
	"bytes"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/internal/envutil"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/spaces"
	"github.com/google/kf/pkg/kf/spaces/fake"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestNewConfigSpaceCommand(t *testing.T) {
	space := "my-space"

	cases := map[string]struct {
		args     []string
		space    v1alpha1.Space
		wantErr  error
		validate func(*testing.T, *v1alpha1.Space)
	}{
		"invalid number of args": {
			// Should have 2 more args
			args:    []string{"set-container-registry"},
			wantErr: errors.New("accepts 2 arg(s), received 0"),
		},

		"set-container-registry valid": {
			args: []string{"set-container-registry", space, "gcr.io/foo"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "container registry", "gcr.io/foo", space.Spec.BuildpackBuild.ContainerRegistry)
			},
		},

		"set-env valid": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					Execution: v1alpha1.SpaceSpecExecution{
						Env: envutil.MapToEnvVars(map[string]string{
							"EXISTS": "FOO",
							"BAR":    "BAZZ",
						}),
					},
				},
			},
			args: []string{"set-env", space, "EXISTS", "REPLACED"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "execution env", map[string]string{
					"EXISTS": "REPLACED",
					"BAR":    "BAZZ",
				}, envutil.EnvVarsToMap(space.Spec.Execution.Env))
			},
		},

		"unset-env valid": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					Execution: v1alpha1.SpaceSpecExecution{
						Env: envutil.MapToEnvVars(map[string]string{
							"EXISTS": "FOO",
							"BAR":    "BAZZ",
						}),
					},
				},
			},
			args: []string{"unset-env", space, "EXISTS"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "execution env", map[string]string{
					"BAR": "BAZZ",
				}, envutil.EnvVarsToMap(space.Spec.Execution.Env))
			},
		},

		"set-buildpack-env valid": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					BuildpackBuild: v1alpha1.SpaceSpecBuildpackBuild{
						Env: envutil.MapToEnvVars(map[string]string{
							"EXISTS": "FOO",
							"BAR":    "BAZZ",
						}),
					},
				},
			},
			args: []string{"set-buildpack-env", space, "EXISTS", "REPLACED"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "buildpack env", map[string]string{
					"EXISTS": "REPLACED",
					"BAR":    "BAZZ",
				}, envutil.EnvVarsToMap(space.Spec.BuildpackBuild.Env))
			},
		},

		"unset-buildpack-env valid": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					BuildpackBuild: v1alpha1.SpaceSpecBuildpackBuild{
						Env: envutil.MapToEnvVars(map[string]string{
							"EXISTS": "FOO",
							"BAR":    "BAZZ",
						}),
					},
				},
			},
			args: []string{"unset-buildpack-env", space, "EXISTS"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "buildpack env", map[string]string{
					"BAR": "BAZZ",
				}, envutil.EnvVarsToMap(space.Spec.BuildpackBuild.Env))
			},
		},

		"set-buildpack-builder valid": {
			args: []string{"set-buildpack-builder", space, "gcr.io/path/to/builder"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "container registry", "gcr.io/path/to/builder", space.Spec.BuildpackBuild.BuilderImage)
			},
		},

		"append-domain valid": {
			args: []string{"append-domain", space, "example.com"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "len(domains)", 1, len(space.Spec.Execution.Domains))
				testutil.AssertEqual(t, "domains", "example.com", space.Spec.Execution.Domains[0].Domain)
			},
		},

		"set-default-domain valid": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					Execution: v1alpha1.SpaceSpecExecution{
						Domains: []v1alpha1.SpaceDomain{
							{Domain: "example.com"},
							{Domain: "other-example.com", Default: true},
						},
					},
				},
			},
			args: []string{"set-default-domain", space, "example.com"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "len(domains)", 2, len(space.Spec.Execution.Domains))
				testutil.AssertEqual(t, "domains", "example.com", space.Spec.Execution.Domains[0].Domain)
				testutil.AssertEqual(t, "default", true, space.Spec.Execution.Domains[0].Default)
				testutil.AssertEqual(t, "unsets previous default", false, space.Spec.Execution.Domains[1].Default)
			},
		},

		"set-default-domain invalid": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					Execution: v1alpha1.SpaceSpecExecution{
						Domains: []v1alpha1.SpaceDomain{
							{Domain: "example.com"},
						},
					},
				},
			},
			wantErr: errors.New("failed to find domain other-example.com"),
			args:    []string{"set-default-domain", space, "other-example.com"},
		},

		"remove-domain valid": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					Execution: v1alpha1.SpaceSpecExecution{
						Domains: []v1alpha1.SpaceDomain{
							{Domain: "example.com"},
							{Domain: "other-example.com"},
						},
					},
				},
			},
			args: []string{"remove-domain", space, "other-example.com"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "len(domains)", 1, len(space.Spec.Execution.Domains))
				testutil.AssertEqual(t, "domains", "example.com", space.Spec.Execution.Domains[0].Domain)
			},
		},

		"set-build-service-account valid": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					Security: v1alpha1.SpaceSpecSecurity{
						BuildServiceAccount: "some-service-account",
					},
				},
			},
			args: []string{"set-build-service-account", space, "some-other-service-account"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "build-service-account", "some-other-service-account", space.Spec.Security.BuildServiceAccount)
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeSpaces := fake.NewFakeClient(ctrl)

			output := tc.space.DeepCopy()
			fakeSpaces.EXPECT().Transform(space, gomock.Any()).DoAndReturn(func(spaceName string, transformer spaces.Mutator) (*v1alpha1.Space, error) {
				if err := transformer(output); err != nil {
					return nil, err
				}
				return output, nil
			})

			buffer := &bytes.Buffer{}

			c := NewConfigSpaceCommand(&config.KfParams{}, fakeSpaces)
			c.SetOutput(buffer)
			c.SetArgs(tc.args)

			gotErr := c.Execute()
			if tc.wantErr != nil || gotErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
				return
			}

			if tc.validate != nil {
				tc.validate(t, output)
			}

			ctrl.Finish()
		})
	}
}

func TestNewConfigSpaceCommand_accessors(t *testing.T) {
	space := v1alpha1.Space{
		Spec: v1alpha1.SpaceSpec{
			Security: v1alpha1.SpaceSpecSecurity{
				BuildServiceAccount: "some-service-account",
			},
			BuildpackBuild: v1alpha1.SpaceSpecBuildpackBuild{
				ContainerRegistry: "gcr.io/foo",
				BuilderImage:      "gcr.io/buildpack-builder:latest",
				Env: envutil.MapToEnvVars(map[string]string{
					"JAVA_VERSION": "11",
					"BAR":          "BAZZ",
				}),
			},
			Execution: v1alpha1.SpaceSpecExecution{
				Env: envutil.MapToEnvVars(map[string]string{
					"PROFILE": "development",
					"BAR":     "BAZZ",
				}),
				Domains: []v1alpha1.SpaceDomain{
					{Domain: "example.com", Default: true},
					{Domain: "other-example.com"},
				},
			},
		},
	}

	cases := map[string]struct {
		args       []string
		space      v1alpha1.Space
		wantErr    error
		wantOutput string
	}{
		"get-execution-env valid": {
			args:  []string{"get-execution-env", "space-name"},
			space: space,
			wantOutput: `- name: BAR
  value: BAZZ
- name: PROFILE
  value: development
`,
		},
		"get-buildpack-env valid": {
			args:  []string{"get-buildpack-env", "space-name"},
			space: space,
			wantOutput: `- name: BAR
  value: BAZZ
- name: JAVA_VERSION
  value: "11"
`,
		},
		"get-buildpack-builder valid": {
			args:       []string{"get-buildpack-builder", "space-name"},
			space:      space,
			wantOutput: "gcr.io/buildpack-builder:latest\n",
		},
		"get-container-registry valid": {
			args:       []string{"get-container-registry", "space-name"},
			space:      space,
			wantOutput: "gcr.io/foo\n",
		},
		"get-domains valid": {
			args:  []string{"get-domains", "space-name"},
			space: space,
			wantOutput: `- default: true
  domain: example.com
- domain: other-example.com
`,
		},
		"get-build-service-account valid": {
			args:       []string{"get-build-service-account", "space-name"},
			space:      space,
			wantOutput: "some-service-account\n",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeSpaces := fake.NewFakeClient(ctrl)

			fakeSpaces.EXPECT().Get("space-name").Return(&tc.space, nil)

			buffer := &bytes.Buffer{}

			c := NewConfigSpaceCommand(&config.KfParams{}, fakeSpaces)
			c.SetOutput(buffer)
			c.SetArgs(tc.args)

			gotErr := c.Execute()
			if tc.wantErr != nil || gotErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
				return
			}

			testutil.AssertEqual(t, "output", tc.wantOutput, buffer.String())
			ctrl.Finish()
		})
	}
}
