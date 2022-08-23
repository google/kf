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
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/internal/envutil"
	"github.com/google/kf/v2/pkg/kf/apps"
	appsfake "github.com/google/kf/v2/pkg/kf/apps/fake"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	injection "github.com/google/kf/v2/pkg/kf/injection/fake"
	"github.com/google/kf/v2/pkg/kf/manifest"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/sourceimage"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"knative.dev/pkg/ptr"
)

var someImage = "some-image"
var dockerAppImage = "gcr.io/docker-app"
var healthCheckApp = "gcr.io/http-health-check-app"

func bldPtr(build v1alpha1.BuildSpec) *v1alpha1.BuildSpec {
	return &build
}

func TestPushCommand(t *testing.T) {
	t.Skip("b/236783219")
	t.Parallel()

	wantMemory := resource.MustParse("2Gi")
	wantDiskQuota := resource.MustParse("2Gi")
	wantCPU := resource.MustParse("2")

	app := manifest.Application{}
	defaultContainer, err := app.ToContainer(&v1alpha1.SpaceStatusRuntimeConfig{})
	testutil.AssertNil(t, "default container err", err)

	defaultSpaceStatusNetworkConfig := v1alpha1.SpaceStatusNetworkConfig{
		Domains: []v1alpha1.SpaceDomain{
			{Domain: "example.com"},
		},
	}

	defaultV2Stack := kfconfig.StackV2Definition{
		Name:  "default",
		Image: "some/stack:latest",
	}

	cflinuxfs3Stack := kfconfig.StackV2Definition{
		Name:  "cflinuxfs3",
		Image: "cflinuxfs3:latest",
	}

	v3Stack := kfconfig.StackV3Definition{
		Name: "v3stack",
	}

	defaultSpaceStatusBuildConfig := v1alpha1.SpaceStatusBuildConfig{
		ContainerRegistry: "some-registry",
		StacksV2: kfconfig.StackV2List{
			defaultV2Stack,
			cflinuxfs3Stack,
		},
		StacksV3: kfconfig.StackV3List{
			v3Stack,
		},
	}

	buildpackWithParams := v1alpha1.BuildpackV2Build("some-image", cflinuxfs3Stack, []string{"some-buildpack"}, true)
	buildpackWithoutSourceOption := apps.WithPushBuild(bldPtr(buildpackWithoutSource(v1alpha1.BuildpackV2Build("some-image", defaultV2Stack, nil, false))))
	buildpackOption := apps.WithPushBuild(bldPtr(v1alpha1.BuildpackV2Build("some-image", defaultV2Stack, nil, false)))
	buildpackV3WithoutSourceOption := apps.WithPushBuild(bldPtr(buildpackWithoutSource(v1alpha1.BuildpackV3Build("some-image", v3Stack, nil))))

	cwd, err := os.Getwd()
	testutil.AssertNil(t, "err", err)
	cwdSourcePathOption := apps.WithPushSourcePath(cwd)

	defaultOptions := []apps.PushOption{
		apps.WithPushGenerateDefaultRoute(true),
		apps.WithPushContainer(corev1.Container{
			ReadinessProbe: defaultContainer.ReadinessProbe,
		}),
	}
	manifestBuildSpec := bldPtr(buildpackWithoutSource(v1alpha1.BuildpackV2Build("some-image", defaultV2Stack, []string{"java", "tomcat"}, true)))

	for tn, tc := range map[string]struct {
		args            []string
		namespace       string
		wantErr         error
		pusherErr       error
		srcImageBuilder SrcImageBuilderFunc
		wantImage       string
		targetSpace     *v1alpha1.Space
		wantOpts        []apps.PushOption
		enableAppDevEx  bool
	}{
		"uses configured properties": {
			namespace: "some-namespace",
			args: []string{
				"example-app",
				"--buildpack", "some-buildpack",
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
			wantImage: "some-reg.io/src-some-namespace-example-app:digest",
			srcImageBuilder: func(dir, srcImage string, filter sourceimage.FileFilter) (string, error) {
				testutil.AssertEqual(t, "path", true, strings.Contains(dir, "example-app"))
				testutil.AssertEqual(t, "path is abs", true, filepath.IsAbs(dir))
				testutil.AssertEqual(t, "srcImage", "some-reg.io/src-some-namespace-example-app", srcImage)
				return srcImage + ":digest", nil
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				apps.WithPushBuild(bldPtr(buildpackWithParams)),
				apps.WithPushAppSpecInstances(v1alpha1.AppSpecInstances{
					Stopped:  true,
					Replicas: ptr.Int32(1),
				}),
				apps.WithPushContainer(corev1.Container{
					Args:    []string{"a", "b"},
					Command: []string{"start-web.sh"},
					Env:     envutil.MapToEnvVars(map[string]string{"env1": "val1", "env2": "val2"}),
					Ports:   nil,
					ReadinessProbe: &corev1.Probe{
						TimeoutSeconds:   28,
						SuccessThreshold: 1,
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{},
						},
					},
				}),
			),
		},
		"uses current working directory for empty path": {
			namespace: "some-namespace",
			args: []string{
				"app-name",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				buildpackWithoutSourceOption,
				cwdSourcePathOption,
			),
		},
		"custom-source": {
			namespace: "some-namespace",
			args: []string{
				"app-name",
				"--source-image", "custom-reg.io/source-image:latest",
			},
			wantImage: "custom-reg.io/source-image:latest",
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				buildpackOption,
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
				apps.WithPushSpace("some-namespace"),
				buildpackOption,
				apps.WithPushAppSpecInstances(v1alpha1.AppSpecInstances{Replicas: ptr.Int32(11)}),
			),
		},
		"instances from manifest": {
			namespace: "some-namespace",
			args: []string{
				"instances-app",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				buildpackWithoutSourceOption,
				apps.WithPushAppSpecInstances(v1alpha1.AppSpecInstances{Replicas: ptr.Int32(9)}),
				apps.WithPushSourcePath("testdata"),
			),
		},
		"bind-service-instance": {
			namespace: "some-namespace",
			args: []string{
				"app-name",
				"--manifest", "testdata/manifest-services.yaml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				buildpackWithoutSourceOption,
				apps.WithPushSourcePath("testdata"),
				apps.WithPushServiceBindings([]v1alpha1.ServiceInstanceBinding{
					{
						Spec: v1alpha1.ServiceInstanceBindingSpec{
							BindingType: v1alpha1.BindingType{
								App: &v1alpha1.AppRef{
									Name: "app-name",
								},
							},
							InstanceRef: corev1.LocalObjectReference{
								Name: "some-service-instance",
							},
						},
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
			wantErr:   errors.New("some error"),
			pusherErr: errors.New("some error"),
			wantImage: "some-reg.io/src-default-app-name",
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("default"),
				buildpackOption,
			),
		},
		"space is not provided": {
			args:    []string{"app-name"},
			wantErr: errors.New(config.EmptySpaceError),
		},
		"v2 stack override": {
			namespace: "some-namespace",
			args: []string{
				"app-name",
				"--stack", "default",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				buildpackWithoutSourceOption,
				cwdSourcePathOption,
			),
		},
		"v3 stack override": {
			namespace: "some-namespace",
			args: []string{
				"app-name",
				"--stack", "v3stack",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				buildpackV3WithoutSourceOption,
				cwdSourcePathOption,
			),
		},
		"stack override does not exist": {
			namespace: "some-namespace",
			args: []string{
				"app-name",
				"--stack", "foo",
			},
			wantErr: errors.New("no matching stack \"foo\" found in space \"target-space\""),
		},
		"stack in manifest does not exist": {
			namespace: "some-namespace",
			args: []string{
				"bad-stack-app",
				"--manifest", "testdata/manifest.yml",
			},
			wantErr: errors.New("no matching stack \"foo\" found in space \"target-space\""),
		},
		"v2 stack when AppDevExperienceBuilds is enabled": {
			namespace: "some-namespace",
			args: []string{
				"app-name",
				"--stack", "default",
			},
			enableAppDevEx: true,
			wantErr:        errors.New("no matching stack \"default\" found in space \"target-space\""),
		},
		"SrcImageBuilder returns an error": {
			namespace: "some-namespace",
			args:      []string{"app-name", "--container-registry=some.registry"},
			wantErr:   errors.New("some error"),
			srcImageBuilder: func(dir, srcImage string, filter sourceimage.FileFilter) (string, error) {
				return "", errors.New("some error")
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
				apps.WithPushSpace("some-namespace"),
				apps.WithPushContainerImage(&someImage),
			),
		},
		"container image with AppDevExperience builds": {
			namespace:      "some-namespace",
			enableAppDevEx: true,
			args: []string{
				"app-name",
				"--docker-image", "some-image",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				apps.WithPushContainerImage(&someImage),
				apps.WithPushADXBuild(false),
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
				apps.WithPushSpace("some-namespace"),
				apps.WithPushContainerImage(&someImage),
				apps.WithPushContainer(corev1.Container{
					Env:            envutil.MapToEnvVars(map[string]string{"WHATNOW": "BROWNCOW"}),
					ReadinessProbe: defaultContainer.ReadinessProbe,
				}),
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
		"invalid container registry and appDevExperienceBuilds": {
			namespace: "some-namespace",
			args: []string{
				"buildpack-app",
				"--container-registry", "some-registry",
				"--manifest", "testdata/manifest.yml",
			},
			enableAppDevEx: true,
			wantErr:        errors.New("--container-registry is not valid with AppDevExperienceBuilds"),
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
				apps.WithPushSpace("some-namespace"),
				apps.WithPushContainerImage(&dockerAppImage),
			),
		},
		"docker app from manifest with AppDevExperience builds": {
			namespace: "some-namespace",
			args: []string{
				"docker-app",
				"--manifest", "testdata/manifest.yml",
			},
			enableAppDevEx: true,
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				apps.WithPushContainerImage(&dockerAppImage),
				apps.WithPushADXBuild(false),
			),
		},
		"buildpack app from manifest": {
			namespace: "some-namespace",
			args: []string{
				"buildpack-app",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				apps.WithPushBuild(manifestBuildSpec),
				apps.WithPushSourcePath("testdata/example-app"),
			),
		},
		"buildpack app from manifest with AppDevExperience builds": {
			namespace: "some-namespace",
			args: []string{
				"buildpack-app-with-stack",
				"--manifest", "testdata/manifest.yml",
			},
			enableAppDevEx: true,
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				apps.WithPushSourcePath("testdata/example-app"),
				apps.WithPushADXBuild(true),
				apps.WithPushADXStack(v3Stack),
				apps.WithPushADXContainerRegistry("some-registry"),
			),
		},
		"manifest missing app": {
			namespace: "some-namespace",
			args: []string{
				"missing-app",
				"--manifest", "testdata/manifest.yml",
			},
			wantErr: errors.New(`the manifest doesn't have an App named "missing-app", available names are: ["auto-buildpack-app" "autoscaling-instances-app" "bad-dockerfile-app" "bad-stack-app" "buildpack-app" "buildpack-app-with-stack" "docker-app" "dockerfile-app" "http-health-check-app" "instances-app" "random-route-app" "resources-app" "routes-app" "tcp-health-check-app" "wildcard-routes-app"]`),
		},
		"create and map routes from manifest": {
			namespace: "some-namespace",
			args: []string{
				"routes-app",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				buildpackWithoutSourceOption,
				apps.WithPushRoutes(buildTestRoutes()),
				apps.WithPushGenerateDefaultRoute(false),
				apps.WithPushSourcePath("testdata"),
			),
		},
		"create and map wildcard routes from manifest": {
			namespace: "some-namespace",
			args: []string{
				"wildcard-routes-app",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				buildpackWithoutSourceOption,
				apps.WithPushRoutes([]v1alpha1.RouteWeightBinding{
					buildRoute("*", "example.com", ""),
					buildRoute("*", "host.example.com", "/foo"),
					buildRoute("*", "shost.example.com", "/bar"),
				}),
				apps.WithPushGenerateDefaultRoute(false),
				apps.WithPushSourcePath("testdata"),
			),
		},
		"create and map default routes": {
			namespace: "some-namespace",
			targetSpace: &v1alpha1.Space{
				Status: v1alpha1.SpaceStatus{
					NetworkConfig: v1alpha1.SpaceStatusNetworkConfig{
						Domains: []v1alpha1.SpaceDomain{
							{Domain: "right.example.com"},
							{Domain: "wrong.example.com"},
						},
					},
					BuildConfig: defaultSpaceStatusBuildConfig,
				},
			},
			args: []string{
				"routes-app",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				buildpackWithoutSourceOption,
				apps.WithPushGenerateDefaultRoute(true),
				cwdSourcePathOption,
			),
		},
		"no-route prevents default route": {
			namespace: "some-namespace",
			args: []string{
				"routes-app",
				"--no-route",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				buildpackWithoutSourceOption,
				apps.WithPushRoutes(nil),
				apps.WithPushGenerateDefaultRoute(false),
				cwdSourcePathOption,
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
				apps.WithPushSpace("some-namespace"),
				buildpackWithoutSourceOption,
				apps.WithPushRoutes(nil),
				apps.WithPushGenerateDefaultRoute(false),
				apps.WithPushSourcePath("testdata"),
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
				buildpackWithoutSourceOption,
				apps.WithPushSpace("some-namespace"),
				apps.WithPushRoutes(buildTestRoutes()),
				apps.WithPushGenerateRandomRoute(true),
				apps.WithPushGenerateDefaultRoute(false),
				apps.WithPushSourcePath("testdata"),
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
				apps.WithPushSpace("some-namespace"),
				buildpackWithoutSourceOption,
				apps.WithPushRoutes([]v1alpha1.RouteWeightBinding{
					buildRoute("withscheme", "example.com", "/path1"),
					buildRoute("noscheme", "example.com", ""),
				}),
				apps.WithPushGenerateDefaultRoute(false),
				cwdSourcePathOption,
			),
		},
		"http-health-check from manifest": {
			namespace: "some-namespace",
			args: []string{
				"http-health-check-app",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				apps.WithPushContainer(corev1.Container{
					ReadinessProbe: &corev1.Probe{
						SuccessThreshold: 1,
						TimeoutSeconds:   42,
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{Path: "/healthz"},
						},
					},
				}),
				apps.WithPushContainerImage(&healthCheckApp),
			),
		},
		"tcp-health-check from manifest": {
			namespace: "some-namespace",
			args: []string{
				"tcp-health-check-app",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				apps.WithPushContainerImage(ptr.String("gcr.io/tcp-health-check-app")),
				apps.WithPushContainer(corev1.Container{
					ReadinessProbe: &corev1.Probe{
						SuccessThreshold: 1,
						TimeoutSeconds:   33,
						ProbeHandler: corev1.ProbeHandler{
							TCPSocket: &corev1.TCPSocketAction{},
						},
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
				apps.WithPushSpace("some-namespace"),
				apps.WithPushContainer(corev1.Container{
					ReadinessProbe: defaultContainer.ReadinessProbe,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory:           wantMemory,
							corev1.ResourceEphemeralStorage: wantDiskQuota,
							corev1.ResourceCPU:              wantCPU,
						},
					},
				}),
				buildpackWithoutSourceOption,
				apps.WithPushSourcePath("testdata"),
			),
		},
		"resource requests from flags": {
			namespace: "some-namespace",
			args: []string{
				"resources-app",
				"--disk-quota", "2Gi",
				"--memory-limit", "2Gi",
				"--cpu-cores", "2",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				apps.WithPushContainer(corev1.Container{
					ReadinessProbe: defaultContainer.ReadinessProbe,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory:           wantMemory,
							corev1.ResourceEphemeralStorage: wantDiskQuota,
							corev1.ResourceCPU:              wantCPU,
						},
					},
				}),
				buildpackWithoutSourceOption,
				cwdSourcePathOption,
			),
		},
		"overrides resource requests from flags": {
			namespace: "some-namespace",
			args: []string{
				"resources-app",
				"--disk-quota", "3Gi",
				"--memory-limit", "3Gi",
				"--cpu-cores", "3",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				apps.WithPushContainer(corev1.Container{
					ReadinessProbe: defaultContainer.ReadinessProbe,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory:           resource.MustParse("3Gi"),
							corev1.ResourceEphemeralStorage: resource.MustParse("3Gi"),
							corev1.ResourceCPU:              resource.MustParse("3"),
						},
					},
				}),
				buildpackWithoutSourceOption,
				apps.WithPushSourcePath("testdata"),
			),
		},
		"overrides http health check": {
			namespace: "some-namespace",
			args: []string{
				"http-health-check-app",
				"--timeout", "50",
				"--health-check-http-endpoint", "/test",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				apps.WithPushContainerImage(ptr.String("gcr.io/http-health-check-app")),
				apps.WithPushContainer(corev1.Container{
					ReadinessProbe: &corev1.Probe{
						SuccessThreshold: 1,
						TimeoutSeconds:   50,
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{Path: "/test"},
						},
					},
				}),
				apps.WithPushContainerImage(&healthCheckApp),
			),
		},
		"bad dockerfile": {
			namespace: "some-namespace",
			args: []string{
				"bad-dockerfile-app",
				"--manifest", "testdata/manifest.yml",
				"--container-registry=some.registry",
			},
			wantErr: errors.New(`the Dockerfile "does-not-exist" couldn't be found under the app root`),
		},
		"good dockerfile": {
			namespace: "some-namespace",
			args: []string{
				"dockerfile-app",
				"--manifest", "testdata/manifest.yml",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				apps.WithPushBuild(bldPtr(buildpackWithoutSource(v1alpha1.DockerfileBuild("some-image", "Dockerfile")))),
				apps.WithPushSourcePath("testdata/dockerfile-app"),
			),
		},
		"provided dockerfile": {
			namespace: "some-namespace",
			args: []string{
				"dockerfile-app",
				"--dockerfile", "testdata/dockerfile-app/Dockerfile",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				apps.WithPushBuild(bldPtr(buildpackWithoutSource(v1alpha1.DockerfileBuild("some-image", "testdata/dockerfile-app/Dockerfile")))),
				cwdSourcePathOption,
			),
		},
		"provided dockerfile when AppDevExperienceBuilds is enabled": {
			namespace: "some-namespace",
			args: []string{
				"dockerfile-app",
				"--dockerfile", "testdata/dockerfile-app/Dockerfile",
			},
			enableAppDevEx: true,
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				apps.WithPushADXDockerfile("testdata/dockerfile-app/Dockerfile"),
				apps.WithPushADXBuild(true),
				apps.WithPushADXContainerRegistry("some-registry"),
				cwdSourcePathOption,
			),
		},
		"variable replacement missing vars": {
			namespace: "some-namespace",
			args: []string{
				"app-accounting-prod",
				"--manifest", "testdata/replacement/manifest.yaml",
				"--vars-file", "testdata/replacement/prod.yaml",
				"--vars-file", "testdata/replacement/ops-vars.json",
			},
			wantErr: errors.New("supplied manifest file testdata/replacement/manifest.yaml resulted in error: no variable found for key: ((db_url))"),
		},
		"variable replacement": {
			namespace: "some-namespace",
			args: []string{
				"app-accounting-prod",
				"--manifest", "testdata/replacement/manifest.yaml",
				"--vars-file", "testdata/replacement/prod.yaml",
				"--vars-file", "testdata/replacement/ops-vars.json",
				"--var", "db_url=notset",
				"--var", "db_url=postgresql://127.0.0.1", // later overrides earlier
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				apps.WithPushAppSpecInstances(v1alpha1.AppSpecInstances{Replicas: ptr.Int32(3)}),
				apps.WithPushContainer(corev1.Container{
					Env:            envutil.MapToEnvVars(map[string]string{"DATABASE_URL": "postgresql://127.0.0.1"}),
					ReadinessProbe: defaultContainer.ReadinessProbe,
				}),
				apps.WithPushContainerImage(ptr.String("gcr.io/app-accounting-prod:latest")),
			),
		},
		"task prevents default route": {
			namespace: "some-namespace",
			args: []string{
				"task-app",
				"--task",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				buildpackWithoutSourceOption,
				apps.WithPushRoutes(nil),
				apps.WithPushGenerateDefaultRoute(false),
				apps.WithPushAppSpecInstances(v1alpha1.AppSpecInstances{Stopped: true}),
				cwdSourcePathOption,
			),
		},
		"legacy push still maintains source image": {
			namespace: "some-namespace",
			args: []string{
				"legacy-push",
				"--container-registry=some.registry",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				buildpackOption,
			),
		},
		"build spec params should not have the source image": {
			namespace: "some-namespace",
			args: []string{
				"normal-push",
			},
			wantOpts: append(defaultOptions,
				apps.WithPushSpace("some-namespace"),
				buildpackWithoutSourceOption,
				cwdSourcePathOption,
			),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			if tc.srcImageBuilder == nil {
				tc.srcImageBuilder = func(dir, srcImage string, filter sourceimage.FileFilter) (string, error) {
					return srcImage, nil
				}
			}

			ctrl := gomock.NewController(t)
			fakePusher := appsfake.NewFakePusher(ctrl)

			fakePusher.
				EXPECT().
				Push(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, appName string, opts ...apps.PushOption) error {
					testutil.AssertEqual(t, "app name", tc.args[0], appName)

					expectOpts := apps.PushOptions(tc.wantOpts)
					actualOpts := apps.PushOptions(opts)

					testutil.AssertEqual(t, "space", expectOpts.Space(), actualOpts.Space())
					testutil.AssertEqual(t, "instances", expectOpts.AppSpecInstances(), actualOpts.AppSpecInstances())
					testutil.AssertEqual(t, "routes", expectOpts.Routes(), actualOpts.Routes())
					testutil.AssertEqual(t, "default route", expectOpts.GenerateDefaultRoute(), actualOpts.GenerateDefaultRoute())
					testutil.AssertEqual(t, "random route", expectOpts.GenerateRandomRoute(), actualOpts.GenerateRandomRoute())
					testutil.AssertEqual(t, "container", expectOpts.Container(), actualOpts.Container())
					testutil.AssertEqual(t, "container image", expectOpts.ContainerImage(), actualOpts.ContainerImage())
					testutil.AssertEqual(t, "source path", expectOpts.SourcePath(), actualOpts.SourcePath())
					testutil.AssertEqual(t, "ADX Builds", expectOpts.ADXBuild(), actualOpts.ADXBuild())
					testutil.AssertEqual(t, "ADX Container Registry", expectOpts.ADXContainerRegistry(), actualOpts.ADXContainerRegistry())
					testutil.AssertEqual(t, "ADX Stack", expectOpts.ADXStack(), actualOpts.ADXStack())
					testutil.AssertEqual(t, "ADX Dockerfile", expectOpts.ADXDockerfile(), actualOpts.ADXDockerfile())

					expectedBuild := expectOpts.Build()
					if expectedBuild != nil {
						actualBuild := actualOpts.Build()
						testutil.AssertNotNil(t, "actualBuild", actualBuild)

						testutil.AssertEqual(t, "name", expectedBuild.Name, actualBuild.Name)
						testutil.AssertEqual(t, "kind", expectedBuild.Kind, actualBuild.Kind)

						actualParams := actualBuild.Params
						var editedParams []v1alpha1.BuildParam
						for _, p := range actualParams {
							if p.Name == "SOURCE_IMAGE" {

								if tc.wantImage != "" {
									testutil.AssertEqual(t, "srcImage", tc.wantImage, p.Value)
								}

								p.Value = "some-image"
							}

							editedParams = append(editedParams, p)
						}
						testutil.AssertEqual(t, "buildParams", expectedBuild.Params, editedParams)
						testutil.AssertEqual(t, "env", expectedBuild.Env, actualBuild.Env)
					}

					return tc.pusherErr
				})

			params := &config.KfParams{
				Space:       tc.namespace,
				TargetSpace: tc.targetSpace,
			}

			ctx := injection.WithInjection(context.Background(), t)
			if tc.enableAppDevEx {
				ff := make(kfconfig.FeatureFlagToggles)
				ff.SetAppDevExperienceBuilds(true)
				ctx = testutil.WithFeatureFlags(ctx, t, ff)
			}

			if params.TargetSpace == nil {
				params.TargetSpace = &v1alpha1.Space{}
				params.TargetSpace.Name = "target-space"
				params.TargetSpace.Status.NetworkConfig = defaultSpaceStatusNetworkConfig
				params.TargetSpace.Status.BuildConfig = defaultSpaceStatusBuildConfig
			}

			c := NewPushCommand(params, fakePusher, tc.srcImageBuilder)
			buffer := &bytes.Buffer{}
			c.SetOutput(buffer)
			c.SetArgs(tc.args)
			c.SetContext(ctx)
			_, gotErr := c.ExecuteC()
			t.Log("Command output:", buffer.String())
			if tc.wantErr != nil || gotErr != nil {
				if fmt.Sprint(tc.wantErr) != fmt.Sprint(gotErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.wantErr, gotErr)
				}

				return
			}
			testutil.AssertEqual(t, "SilenceUsage", true, c.SilenceUsage)

		})
	}
}

