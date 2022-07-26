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
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/internal/envutil"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/spaces"
	"github.com/google/kf/v2/pkg/kf/spaces/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewConfigSpaceCommand(t *testing.T) {
	t.Skip("b/236783219")
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
			wantErr: errors.New("accepts between 1 and 2 arg(s), received 0"),
		},

		"set-container-registry valid": {
			args: []string{"set-container-registry", space, "gcr.io/foo"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "container registry", "gcr.io/foo", space.Spec.BuildConfig.ContainerRegistry)
			},
		},

		"set with targeted space": {
			args: []string{"set-container-registry", "gcr.io/foo"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "container registry", "gcr.io/foo", space.Spec.BuildConfig.ContainerRegistry)
			},
			space: v1alpha1.Space{
				ObjectMeta: metav1.ObjectMeta{
					Name: space,
				},
			},
		},

		"set-env valid": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					RuntimeConfig: v1alpha1.SpaceSpecRuntimeConfig{
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
				}, envutil.EnvVarsToMap(space.Spec.RuntimeConfig.Env))
			},
		},

		"unset-env valid": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					RuntimeConfig: v1alpha1.SpaceSpecRuntimeConfig{
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
				}, envutil.EnvVarsToMap(space.Spec.RuntimeConfig.Env))
			},
		},

		"set-buildpack-env valid": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					BuildConfig: v1alpha1.SpaceSpecBuildConfig{
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
				}, envutil.EnvVarsToMap(space.Spec.BuildConfig.Env))
			},
		},

		"unset-buildpack-env valid": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					BuildConfig: v1alpha1.SpaceSpecBuildConfig{
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
				}, envutil.EnvVarsToMap(space.Spec.BuildConfig.Env))
			},
		},

		"append-domain valid": {
			args: []string{"append-domain", space, "example.com"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "len(domains)", 1, len(space.Spec.NetworkConfig.Domains))
				testutil.AssertEqual(t, "domains", "example.com", space.Spec.NetworkConfig.Domains[0].Domain)
			},
		},

		"set-default-domain valid": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					NetworkConfig: v1alpha1.SpaceSpecNetworkConfig{
						Domains: []v1alpha1.SpaceDomain{
							{Domain: "example.com"},
							{Domain: "other-example.com"},
						},
					},
				},
			},
			args: []string{"set-default-domain", space, "other-example.com"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "len(domains)", 2, len(space.Spec.NetworkConfig.Domains))
				testutil.AssertEqual(t, "domains[0].domain", "other-example.com", space.Spec.NetworkConfig.Domains[0].Domain)
				testutil.AssertEqual(t, "domains[1].domain", "example.com", space.Spec.NetworkConfig.Domains[1].Domain)
			},
		},

		"set-default-domain not-present": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					NetworkConfig: v1alpha1.SpaceSpecNetworkConfig{
						Domains: []v1alpha1.SpaceDomain{
							{Domain: "example.com"},
						},
					},
				},
			},
			args: []string{"set-default-domain", space, "other-example.com"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "len(domains)", 2, len(space.Spec.NetworkConfig.Domains))
				testutil.AssertEqual(t, "domains[0].domain", "other-example.com", space.Spec.NetworkConfig.Domains[0].Domain)
				testutil.AssertEqual(t, "domains[1].domain", "example.com", space.Spec.NetworkConfig.Domains[1].Domain)
			},
		},

		"remove-domain valid": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					NetworkConfig: v1alpha1.SpaceSpecNetworkConfig{
						Domains: []v1alpha1.SpaceDomain{
							{Domain: "example.com"},
							{Domain: "other-example.com"},
						},
					},
				},
			},
			args: []string{"remove-domain", space, "other-example.com"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "len(domains)", 1, len(space.Spec.NetworkConfig.Domains))
				testutil.AssertEqual(t, "domains", "example.com", space.Spec.NetworkConfig.Domains[0].Domain)
			},
		},

		"set-build-service-account valid": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					BuildConfig: v1alpha1.SpaceSpecBuildConfig{
						ServiceAccount: "some-service-account",
					},
				},
			},
			args: []string{"set-build-service-account", space, "some-other-service-account"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "build-service-account", "some-other-service-account", space.Spec.BuildConfig.ServiceAccount)
			},
		},

		"set-app-ingress-policy DenyAll": {
			space: v1alpha1.Space{},
			args:  []string{"set-app-ingress-policy", space, "DenyAll"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "policy", "DenyAll", space.Spec.NetworkConfig.AppNetworkPolicy.Ingress)
			},
		},
		"set-app-egress-policy DenyAll": {
			space: v1alpha1.Space{},
			args:  []string{"set-app-egress-policy", space, "DenyAll"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "policy", "DenyAll", space.Spec.NetworkConfig.AppNetworkPolicy.Egress)
			},
		},
		"set-build-ingress-policy DenyAll": {
			space: v1alpha1.Space{},
			args:  []string{"set-build-ingress-policy", space, "DenyAll"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "policy", "DenyAll", space.Spec.NetworkConfig.BuildNetworkPolicy.Ingress)
			},
		},
		"set-build-egress-policy DenyAll": {
			space: v1alpha1.Space{},
			args:  []string{"set-build-egress-policy", space, "DenyAll"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "policy", "DenyAll", space.Spec.NetworkConfig.BuildNetworkPolicy.Egress)
			},
		},
		"set-nodeselector valid": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					RuntimeConfig: v1alpha1.SpaceSpecRuntimeConfig{},
				},
			},
			args: []string{"set-nodeselector", space, "DISKTYPE", "SSD"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "nodeselector", map[string]string{
					"DISKTYPE": "SSD",
				}, space.Spec.RuntimeConfig.NodeSelector)
			},
		},
		"unset-nodeselector valid": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					RuntimeConfig: v1alpha1.SpaceSpecRuntimeConfig{
						NodeSelector: map[string]string{
							"DISKTYPE": "HDD",
							"CPU":      "X86",
						},
					},
				},
			},
			args: []string{"unset-nodeselector", space, "CPU"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "nodeselector", map[string]string{
					"DISKTYPE": "HDD",
				}, space.Spec.RuntimeConfig.NodeSelector)
			},
		},
		"unset-nodeselector invalid name": {
			space: v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					RuntimeConfig: v1alpha1.SpaceSpecRuntimeConfig{
						NodeSelector: map[string]string{
							"DISKTYPE": "HDD",
						},
					},
				},
			},
			args: []string{"unset-nodeselector", space, "CPU"},
			validate: func(t *testing.T, space *v1alpha1.Space) {
				testutil.AssertEqual(t, "nodeselector", map[string]string{
					"DISKTYPE": "HDD",
				}, space.Spec.RuntimeConfig.NodeSelector)
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeSpaces := fake.NewFakeClient(ctrl)

			output := tc.space.DeepCopy()
			fakeSpaces.EXPECT().Transform(gomock.Any(), space, gomock.Any()).DoAndReturn(func(ctx context.Context, spaceName string, transformer spaces.Mutator) (*v1alpha1.Space, error) {
				if err := transformer(output); err != nil {
					return nil, err
				}
				return output, nil
			})

			fakeSpaces.EXPECT().WaitForConditionReadyTrue(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			buffer := &bytes.Buffer{}

			c := NewConfigSpaceCommand(&config.KfParams{
				Space: tc.space.GetName(),
			}, fakeSpaces)
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
		})
	}
}

