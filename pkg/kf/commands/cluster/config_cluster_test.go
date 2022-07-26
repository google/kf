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

package cluster

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/configmaps"
	"github.com/google/kf/v2/pkg/kf/configmaps/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

func TestNewConfigClusterCommand(t *testing.T) {
	defaultsConfigMap := *configMapFromTestFile(t, kfconfig.DefaultsConfigName)
	ffName := "enable_route_services"
	cases := map[string]struct {
		args      []string
		configMap v1.ConfigMap
		wantErr   error
		validate  func(*testing.T, *v1.ConfigMap)
	}{
		"set-feature-flag valid": {
			configMap: defaultsConfigMap,
			args:      []string{"set-feature-flag", ffName, "true"},
			validate: func(t *testing.T, cm *v1.ConfigMap) {
				newDefaults, err := kfconfig.NewDefaultsConfigFromConfigMap(cm)
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "route services feature flag", true, newDefaults.FeatureFlags.RouteServices().IsEnabled())
			},
		},

		"unset-feature-flag valid": {
			configMap: defaultsConfigMap,
			args:      []string{"unset-feature-flag", ffName},
			validate: func(t *testing.T, cm *v1.ConfigMap) {
				newDefaults, err := kfconfig.NewDefaultsConfigFromConfigMap(cm)
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "route services feature flag", true, newDefaults.FeatureFlags.RouteServices().IsDisabled())
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeConfigs := fake.NewFakeClient(ctrl)

			output := tc.configMap.DeepCopy()
			fakeConfigs.EXPECT().Transform(gomock.Any(), v1alpha1.KfNamespace, kfconfig.DefaultsConfigName, gomock.Any()).DoAndReturn(func(ctx context.Context, namespace, configMapName string, transformer configmaps.Mutator) (*v1.ConfigMap, error) {
				if err := transformer(output); err != nil {
					return nil, err
				}
				return output, nil
			})

			buffer := &bytes.Buffer{}

			c := NewConfigClusterCommand(&config.KfParams{}, fakeConfigs)
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

func TestNewConfigClusterCommand_accessors(t *testing.T) {
	cm := configMapFromTestFile(t, kfconfig.DefaultsConfigName)

	cases := map[string]struct {
		args       []string
		wantErr    error
		wantOutput string
	}{
		"get-feature-flags valid": {
			args: []string{"get-feature-flags"},
			wantOutput: `disable_custom_builds: false
enable_appdevexperience_builds: false
enable_custom_buildpacks: true
enable_custom_stacks: true
enable_dockerfile_builds: true
enable_route_services: false
`,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeConfigs := fake.NewFakeClient(ctrl)

			fakeConfigs.EXPECT().Get(gomock.Any(), v1alpha1.KfNamespace, kfconfig.DefaultsConfigName).Return(cm, nil)

			buffer := &bytes.Buffer{}

			c := NewConfigClusterCommand(&config.KfParams{}, fakeConfigs)
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
	obj := &v1.ConfigMap{}

	wrapper := DiffWrapper(os.Stdout, func(cm *v1.ConfigMap) error {
		// don't mutate the object
		return nil
	})

	wrapper(obj)

	// Output: No changes
}

func ExampleDiffWrapper_changes() {
	obj := &v1.ConfigMap{}
	obj.Data = map[string]string{"some-key": "some-value"}

	contents := &bytes.Buffer{}
	wrapper := DiffWrapper(contents, func(obj *v1.ConfigMap) error {
		obj.Data = map[string]string{"some-key": "new-value"}
		return nil
	})

	fmt.Println("Error:", wrapper(obj))
	firstLine := strings.Split(contents.String(), "\n")[0]
	fmt.Println("First line:", firstLine)

	// Output: Error: <nil>
	// First line: ConfigMap Diff (-old +new):
}

func ExampleDiffWrapper_err() {
	obj := &v1.ConfigMap{}
	wrapper := DiffWrapper(os.Stdout, func(_ *v1.ConfigMap) error {
		return errors.New("some-error")
	})

	fmt.Println(wrapper(obj))

	// Output: some-error
}

// configMapFromTestFile creates a v1.ConfigMap from the local YAML file under /testdata.
func configMapFromTestFile(t *testing.T, name string) *v1.ConfigMap {
	t.Helper()

	b, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s.yaml", name))
	if err != nil {
		t.Fatalf("ReadFile() = %v", err)
	}

	var orig v1.ConfigMap

	if err := yaml.Unmarshal(b, &orig); err != nil {
		t.Fatalf("yaml.Unmarshal() = %v", err)
	}

	return &orig
}
