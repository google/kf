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

package kf

import (
	"fmt"
	"time"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/builds"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/envutil"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/kf"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate go run internal/tools/option-builder/option-builder.go options.yml options.go

// pusher deploys source code to Knative. It should be created via NewPusher.
type pusher struct {
	deployer    Deployer
	bl          Logs
	buildClient builds.ClientInterface
}

// Pusher deploys applications.
type Pusher interface {
	// Push deploys an application.
	Push(appName, srcImageName string, opts ...PushOption) error
}

// NewPusher creates a new Pusher.
func NewPusher(d Deployer, bl Logs, buildClient builds.ClientInterface) Pusher {
	return &pusher{
		deployer:    d,
		bl:          bl,
		buildClient: buildClient,
	}
}

// Push deploys an application to Knative. It can be configured via
// Options.
func (p *pusher) Push(appName, srcImage string, opts ...PushOption) error {
	cfg, err := p.setupConfig(appName, srcImage, opts)
	if err != nil {
		return err
	}

	var envs []corev1.EnvVar
	if len(cfg.EnvironmentVariables) > 0 {
		var err error
		envs = envutil.MapToEnvVars(cfg.EnvironmentVariables)
		if err != nil {
			return kf.ConfigErr{Reason: err.Error()}
		}
	}

	imageName, err := p.buildSpec(
		cfg.Namespace,
		appName,
		srcImage,
		cfg.ContainerRegistry,
		cfg.ServiceAccount,
		cfg.Buildpack,
	)
	if err != nil {
		return err
	}

	s := p.initService(appName, cfg.Namespace)
	s.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Image = imageName
	s.Spec.RunLatest.Configuration.RevisionTemplate.Spec.ServiceAccountName = cfg.ServiceAccount

	if cfg.Grpc {
		s.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Ports = []corev1.ContainerPort{{Name: "h2c", ContainerPort: 8080}}
	}

	if len(envs) > 0 {
		s.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Env = envs
	}

	resultingService, err := p.deployer.Deploy(s, WithDeployNamespace(cfg.Namespace))
	if err != nil {
		return fmt.Errorf("failed to deploy: %s", err)
	}

	if err := p.bl.DeployLogs(cfg.Output, appName, resultingService.ResourceVersion, cfg.Namespace); err != nil {
		return err
	}

	fmt.Fprintf(cfg.Output, "%q successfully deployed\n", appName)
	return nil
}

func (p *pusher) setupConfig(appName, srcImage string, opts []PushOption) (pushConfig, error) {
	cfg := PushOptionDefaults().Extend(opts).toConfig()

	if appName == "" {
		return pushConfig{}, kf.ConfigErr{Reason: "invalid app name"}
	}
	if srcImage == "" {
		return pushConfig{}, kf.ConfigErr{Reason: "invalid source image"}
	}

	return cfg, nil
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

func (p *pusher) buildSpec(
	namespace string,
	appName string,
	srcImage string,
	containerRegistry string,
	serviceAccount string,
	buildpack string,
) (string, error) {
	appImageName := AppImageName(namespace, appName)
	imageDestination := JoinRepositoryImage(containerRegistry, appImageName)

	args := map[string]string{
		"IMAGE": imageDestination,
	}

	if buildpack != "" {
		args["BUILDPACK"] = buildpack
	}

	buildName := BuildName()

	if _, err := p.buildClient.Create(
		buildName,
		builds.BuildpackTemplate(),
		builds.WithCreateServiceAccount(serviceAccount),
		builds.WithCreateArgs(args),
		builds.WithCreateSourceImage(srcImage),
		builds.WithCreateNamespace(namespace),
	); err != nil {
		return "", err
	}

	if err := p.buildClient.Tail(buildName, builds.WithTailNamespace(namespace)); err != nil {
		return "", err
	}

	if _, err := p.buildClient.Status(buildName); err != nil {
		return "", err
	}

	return imageDestination, nil
}

func (p *pusher) initService(appName, namespace string) serving.Service {
	s := serving.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: namespace,
		},
		Spec: serving.ServiceSpec{
			RunLatest: &serving.RunLatestType{
				Configuration: serving.ConfigurationSpec{
					RevisionTemplate: serving.RevisionTemplateSpec{
						Spec: serving.RevisionSpec{
							Container: corev1.Container{
								ImagePullPolicy: "Always",
							},
						},
					},
				},
			},
		},
	}

	return s
}