func TestNewConfigSpaceCommand_accessors(t *testing.T) {
	space := v1alpha1.Space{
		ObjectMeta: metav1.ObjectMeta{
			Name: "space-name",
		},
		Spec: v1alpha1.SpaceSpec{
			BuildConfig: v1alpha1.SpaceSpecBuildConfig{
				ServiceAccount:    "some-service-account",
				ContainerRegistry: "gcr.io/foo",
				Env: envutil.MapToEnvVars(map[string]string{
					"JAVA_VERSION": "11",
					"BAR":          "BAZZ",
				}),
			},
			RuntimeConfig: v1alpha1.SpaceSpecRuntimeConfig{
				Env: envutil.MapToEnvVars(map[string]string{
					"PROFILE": "development",
					"BAR":     "BAZZ",
				}),
				NodeSelector: map[string]string{
					"DISKTYPE": "SSD",
					"CPU":      "X86",
				},
			},
			NetworkConfig: v1alpha1.SpaceSpecNetworkConfig{
				Domains: []v1alpha1.SpaceDomain{
					{Domain: "example.com"},
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
		"get-execution-env targeted space": {
			args:  []string{"get-execution-env"},
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
		"get-container-registry valid": {
			args:       []string{"get-container-registry", "space-name"},
			space:      space,
			wantOutput: "gcr.io/foo\n",
		},
		"get-domains valid": {
			args:  []string{"get-domains", "space-name"},
			space: space,
			wantOutput: `- domain: example.com
- domain: other-example.com
`,
		},
		"get-build-service-account valid": {
			args:       []string{"get-build-service-account", "space-name"},
			space:      space,
			wantOutput: "some-service-account\n",
		},
		"get-nodeselector valid": {
			args:  []string{"get-nodeselector", "space-name"},
			space: space,
			wantOutput: `CPU: X86
DISKTYPE: SSD
`,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeSpaces := fake.NewFakeClient(ctrl)

			fakeSpaces.EXPECT().Get(gomock.Any(), "space-name").Return(&tc.space, nil)

			buffer := &bytes.Buffer{}

			c := NewConfigSpaceCommand(&config.KfParams{
				Space: tc.space.GetName(),
			}, fakeSpaces)
			c.SetOutput(buffer)
			c.SetArgs(tc.args)

			gotErr := c.Execute()
			if tc.wantErr != nil || gotErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
				return
			}

			testutil.AssertEqual(t, "output", tc.wantOutput, buffer.String())
		})
	}
}

func ExampleDiffWrapper_noDiff() {
	obj := &v1alpha1.Space{}

	wrapper := DiffWrapper(os.Stdout, func(s *v1alpha1.Space) error {
		// don't mutate the object
		return nil
	})

	wrapper(obj)

	// Output: No changes
}

func ExampleDiffWrapper_changes() {
	obj := &v1alpha1.Space{}
	obj.Name = "opaque"

	contents := &bytes.Buffer{}
	wrapper := DiffWrapper(contents, func(obj *v1alpha1.Space) error {
		obj.Name = "docker-creds"
		return nil
	})

	fmt.Println("Error:", wrapper(obj))
	firstLine := strings.Split(contents.String(), "\n")[0]
	fmt.Println("First line:", firstLine)

	// Output: Error: <nil>
	// First line: Space Diff (-old +new):
}

func ExampleDiffWrapper_err() {
	obj := &v1alpha1.Space{}

	wrapper := DiffWrapper(os.Stdout, func(_ *v1alpha1.Space) error {
		return errors.New("some-error")
	})

	fmt.Println(wrapper(obj))

	// Output: some-error
}
