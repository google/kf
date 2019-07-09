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
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/internal/envutil"
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
		"base wrong number of args": {
			args:    []string{"test"},
			wantErr: errors.New("accepts 0 arg(s), received 1"),
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
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeSpaces := fake.NewFakeClient(ctrl)

			output := tc.space.DeepCopy()
			fakeSpaces.EXPECT().Transform(space, gomock.Any()).DoAndReturn(func(spaceName string, transformer spaces.Mutator) error {
				return transformer(output)
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
