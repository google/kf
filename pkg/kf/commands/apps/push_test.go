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

package apps

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/kf"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/fake"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestPushCommand(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		args              []string
		namespace         string
		containerRegistry string
		buildpack         string
		path              string
		serviceAccount    string
		sourceImage       string
		grpc              bool
		wantErr           error
		pusherErr         error
		envVars           []string
		wantEnvMap        map[string]string
		srcImageBuilder   SrcImageBuilderFunc
		wantImagePrefix   string
	}{
		"uses configured properties": {
			namespace:         "some-namespace",
			args:              []string{"app-name"},
			containerRegistry: "some-reg.io",
			serviceAccount:    "some-service-account",
			path:              "some-path",
			grpc:              true,
			buildpack:         "some-buildpack",
			envVars:           []string{"env1=val1", "env2=val2"},
			wantEnvMap:        map[string]string{"env1": "val1", "env2": "val2"},
			wantImagePrefix:   "some-reg.io/src-some-namespace-app-name",
			srcImageBuilder: func(dir, srcImage string, rebase bool) error {
				testutil.AssertEqual(t, "path", true, strings.Contains(dir, "some-path"))
				testutil.AssertEqual(t, "path is abs", true, filepath.IsAbs(dir))
				return nil
			},
		},
		"uses current working directory for empty path": {
			args:              []string{"app-name"},
			containerRegistry: "some-reg.io",
			path:              "",
			srcImageBuilder: func(dir, srcImage string, rebase bool) error {
				cwd, err := os.Getwd()
				testutil.AssertNil(t, "cwd err", err)
				testutil.AssertEqual(t, "path", cwd, dir)
				return nil
			},
		},
		"custom-source": {
			namespace:         "some-namespace",
			args:              []string{"app-name"},
			containerRegistry: "some-reg.io",
			serviceAccount:    "some-service-account",
			grpc:              true,
			buildpack:         "some-buildpack",
			envVars:           []string{"env1=val1", "env2=val2"},
			wantEnvMap:        map[string]string{"env1": "val1", "env2": "val2"},
			sourceImage:       "custom-reg.io/source-image:latest",
			wantImagePrefix:   "custom-reg.io/source-image:latest",
		},
		"service create error": {
			namespace:         "default",
			args:              []string{"app-name"},
			wantErr:           errors.New("some error"),
			pusherErr:         errors.New("some error"),
			containerRegistry: "some-reg.io",
			serviceAccount:    "some-service-account",
			wantImagePrefix:   "some-reg.io/src-default-app-name",
		},
		"container-registry is not provided": {
			namespace:         "some-namespace",
			args:              []string{"app-name"},
			containerRegistry: "",
			wantErr:           errors.New("container-registry is required"),
		},
		"SrcImageBuilder returns an error": {
			args:              []string{"app-name"},
			containerRegistry: "some-reg.io",
			wantErr:           errors.New("some error"),
			srcImageBuilder: func(dir, srcImage string, rebase bool) error {
				return errors.New("some error")
			},
		},
		"invalid environment variable, returns error": {
			args:              []string{"app-name"},
			envVars:           []string{"invalid"},
			containerRegistry: "some-reg.io",
			wantErr:           errors.New("malformed environment variable: invalid"),
		},
	} {
		t.Run(tn, func(t *testing.T) {

			if tc.path != "" {
				os.MkdirAll(tc.path, 0755)
				defer os.RemoveAll(tc.path)
			}

			if tc.srcImageBuilder == nil {
				tc.srcImageBuilder = func(dir, srcImage string, rebase bool) error { return nil }
			}

			ctrl := gomock.NewController(t)
			fakePusher := fake.NewFakePusher(ctrl)

			fakePusher.
				EXPECT().
				Push(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(appName, srcImage string, opts ...kf.PushOption) error {
					testutil.AssertEqual(t, "app name", tc.args[0], appName)
					testutil.AssertEqual(t, "namespace", tc.namespace, kf.PushOptions(opts).Namespace())
					testutil.AssertEqual(t, "container registry", tc.containerRegistry, kf.PushOptions(opts).ContainerRegistry())
					testutil.AssertEqual(t, "buildpack", tc.buildpack, kf.PushOptions(opts).Buildpack())
					testutil.AssertEqual(t, "service account", tc.serviceAccount, kf.PushOptions(opts).ServiceAccount())
					testutil.AssertEqual(t, "grpc", tc.grpc, kf.PushOptions(opts).Grpc())
					testutil.AssertEqual(t, "env vars", tc.wantEnvMap, kf.PushOptions(opts).EnvironmentVariables())

					if !strings.HasPrefix(srcImage, tc.wantImagePrefix) {
						t.Errorf("Wanted srcImage to start with %s got: %s", tc.wantImagePrefix, srcImage)
					}

					return tc.pusherErr
				})

			c := NewPushCommand(&config.KfParams{
				Namespace: tc.namespace,
			},
				fakePusher,
				tc.srcImageBuilder,
			)
			buffer := &bytes.Buffer{}
			c.SetOutput(buffer)

			c.Flags().Set("container-registry", tc.containerRegistry)
			c.Flags().Set("service-account", tc.serviceAccount)
			c.Flags().Set("path", tc.path)
			c.Flags().Set("grpc", strconv.FormatBool(tc.grpc))
			c.Flags().Set("buildpack", tc.buildpack)
			c.Flags().Set("source-image", tc.sourceImage)

			for _, env := range tc.envVars {
				c.Flags().Set("env", env)
			}
			gotErr := c.RunE(c, tc.args)
			if tc.wantErr != nil || gotErr != nil {
				if fmt.Sprint(tc.wantErr) != fmt.Sprint(gotErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.wantErr, gotErr)
				}

				return
			}

			ctrl.Finish()
		})
	}
}
