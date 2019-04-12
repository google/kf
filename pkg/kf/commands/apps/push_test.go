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
	"path/filepath"
	"strconv"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/fake"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	"github.com/golang/mock/gomock"
)

func TestPushCommand(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		args              []string
		namespace         string
		containerRegistry string
		dockerImage       string
		buildpack         string
		path              string
		serviceAccount    string
		grpc              bool
		wantErr           error
		pusherErr         error
		envVars           []string
	}{
		"uses configured properties": {
			namespace:         "some-namespace",
			args:              []string{"app-name"},
			containerRegistry: "some-reg.io",
			dockerImage:       "some-docker-image",
			serviceAccount:    "some-service-account",
			path:              "some-path",
			grpc:              true,
			buildpack:         "some-buildpack",
			envVars:           []string{"env1=val1", "env2=val2"},
		},
		"service create error": {
			args:              []string{"app-name"},
			wantErr:           errors.New("some error"),
			pusherErr:         errors.New("some error"),
			containerRegistry: "some-reg.io",
			serviceAccount:    "some-service-account",
			path:              "some-path",
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakePusher := fake.NewFakePusher(ctrl)

			fakePusher.
				EXPECT().
				Push(gomock.Any(), gomock.Any()).
				DoAndReturn(func(appName string, opts ...kf.PushOption) error {
					if p := kf.PushOptions(opts).Path(); filepath.Base(p) != tc.path {
						t.Fatalf("expected path %s, got %s", filepath.Base(tc.path), p)
					}
					if p := kf.PushOptions(opts).Path(); !filepath.IsAbs(p) {
						t.Fatalf("expected path to be an absolute: %s", p)
					}

					testutil.AssertEqual(t, "app name", tc.args[0], appName)
					testutil.AssertEqual(t, "namespace", tc.namespace, kf.PushOptions(opts).Namespace())
					testutil.AssertEqual(t, "container registry", tc.containerRegistry, kf.PushOptions(opts).ContainerRegistry())
					testutil.AssertEqual(t, "docker image", tc.dockerImage, kf.PushOptions(opts).DockerImage())
					testutil.AssertEqual(t, "buildpack", tc.buildpack, kf.PushOptions(opts).Buildpack())
					testutil.AssertEqual(t, "service account", tc.serviceAccount, kf.PushOptions(opts).ServiceAccount())
					testutil.AssertEqual(t, "grpc", tc.grpc, kf.PushOptions(opts).Grpc())
					testutil.AssertEqual(t, "env vars", tc.envVars, kf.PushOptions(opts).EnvironmentVariables())

					return tc.pusherErr
				})

			buffer := &bytes.Buffer{}

			c := NewPushCommand(&config.KfParams{
				Namespace: tc.namespace,
				Output:    buffer,
			}, fakePusher)

			c.Flags().Set("container-registry", tc.containerRegistry)
			c.Flags().Set("docker-image", tc.dockerImage)
			c.Flags().Set("service-account", tc.serviceAccount)
			c.Flags().Set("path", tc.path)
			c.Flags().Set("grpc", strconv.FormatBool(tc.grpc))
			c.Flags().Set("buildpack", tc.buildpack)

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
