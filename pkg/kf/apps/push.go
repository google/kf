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
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/internal/envutil"
	"github.com/google/kf/pkg/kf/sources"
	corev1 "k8s.io/api/core/v1"
)

//go:generate go run ../internal/tools/option-builder/option-builder.go push-options.yml push_options.go

// pusher deploys source code to Knative. It should be created via NewPusher.
type pusher struct {
	appsClient Client
}

// Pusher deploys applications.
type Pusher interface {
	// Push deploys an application.
	Push(appName string, opts ...PushOption) error
}

// NewPusher creates a new Pusher.
func NewPusher(appsClient Client) Pusher {
	return &pusher{
		appsClient: appsClient,
	}
}

func newApp(appName string, opts ...PushOption) (*v1alpha1.App, error) {

	cfg := PushOptionDefaults().Extend(opts).toConfig()

	var envs []corev1.EnvVar
	if len(cfg.EnvironmentVariables) > 0 {
		envs = envutil.MapToEnvVars(cfg.EnvironmentVariables)
	}

	src := sources.NewKfSource()
	switch {
	case cfg.ContainerImage != "":
		src.SetContainerImageSource(cfg.ContainerImage)

	case cfg.DockerfilePath != "":
		src.SetDockerfilePath(cfg.DockerfilePath)
		src.SetDockerfileSource(cfg.SourceImage)

	default: // default to buildpack build
		src.SetBuildpackBuildEnv(envs)
		src.SetBuildpackBuildBuildpack(cfg.Buildpack)
		src.SetBuildpackBuildSource(cfg.SourceImage)
		src.SetBuildpackBuildStack(cfg.Stack)
	}

	app := NewKfApp()
	app.SetName(appName)
	app.SetNamespace(cfg.Namespace)
	app.SetSource(src)
	app.SetResourceRequests(cfg.ResourceRequests)
	app.Spec.Instances = cfg.AppSpecInstances
	app.SetHealthCheck(cfg.HealthCheck)
	app.Spec.Routes = cfg.Routes
	app.Spec.ServiceBindings = cfg.ServiceBindings
	app.SetCommand(cfg.Command)
	app.SetArgs(cfg.Args)

	if cfg.Grpc {
		app.SetContainerPorts([]corev1.ContainerPort{{Name: "h2c", ContainerPort: 8080}})
	}

	if len(envs) > 0 {
		app.SetEnvVars(envs)
	}

	return app.ToApp(), nil
}

// Push deploys an application to Knative. It can be configured via
// Optionapp.
func (p *pusher) Push(appName string, opts ...PushOption) error {
	cfg := PushOptionDefaults().Extend(opts).toConfig()

	app, err := newApp(appName, opts...)
	if err != nil {
		return fmt.Errorf("failed to create app: %s", err)
	}

	var hasDefaultRoutes bool
	app.Spec.Routes, hasDefaultRoutes = setupRoutes(cfg, app.Name, app.Spec.Routes)

	// Scaling
	if noScaling(app.Spec.Instances) {
		// Default to 1
		singleInstance := 1
		app.Spec.Instances.Exactly = &singleInstance
	}

	resultingApp, err := p.appsClient.Upsert(
		app.Namespace,
		app,
		mergeApps(cfg, hasDefaultRoutes),
	)
	if err != nil {
		return fmt.Errorf("failed to push app: %s", err)
	}

	if err := p.appsClient.DeployLogsForApp(cfg.Output, resultingApp); err != nil {
		return err
	}

	status := "deployed"
	if resultingApp.Spec.Instances.Stopped {
		status = "deployed without starting"
	}

	_, err = fmt.Fprintf(cfg.Output, "%q successfully %s\n", appName, status)
	return err
}

func setupRoutes(cfg pushConfig, appName string, r []v1alpha1.RouteSpecFields) (routes []v1alpha1.RouteSpecFields, hasDefaultRoutes bool) {
	switch {
	case len(r) != 0:
		// Don't overwrite the routes
		return r, false
	case cfg.DefaultRouteDomain != "":
		return []v1alpha1.RouteSpecFields{
			{
				Domain:   cfg.DefaultRouteDomain,
				Hostname: appName,
			},
		}, true
	case cfg.RandomRouteDomain != "":
		return []v1alpha1.RouteSpecFields{
			{
				Domain: cfg.RandomRouteDomain,
				Hostname: strings.Join([]string{
					appName,
					strconv.FormatUint(rand.Uint64(), 36),
					strconv.FormatUint(uint64(time.Now().UnixNano()), 36),
				}, "-"),
			},
		}, true
	default:
		return nil, false
	}
}

func noScaling(instances v1alpha1.AppSpecInstances) bool {
	return instances.Exactly == nil &&
		instances.Min == nil &&
		instances.Max == nil
}

func mergeApps(cfg pushConfig, hasDefaultRoutes bool) func(newapp, oldapp *v1alpha1.App) *v1alpha1.App {
	return func(newapp, oldapp *v1alpha1.App) *v1alpha1.App {

		if len(oldapp.Spec.Routes) > 0 && hasDefaultRoutes {
			newapp.Spec.Routes = oldapp.Spec.Routes
		}

		// Scaling overrides
		if noScaling(cfg.AppSpecInstances) {
			// Looks like the user did not set a new value, use the old one
			newapp.Spec.Instances.Exactly = oldapp.Spec.Instances.Exactly
			newapp.Spec.Instances.Min = oldapp.Spec.Instances.Min
			newapp.Spec.Instances.Max = oldapp.Spec.Instances.Max
		}

		// Default scaling
		if noScaling(cfg.AppSpecInstances) && noScaling(oldapp.Spec.Instances) {
			// No scaling in old or new, go with a default of 1. This is to
			// match expectaions for CF users. See
			// https://github.com/google/kf/issues/8 for more context.
			singleInstance := 1
			newapp.Spec.Instances.Exactly = &singleInstance
		}

		newapp.ResourceVersion = oldapp.ResourceVersion
		newEnvs := envutil.GetAppEnvVars(newapp)
		oldEnvs := envutil.GetAppEnvVars(oldapp)
		envutil.SetAppEnvVars(newapp, envutil.DeduplicateEnvVars(append(oldEnvs, newEnvs...)))

		return newapp
	}
}

// SourceImageName gets the image name for source code for an application.
func SourceImageName(namespace, appName string) string {
	return fmt.Sprintf("src-%s-%s:%d", namespace, appName, time.Now().UnixNano())
}

// JoinRepositoryImage joins a repository and image name.
func JoinRepositoryImage(repository, imageName string) string {
	return fmt.Sprintf("%s/%s", repository, imageName)
}
