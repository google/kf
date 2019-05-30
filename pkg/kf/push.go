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
	"encoding/json"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/builds"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/envutil"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/kf"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate go run internal/tools/option-builder/option-builder.go options.yml options.go

// pusher deploys source code to Knative. It should be created via NewPusher.
type pusher struct {
	deployer Deployer
	bl       Logs
	out      io.Writer
}

// Pusher deploys applications.
type Pusher interface {
	// Push deploys an application.
	Push(appName, srcImageName string, opts ...PushOption) error
}

// NewPusher creates a new Pusher.
func NewPusher(d Deployer, bl Logs) Pusher {
	return &pusher{
		deployer: d,
		bl:       bl,
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

	buildSpec, imageName, err := p.buildSpec(
		appName,
		srcImage,
		cfg.ContainerRegistry,
		cfg.ServiceAccount,
		cfg.Buildpack,
	)
	if err != nil {
		return err
	}

	s := p.initService(appName, cfg.Namespace, buildSpec)
	s.Spec.RunLatest.Configuration.Build = buildSpec
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
		return pushConfig{}, kf.ConfigErr{"invalid app name"}
	}
	if srcImage == "" {
		return pushConfig{}, kf.ConfigErr{"invalid source image"}
	}

	return cfg, nil
}

func (p *pusher) imageName(appName string, srcCodeImage bool) string {
	var prefix string
	if srcCodeImage {
		prefix = "src-"
	}
	return fmt.Sprintf("%s%s:%d", prefix, appName, time.Now().UnixNano())
}

const (
	buildAPIVersion = "build.knative.dev/v1alpha1"
)

func (p *pusher) buildSpec(
	appName string,
	srcImage string,
	containerRegistry string,
	serviceAccount string,
	buildpack string,
) (*serving.RawExtension, string, error) {
	imageName := path.Join(containerRegistry, p.imageName(appName, false))

	args := map[string]string{
		"IMAGE": imageName,
	}

	if buildpack != "" {
		args["BUILDPACK"] = buildpack
	}

	// Knative Build wants a Build, but the RawExtension (used by the
	// Configuration object) wants a BuildSpec. Therefore, we have to manually
	// create the required JSON.
	buildSpec := builds.PopulateTemplate(
		"", // no name provided
		build.TemplateInstantiationSpec{Name: "buildpack", Kind: build.ClusterBuildTemplateKind},
		builds.WithCreateServiceAccount(serviceAccount),
		builds.WithCreateArgs(args),
		builds.WithCreateSourceImage(srcImage),
		builds.WithCreateNamespace(""), // set blank namespace so Knative can choose
	)

	buildSpecRaw, err := json.Marshal(buildSpec)
	if err != nil {
		return nil, "", err
	}

	return &serving.RawExtension{
		Raw: buildSpecRaw,
	}, imageName, nil
}

func (p *pusher) initService(appName, namespace string, build *serving.RawExtension) serving.Service {
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
					Build: build,
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