func TestPushTimeout(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		lookup func(string) (string, bool)
		assert func(*testing.T, time.Duration, error)
	}{
		{
			name: "defaults to 15 minutes",
			lookup: func(name string) (string, bool) {
				return "", false
			},
			assert: func(t *testing.T, d time.Duration, err error) {
				testutil.AssertEqual(t, "duration", 15*time.Minute, d)
				testutil.AssertErrorsEqual(t, nil, err)
			},
		},
		{
			name: "uses CF_STARTUP_TIMEOUT over KF_STARTUP_TIMEOUT",
			lookup: func(name string) (string, bool) {
				switch name {
				case "CF_STARTUP_TIMEOUT":
					return "19s", true
				case "KF_STARTUP_TIMEOUT":
					return "20s", true
				}

				return "", false
			},
			assert: func(t *testing.T, d time.Duration, err error) {
				testutil.AssertEqual(t, "duration", 19*time.Second, d)
				testutil.AssertErrorsEqual(t, nil, err)
			},
		},
		{
			name: "uses KF_STARTUP_TIMEOUT",
			lookup: func(name string) (string, bool) {
				switch name {
				case "KF_STARTUP_TIMEOUT":
					return "20s", true
				}

				return "", false
			},
			assert: func(t *testing.T, d time.Duration, err error) {
				testutil.AssertEqual(t, "duration", 20*time.Second, d)
				testutil.AssertErrorsEqual(t, nil, err)
			},
		},
		{
			name: "invalid CF_STARTUP_TIMEOUT",
			lookup: func(name string) (string, bool) {
				switch name {
				case "CF_STARTUP_TIMEOUT":
					return "invalid", true
				}

				return "", false
			},
			assert: func(t *testing.T, d time.Duration, err error) {
				testutil.AssertErrorsEqual(t, errors.New(`time: invalid duration "invalid"`), err)
			},
		},
		{
			name: "invalid KF_STARTUP_TIMEOUT",
			lookup: func(name string) (string, bool) {
				switch name {
				case "KF_STARTUP_TIMEOUT":
					return "invalid", true
				}

				return "", false
			},
			assert: func(t *testing.T, d time.Duration, err error) {
				testutil.AssertErrorsEqual(t, errors.New(`time: invalid duration "invalid"`), err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d, err := pushTimeout(tc.lookup)
			tc.assert(t, d, err)
		})
	}
}

func buildRoute(hostname, domain, path string) v1alpha1.RouteWeightBinding {
	return v1alpha1.RouteWeightBinding{
		RouteSpecFields: v1alpha1.RouteSpecFields{
			Hostname: hostname,
			Domain:   domain,
			Path:     path,
		},
	}
}

func buildTestRoutes() []v1alpha1.RouteWeightBinding {
	return []v1alpha1.RouteWeightBinding{
		buildRoute("", "example.com", ""),
		buildRoute("", "www.example.com", "/foo"),
		buildRoute("host", "example.com", "/foo"),
	}
}

func buildpackWithoutSource(b v1alpha1.BuildSpec) v1alpha1.BuildSpec {
	// Remove the source param.
	for i, p := range b.Params {
		if p.Name != v1alpha1.SourceImageParamName {
			continue
		}

		b.Params = append(b.Params[:i], b.Params[i+1:]...)
		break
	}

	return b
}
