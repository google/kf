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
	"time"

	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/internal/envutil"
	"github.com/google/kf/pkg/kf/internal/kf"
	"github.com/google/kf/pkg/kf/routes"
	"github.com/google/kf/pkg/kf/sources"
	corev1 "k8s.io/api/core/v1"
)

//go:generate go run ../internal/tools/option-builder/option-builder.go push-options.yml push_options.go

// pusher deploys source code to Knative. It should be created via NewPusher.
type pusher struct {
	appsClient   Client
	routesClient routes.Client
}

// Pusher deploys applications.
type Pusher interface {
	// Push deploys an application.
	Push(appName string, opts ...PushOption) error
}

// NewPusher creates a new Pusher.
func NewPusher(appsClient Client, routesClient routes.Client) Pusher {
	return &pusher{
		appsClient:   appsClient,
		routesClient: routesClient,
	}
}

func newApp(appName string, opts ...PushOption) (*v1alpha1.App, error) {

	cfg := PushOptionDefaults().Extend(opts).toConfig()

	var envs []corev1.EnvVar
	if len(cfg.EnvironmentVariables) > 0 {
		var err error
		envs = envutil.MapToEnvVars(cfg.EnvironmentVariables)
		if err != nil {
			return nil, kf.ConfigErr{Reason: err.Error()}
		}
	}

	src := sources.NewKfSource()
	src.SetBuildpackBuildSource(cfg.SourceImage)
	src.SetContainerImageSource(cfg.ContainerImage)
	src.SetBuildpackBuildRegistry(cfg.ContainerRegistry)
	src.SetBuildpackBuildEnv(envs)
	src.SetBuildpackBuildBuildpack(cfg.Buildpack)

	app := NewKfApp()
	app.SetName(appName)
	app.SetNamespace(cfg.Namespace)
	app.SetServiceAccount(cfg.ServiceAccount)
	app.SetSource(src)

	fmt.Println("about to set mem")
	app.SetMemory(cfg.Memory)
	app.SetStorage(cfg.DiskQuota)
	app.SetCPU(cfg.CPU)
	app.Spec.Instances.Stopped = cfg.NoStart

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
	fmt.Println("pushing app")
	cfg := PushOptionDefaults().Extend(opts).toConfig()

	app, err := newApp(appName, opts...)
	if err != nil {
		return fmt.Errorf("failed to create app: %s", err)
	}

	resultingApp, err := p.appsClient.Upsert(app.Namespace, app, mergeApps)
	if err != nil {
		return fmt.Errorf("failed to push app: %s", err)
	}

	merger := routes.Merger(func(newR, oldR *v1alpha1.Route) *v1alpha1.Route {
		newR.ObjectMeta = *oldR.ObjectMeta.DeepCopy()
		newR.Spec.KnativeServiceNames = append(oldR.Spec.KnativeServiceNames, appName)
		return newR
	})

	routes := cfg.Routes
	for _, route := range routes {
		_, err := p.routesClient.Upsert(cfg.Namespace, route, merger)
		if err != nil {
			return fmt.Errorf("failed to add route: %s", err)
		}
	}

	if err := p.appsClient.DeployLogs(
		cfg.Output,
		appName,
		resultingApp.ResourceVersion,
		app.Namespace,
		cfg.NoStart,
	); err != nil {
		return err
	}

	status := "deployed"
	if cfg.NoStart {
		status = "deployed without starting"
	}

	_, err = fmt.Fprintf(cfg.Output, "%q successfully %s\n", appName, status)
	if err != nil {
		return err
	}
	return nil
}

func mergeApps(newapp, oldapp *v1alpha1.App) *v1alpha1.App {
	newapp.ResourceVersion = oldapp.ResourceVersion
	newEnvs := envutil.GetAppEnvVars(newapp)
	oldEnvs := envutil.GetAppEnvVars(oldapp)
	envutil.SetAppEnvVars(newapp, envutil.DeduplicateEnvVars(append(oldEnvs, newEnvs...)))
	return newapp
}

// AppImageName gets the image name for an application.
func AppImageName(namespace, appName string) string {
	return fmt.Sprintf("app-%s-%s:%d", namespace, appName, time.Now().UnixNano())
}

// SourceImageName gets the image name for source code for an application.
func SourceImageName(namespace, appName string) string {
	return fmt.Sprintf("src-%s-%s:%d", namespace, appName, time.Now().UnixNano())
}

// JoinRepositoryImage joins a repository and image name.
func JoinRepositoryImage(repository, imageName string) string {
	return fmt.Sprintf("%s/%s", repository, imageName)
}

// BuildName gets a build name based on the current time.
// Build names are limited by Knative to be 64 characters long.
func BuildName() string {
	return fmt.Sprintf("build-%d", time.Now().UnixNano())
}
