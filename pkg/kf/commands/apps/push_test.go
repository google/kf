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
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/apps"
	appsfake "github.com/google/kf/pkg/kf/apps/fake"
	"github.com/google/kf/pkg/kf/commands/config"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	svbFake "github.com/google/kf/pkg/kf/service-bindings/fake"
	"github.com/google/kf/pkg/kf/testutil"
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func dummyBindingInstance(appName, instanceName string) *v1beta1.ServiceBinding {
	instance := v1beta1.ServiceBinding{}
	instance.Name = fmt.Sprintf("kf-binding-%s-%s", appName, instanceName)

	return &instance
}

func TestPushCommand(t *testing.T) {
	t.Parallel()

	wantMemory := resource.MustParse("2Gi")
	wantDiskQuota := resource.MustParse("2Gi")
	wantCPU := resource.MustParse("2")

	defaultTCPHealthCheck := &corev1.Probe{
		Handler: corev1.Handler{
			TCPSocket: &corev1.TCPSocketAction{},
		},
	}

	defaultSpaceSpecExecution := v1alpha1.SpaceSpecExecution{
		Domains: []v1alpha1.SpaceDomain{
			{Domain: "example.com", Default: true},
		},
	}

	defaultOptions := []apps.PushOption{
		apps.WithPushHealthCheck(defaultTCPHealthCheck),
		apps.WithPushDefaultRouteDomain("example.com"),
	}

	for tn, tc := range map[string]struct {
		args            []string
		namespace       string
		wantErr         error
		pusherErr       error
		srcImageBuilder SrcImageBuilderFunc
		wantImagePrefix string
		targetSpace     *v1alpha1.Space
		wantOpts        []apps.PushOption
		setup           func(t *testing.T, f *svbFake.FakeClientInterface)
	}{
		"uses configured properties": {
			namespace: "some-namespace",
			args: []string{
				"example-app",
				"--buildpack", "some-buildpack",
				"--grpc",
				"--env", "env1=val1",
				"-e", "env2=val2",
				"--container-registry", "some-reg.io",
				"--instances", "1",
				"--path", "testdata/example-app",
				"--no-start",
				"-u", "http",
				"-t", "28",
				"-s", "cflinuxfs3",
				"--entrypoint", "start-web.sh",
				"--args", "a",
				"--args", "b",
			},
			wantImagePrefix: "some-reg.io/src-some-namespace-example-app",
			srcImageBuilder: func(dir, srcImage string, rebase bool) error {
				testutil.AssertEqual(t, "path", true, strings.Contains(dir, "example-app"))
				testutil.AssertEqual(t, "path is abs", true, filepath.IsAbs(dir))
				return nil
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushGrpc(true),
				apps.WithPushBuildpack("some-buildpack"),
				apps.WithPushStack("cflinuxfs3"),
				apps.WithPushEnvironmentVariables(map[string]string{"env1": "val1", "env2": "val2"}),
				apps.WithPushNoStart(true),
				apps.WithPushExactScale(intPtr(1)),
				apps.WithPushArgs([]string{"a", "b"}),
				apps.WithPushCommand([]string{"start-web.sh"}),
				apps.WithPushHealthCheck(&corev1.Probe{
					TimeoutSeconds: 28,
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{},
					},
				}),
			),
		},
		"uses current working directory for empty path": {
			namespace: "some-namespace",
			args: []string{
				"app-name",
			},
			srcImageBuilder: func(dir, srcImage string, rebase bool) error {
				cwd, err := os.Getwd()
				testutil.AssertNil(t, "cwd err", err)
				testutil.AssertEqual(t, "path", cwd, dir)
				return nil
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
			),
		},
		"custom-source": {
			namespace: "some-namespace",
			args: []string{
				"app-name",
				"--source-image", "custom-reg.io/source-image:latest",
			},
			wantImagePrefix: "custom-reg.io/source-image:latest",
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
			),
		},
		"override manifest instances": {
			namespace: "some-namespace",
			args: []string{
				"instances-app",
				"--instances=11",
				"--source-image", "custom-reg.io/source-image:latest",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushExactScale(intPtr(11)),
			),
		},
		"instances from manifest": {
			namespace: "some-namespace",
			args: []string{
				"instances-app",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushExactScale(intPtr(9)),
			),
		},
		"override manifest min instances": {
			namespace: "some-namespace",
			args: []string{
				"autoscaling-instances-app",
				"--min-scale=11",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushMinScale(intPtr(11)),
				apps.WithPushMaxScale(intPtr(11)),
			),
		},
		"override manifest max instances": {
			namespace: "some-namespace",
			args: []string{
				"autoscaling-instances-app",
				"--max-scale=13",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
				// Manifest has 9 for min
				apps.WithPushMinScale(intPtr(9)),
				apps.WithPushMaxScale(intPtr(13)),
			),
		},
		"min and max instances from manifest": {
			namespace: "some-namespace",
			args: []string{
				"autoscaling-instances-app",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushMinScale(intPtr(9)),
				apps.WithPushMaxScale(intPtr(11)),
			),
		},
		"bind-service-instance": {
			namespace: "some-namespace",
			args: []string{
				"app-name",
				"--manifest", "testdata/manifest-services.yaml",
			},
			srcImageBuilder: func(dir, srcImage string, rebase bool) error {
				cwd, err := os.Getwd()
				testutil.AssertNil(t, "cwd err", err)
				testutil.AssertEqual(t, "path", cwd, dir)
				return nil
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushServiceBindings([]v1alpha1.AppSpecServiceBinding{
					{
						Instance:    "some-service-instance",
						BindingName: "some-service-instance",
					},
				}),
			),
		},
		"service create error": {
			namespace: "default",
			args: []string{
				"app-name",
				"--container-registry", "some-reg.io",
			},
			wantErr:         errors.New("some error"),
			pusherErr:       errors.New("some error"),
			wantImagePrefix: "some-reg.io/src-default-app-name",
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("default"),
			),
		},
		"namespace is not provided": {
			args:    []string{"app-name"},
			wantErr: errors.New(utils.EmptyNamespaceError),
		},
		"container-registry comes from space": {
			namespace: "some-namespace",
			args: []string{
				"buildpack-app",
				"--manifest", "testdata/manifest.yml",
			},
			targetSpace: &v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					Execution: defaultSpaceSpecExecution,
					BuildpackBuild: v1alpha1.SpaceSpecBuildpackBuild{
						ContainerRegistry: "space-reg.io",
					},
				},
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushBuildpack("java,tomcat"),
			),
		},
		"SrcImageBuilder returns an error": {
			namespace: "some-namespace",
			args:      []string{"app-name"},
			wantErr:   errors.New("some error"),
			srcImageBuilder: func(dir, srcImage string, rebase bool) error {
				return errors.New("some error")
			},
		},
		"invalid environment variable, returns error": {
			namespace: "some-namespace",
			args: []string{
				"app-name",
				"--env", "invalid",
			},
			wantErr: errors.New("malformed environment variable: invalid"),
		},
		"container image": {
			namespace: "some-namespace",
			args: []string{
				"app-name",
				"--docker-image", "some-image",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushContainerImage("some-image"),
			),
		},
		"container image with env vars": {
			namespace: "some-namespace",
			args: []string{
				"app-name",
				"--docker-image", "some-image",
				"--env", "WHATNOW=BROWNCOW",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushContainerImage("some-image"),
				apps.WithPushEnvironmentVariables(map[string]string{"WHATNOW": "BROWNCOW"}),
			),
		},
		"invalid buildpack and container image": {
			namespace: "some-namespace",
			args: []string{
				"buildpack-app",
				"--docker-image", "some-image",
				"--buildpack", "some-buildpack",
				"--manifest", "testdata/manifest.yml",
			},
			wantErr: errors.New("cannot use buildpack and docker image simultaneously"),
		},
		"invalid container registry and container image": {
			namespace: "some-namespace",
			args: []string{
				"buildpack-app",
				"--docker-image", "some-image",
				"--container-registry", "some-registry",
				"--manifest", "testdata/manifest.yml",
			},
			wantErr: errors.New("--container-registry can only be used with source pushes, not containers"),
		},
		"invalid path and container image": {
			namespace: "some-namespace",
			args: []string{
				"auto-buildpack-app",
				"--docker-image", "some-image",
				"--manifest", "testdata/manifest.yml",
			},
			wantErr: errors.New("cannot use path and docker image simultaneously"),
		},
		"docker app from manifest": {
			namespace: "some-namespace",
			args: []string{
				"docker-app",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushContainerImage("gcr.io/docker-app"),
			),
		},
		"buildpack app from manifest": {
			namespace: "some-namespace",
			args: []string{
				"buildpack-app",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushBuildpack("java,tomcat"),
			),
		},
		"manifest missing app": {
			namespace: "some-namespace",
			args: []string{
				"missing-app",
				"--manifest", "testdata/manifest.yml",
			},
			wantErr: errors.New("no app missing-app found in the Manifest"),
		},
		"create and map routes from manifest": {
			namespace: "some-namespace",
			args: []string{
				"routes-app",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushRoutes(buildTestRoutes()),
				apps.WithPushDefaultRouteDomain(""),
			),
		},
		"create and map default routes": {
			namespace: "some-namespace",
			targetSpace: &v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					Execution: v1alpha1.SpaceSpecExecution{
						Domains: []v1alpha1.SpaceDomain{
							{Domain: "wrong.example.com"},
							{Domain: "right.example.com", Default: true},
						},
					},
				},
			},
			args: []string{
				"routes-app",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushDefaultRouteDomain("right.example.com"),
			),
		},
		"no-route prevents default route": {
			namespace: "some-namespace",
			args: []string{
				"routes-app",
				"--no-route",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushRoutes(nil),
				apps.WithPushDefaultRouteDomain(""),
			),
		},
		"no-route overrides manifest": {
			namespace: "some-namespace",
			args: []string{
				"routes-app",
				"--manifest", "testdata/manifest.yml",
				"--no-route",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushRoutes(nil),
				apps.WithPushDefaultRouteDomain(""),
			),
		},
		"random-route and no-route both set": {
			namespace: "some-namespace",
			args: []string{
				"random-route-app",
				"--manifest", "testdata/manifest.yml",
				"--no-route",
			},
			wantErr: errors.New("can not use random-route and no-route together"),
		},
		"random-route overrides manifest": {
			namespace: "some-namespace",
			args: []string{
				"routes-app",
				"--manifest", "testdata/manifest.yml",
				"--random-route",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushRoutes(buildTestRoutes()),
				apps.WithPushRandomRouteDomain("example.com"),
				apps.WithPushDefaultRouteDomain(""),
			),
		},
		"create and map routes from flags": {
			namespace: "some-namespace",
			args: []string{
				"routes-app",
				"--route=https://withscheme.example.com/path1",
				"--route=noscheme.example.com",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushRoutes([]v1alpha1.RouteSpecFields{
					buildRoute("withscheme", "example.com", "/path1"),
					buildRoute("noscheme", "example.com", ""),
				}),
				apps.WithPushDefaultRouteDomain(""),
			),
		},
		"http-health-check from manifest": {
			namespace: "some-namespace",
			args: []string{
				"http-health-check-app",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushContainerImage("gcr.io/http-health-check-app"),
				apps.WithPushHealthCheck(&corev1.Probe{
					TimeoutSeconds: 42,
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{Path: "/healthz"},
					},
				}),
			),
		},
		"tcp-health-check from manifest": {
			namespace: "some-namespace",
			args: []string{
				"tcp-health-check-app",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushContainerImage("gcr.io/tcp-health-check-app"),
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushHealthCheck(&corev1.Probe{
					TimeoutSeconds: 33,
					Handler: corev1.Handler{
						TCPSocket: &corev1.TCPSocketAction{},
					},
				}),
			),
		},
		"bad timeout": {
			namespace: "some-namespace",
			args: []string{
				"tcp-health-check-app",
				"-t", "-1",
			},
			wantErr: errors.New("health check timeouts can't be negative"),
		},
		"resource requests from manifest": {
			namespace: "some-namespace",
			args: []string{
				"resources-app",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushMemory(&wantMemory),
				apps.WithPushDiskQuota(&wantDiskQuota),
				apps.WithPushCPU(&wantCPU),
			),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			if tc.srcImageBuilder == nil {
				tc.srcImageBuilder = func(dir, srcImage string, rebase bool) error { return nil }
			}

			ctrl := gomock.NewController(t)
			fakeApps := appsfake.NewFakeClient(ctrl)
			fakePusher := appsfake.NewFakePusher(ctrl)
			svbClient := svbFake.NewFakeClientInterface(ctrl)

			fakePusher.
				EXPECT().
				Push(gomock.Any(), gomock.Any()).
				DoAndReturn(func(appName string, opts ...apps.PushOption) error {
					testutil.AssertEqual(t, "app name", tc.args[0], appName)

					expectOpts := apps.PushOptions(tc.wantOpts)
					actualOpts := apps.PushOptions(opts)
					testutil.AssertEqual(t, "namespace", expectOpts.Namespace(), actualOpts.Namespace())
					testutil.AssertEqual(t, "buildpack", expectOpts.Buildpack(), actualOpts.Buildpack())
					testutil.AssertEqual(t, "grpc", expectOpts.Grpc(), actualOpts.Grpc())
					testutil.AssertEqual(t, "env vars", expectOpts.EnvironmentVariables(), actualOpts.EnvironmentVariables())
					testutil.AssertEqual(t, "exact scale bound", expectOpts.ExactScale(), actualOpts.ExactScale())
					testutil.AssertEqual(t, "min scale bound", expectOpts.MinScale(), actualOpts.MinScale())
					testutil.AssertEqual(t, "max scale bound", expectOpts.MaxScale(), actualOpts.MaxScale())
					testutil.AssertEqual(t, "no start", expectOpts.NoStart(), actualOpts.NoStart())
					testutil.AssertEqual(t, "routes", expectOpts.Routes(), actualOpts.Routes())
					testutil.AssertEqual(t, "memory requests", expectOpts.Memory(), actualOpts.Memory())
					testutil.AssertEqual(t, "storage requests", expectOpts.DiskQuota(), actualOpts.DiskQuota())
					testutil.AssertEqual(t, "cpu requests", expectOpts.CPU(), actualOpts.CPU())
					testutil.AssertEqual(t, "health check", expectOpts.HealthCheck(), actualOpts.HealthCheck())
					testutil.AssertEqual(t, "default route", expectOpts.DefaultRouteDomain(), actualOpts.DefaultRouteDomain())
					testutil.AssertEqual(t, "random route", expectOpts.RandomRouteDomain(), actualOpts.RandomRouteDomain())
					testutil.AssertEqual(t, "command", expectOpts.Command(), actualOpts.Command())
					testutil.AssertEqual(t, "args", expectOpts.Args(), actualOpts.Args())

					if !strings.HasPrefix(actualOpts.SourceImage(), tc.wantImagePrefix) {
						t.Errorf("Wanted srcImage to start with %s got: %s", tc.wantImagePrefix, actualOpts.SourceImage())
					}
					testutil.AssertEqual(t, "containerImage", expectOpts.ContainerImage(), actualOpts.ContainerImage())

					return tc.pusherErr
				})

			params := &config.KfParams{
				Namespace:   tc.namespace,
				TargetSpace: tc.targetSpace,
			}

			if params.TargetSpace == nil {
				params.SetTargetSpaceToDefault()
				params.TargetSpace.Spec.Execution = defaultSpaceSpecExecution
			}

			if tc.setup != nil {
				tc.setup(t, svbClient)
			}

			c := NewPushCommand(params, fakeApps, fakePusher, tc.srcImageBuilder, svbClient)
			buffer := &bytes.Buffer{}
			c.SetOutput(buffer)
			c.SetArgs(tc.args)
			_, gotErr := c.ExecuteC()
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

func buildRoute(hostname, domain, path string) v1alpha1.RouteSpecFields {
	return v1alpha1.RouteSpecFields{
		Hostname: hostname,
		Domain:   domain,
		Path:     path,
	}
}

func buildTestRoutes() []v1alpha1.RouteSpecFields {
	return []v1alpha1.RouteSpecFields{
		buildRoute("", "example.com", ""),
		buildRoute("", "www.example.com", "/foo"),
		buildRoute("host", "example.com", "/foo"),
	}
}

func intPtr(i int) *int {
	return &i
}
