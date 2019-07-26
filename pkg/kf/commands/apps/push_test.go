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
	"github.com/google/kf/pkg/kf/commands/utils"
	servicebindings "github.com/google/kf/pkg/kf/service-bindings"
	svbFake "github.com/google/kf/pkg/kf/service-bindings/fake"
	"github.com/google/kf/pkg/kf/testutil"
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type routeParts struct {
	hostname string
	domain   string
	path     string
}

func dummyBindingInstance(appName, instanceName string) *v1beta1.ServiceBinding {
	instance := v1beta1.ServiceBinding{}
	instance.Name = fmt.Sprintf("kf-binding-%s-%s", appName, instanceName)

	return &instance
}

func TestPushCommand(t *testing.T) {
	t.Parallel()

	routes := createTestRoutes([]routeParts{
		{
			hostname: "",
			domain:   "example.com",
			path:     "",
		},
		{
			hostname: "",
			domain:   "www.example.com",
			path:     "/foo",
		},
		{
			hostname: "host",
			domain:   "example.com",
			path:     "/foo",
		},
	})

	wantMemory := resource.MustParse("2Gi")
	wantDiskQuota := resource.MustParse("2Gi")
	wantCPU := resource.MustParse("2")

	for tn, tc := range map[string]struct {
		args            []string
		namespace       string
		manifestFile    string
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
				"--service-account", "some-service-account",
				"--grpc",
				"--env", "env1=val1",
				"-e", "env2=val2",
				"--container-registry", "some-reg.io",
				"--instances", "1",
				"--path", "testdata/example-app",
				"--no-start",
			},
			wantImagePrefix: "some-reg.io/src-some-namespace-example-app",
			srcImageBuilder: func(dir, srcImage string, rebase bool) error {
				testutil.AssertEqual(t, "path", true, strings.Contains(dir, "example-app"))
				testutil.AssertEqual(t, "path is abs", true, filepath.IsAbs(dir))
				return nil
			},
			wantOpts: []apps.PushOption{
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushContainerRegistry("some-reg.io"),
				apps.WithPushServiceAccount("some-service-account"),
				apps.WithPushGrpc(true),
				apps.WithPushBuildpack("some-buildpack"),
				apps.WithPushEnvironmentVariables(map[string]string{"env1": "val1", "env2": "val2"}),
				apps.WithPushMinScale(1),
				apps.WithPushMaxScale(1),
				apps.WithPushNoStart(true),
			},
		},
		"uses current working directory for empty path": {
			namespace: "some-namespace",
			args: []string{
				"app-name",
				"--container-registry", "some-reg.io",
			},
			srcImageBuilder: func(dir, srcImage string, rebase bool) error {
				cwd, err := os.Getwd()
				testutil.AssertNil(t, "cwd err", err)
				testutil.AssertEqual(t, "path", cwd, dir)
				return nil
			},
			wantOpts: []apps.PushOption{
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushContainerRegistry("some-reg.io"),
				apps.WithPushMinScale(1),
				apps.WithPushMaxScale(1),
			},
		},
		"custom-source": {
			namespace: "some-namespace",
			args: []string{
				"app-name",
				"--container-registry", "some-reg.io",
				"--source-image", "custom-reg.io/source-image:latest",
			},
			wantImagePrefix: "custom-reg.io/source-image:latest",
			wantOpts: []apps.PushOption{
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushContainerRegistry("some-reg.io"),
				apps.WithPushMinScale(1),
				apps.WithPushMaxScale(1),
			},
		},
		"specify-instances": {
			namespace: "some-namespace",
			args: []string{
				"app-name",
				"--container-registry", "some-reg.io",
				"--instances", "2",
				"--source-image", "custom-reg.io/source-image:latest",
			},
			wantImagePrefix: "custom-reg.io/source-image:latest",
			wantOpts: []apps.PushOption{
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushMinScale(2),
				apps.WithPushMaxScale(2),
				apps.WithPushContainerRegistry("some-reg.io"),
			},
		},
		"bind-service-instance": {
			namespace:    "some-namespace",
			manifestFile: "testdata/manifest-services.yaml",
			args: []string{
				"app-name",
				"--container-registry", "some-reg.io",
				"--manifest", "testdata/manifest-services.yaml",
			},
			srcImageBuilder: func(dir, srcImage string, rebase bool) error {
				cwd, err := os.Getwd()
				testutil.AssertNil(t, "cwd err", err)
				testutil.AssertEqual(t, "path", cwd, dir)
				return nil
			},
			wantOpts: []apps.PushOption{
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushContainerRegistry("some-reg.io"),
				apps.WithPushMinScale(1),
				apps.WithPushMaxScale(1),
			},
			setup: func(t *testing.T, f *svbFake.FakeClientInterface) {
				f.EXPECT().GetOrCreate("some-service-instance", "app-name", gomock.Any()).Do(func(instance, app string, opts ...servicebindings.CreateOption) {
					config := servicebindings.CreateOptions(opts)
					testutil.AssertEqual(t, "params", map[string]interface{}{}, config.Params())
					testutil.AssertEqual(t, "namespace", "some-namespace", config.Namespace())
					testutil.AssertEqual(t, "binding-name", "some-service-instance", config.BindingName())
				}).Return(dummyBindingInstance("app-name", "some-service-instance"), true, nil)
			},
		},
		"service create error": {
			namespace:       "default",
			args:            []string{"app-name", "--container-registry", "some-reg.io"},
			wantErr:         errors.New("some error"),
			pusherErr:       errors.New("some error"),
			wantImagePrefix: "some-reg.io/src-default-app-name",
			wantOpts: []apps.PushOption{
				apps.WithPushNamespace("default"),
				apps.WithPushContainerRegistry("some-reg.io"),
				apps.WithPushMinScale(1),
				apps.WithPushMaxScale(1),
			},
		},
		"namespace is not provided": {
			args:    []string{"app-name"},
			wantErr: errors.New(utils.EmptyNamespaceError),
		},
		"container-registry is not provided": {
			namespace: "some-namespace",
			args:      []string{"app-name"},
			wantErr:   errors.New("container-registry is required for buildpack apps"),
		},
		"container-registry comes from space": {
			namespace: "some-namespace",
			args: []string{
				"buildpack-app",
				"--manifest", "testdata/manifest.yml",
			},
			targetSpace: &v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					BuildpackBuild: v1alpha1.SpaceSpecBuildpackBuild{
						ContainerRegistry: "space-reg.io",
					},
				},
			},
			wantOpts: []apps.PushOption{
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushContainerRegistry("space-reg.io"),
				apps.WithPushBuildpack("java,tomcat"),
				apps.WithPushMinScale(1),
				apps.WithPushMaxScale(1),
			},
		},
		"SrcImageBuilder returns an error": {
			namespace: "some-namespace",
			args:      []string{"app-name", "--container-registry", "some-reg.io"},
			wantErr:   errors.New("some error"),
			srcImageBuilder: func(dir, srcImage string, rebase bool) error {
				return errors.New("some error")
			},
		},
		"invalid environment variable, returns error": {
			namespace: "some-namespace",
			args: []string{
				"app-name",
				"--container-registry", "some-reg.io",
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
			wantOpts: []apps.PushOption{
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushContainerImage("some-image"),
				apps.WithPushMinScale(1),
				apps.WithPushMaxScale(1),
			},
		},
		"container image with env vars": {
			namespace: "some-namespace",
			args: []string{
				"app-name",
				"--docker-image", "some-image",
				"--env", "WHATNOW=BROWNCOW",
			},
			wantOpts: []apps.PushOption{
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushContainerImage("some-image"),
				apps.WithPushMinScale(1),
				apps.WithPushMaxScale(1),
				apps.WithPushEnvironmentVariables(map[string]string{"WHATNOW": "BROWNCOW"}),
			},
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
			wantOpts: []apps.PushOption{
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushContainerImage("gcr.io/docker-app"),
				apps.WithPushMinScale(1),
				apps.WithPushMaxScale(1),
			},
		},
		"buildpack app from manifest": {
			namespace: "some-namespace",
			args: []string{
				"buildpack-app",
				"--manifest", "testdata/manifest.yml",
				"--container-registry", "some-registry.io",
			},
			wantOpts: []apps.PushOption{
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushBuildpack("java,tomcat"),
				apps.WithPushContainerRegistry("some-registry.io"),
				apps.WithPushMinScale(1),
				apps.WithPushMaxScale(1),
			},
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
				"--container-registry", "some-registry.io",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: []apps.PushOption{
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushRoutes(routes),
				apps.WithPushContainerRegistry("some-registry.io"),
				apps.WithPushMinScale(1),
				apps.WithPushMaxScale(1),
			},
		},
		"resource requests from manifest": {
			namespace: "some-namespace",
			args: []string{
				"resources-app",
				"--manifest", "testdata/manifest.yml",
				"--container-registry", "some-registry.io",
			},
			wantOpts: []apps.PushOption{
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushContainerRegistry("some-registry.io"),
				apps.WithPushMinScale(1),
				apps.WithPushMaxScale(1),
				apps.WithPushMemory(&wantMemory),
				apps.WithPushDiskQuota(&wantDiskQuota),
				apps.WithPushCPU(&wantCPU),
			},
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
					testutil.AssertEqual(t, "container registry", expectOpts.ContainerRegistry(), actualOpts.ContainerRegistry())
					testutil.AssertEqual(t, "buildpack", expectOpts.Buildpack(), actualOpts.Buildpack())
					testutil.AssertEqual(t, "service account", expectOpts.ServiceAccount(), actualOpts.ServiceAccount())
					testutil.AssertEqual(t, "grpc", expectOpts.Grpc(), actualOpts.Grpc())
					testutil.AssertEqual(t, "env vars", expectOpts.EnvironmentVariables(), actualOpts.EnvironmentVariables())
					testutil.AssertEqual(t, "min scale bound", expectOpts.MinScale(), actualOpts.MinScale())
					testutil.AssertEqual(t, "max scale bound", expectOpts.MaxScale(), actualOpts.MaxScale())
					testutil.AssertEqual(t, "no start", expectOpts.NoStart(), actualOpts.NoStart())
					testutil.AssertEqual(t, "routes", expectOpts.Routes(), actualOpts.Routes())
					testutil.AssertEqual(t, "memory requests", expectOpts.Memory(), actualOpts.Memory())
					testutil.AssertEqual(t, "storage requests", expectOpts.DiskQuota(), actualOpts.DiskQuota())
					testutil.AssertEqual(t, "cpu requests", expectOpts.CPU(), actualOpts.CPU())

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

func createTestRoutes(routes []routeParts) []*v1alpha1.Route {
	newRoutes := []*v1alpha1.Route{}
	for _, route := range routes {
		r := &v1alpha1.Route{
			TypeMeta: metav1.TypeMeta{
				Kind: "Route",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "some-namespace",
				Name: v1alpha1.GenerateName(
					route.hostname,
					route.domain,
					route.path,
				),
			},
			Spec: v1alpha1.RouteSpec{
				Hostname:            route.hostname,
				Domain:              route.domain,
				Path:                route.path,
				KnativeServiceNames: []string{"routes-app"},
			},
		}
		newRoutes = append(newRoutes, r)
	}
	return newRoutes
}
