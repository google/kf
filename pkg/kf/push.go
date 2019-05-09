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
	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

//go:generate go run internal/tools/option-builder/option-builder.go options.yml

// pusher deploys source code to Knative. It should be created via NewPusher.
type pusher struct {
	c   cserving.ServingV1alpha1Interface
	l   AppLister
	bl  Logs
	out io.Writer
}

// AppLister lists the deployed apps.
type AppLister interface {
	// List lists the deployed apps.
	List(opts ...ListOption) ([]serving.Service, error)

	// ListConfigurations the deployed configurations in a namespace.
	ListConfigurations(...ListConfigurationsOption) ([]serving.Configuration, error)
}

// Logs handles build and deploy logs.
type Logs interface {
	// Tail writes the logs for the build and deploy stage to the given out.
	// The method exits once the logs are done streaming.
	Tail(out io.Writer, appName, resourceVersion, namespace string) error
}

// Pusher deploys applications.
type Pusher interface {
	// Push deploys an application.
	Push(appName, srcImageName string, opts ...PushOption) error
}

// NewPusher creates a new Pusher.
func NewPusher(l AppLister, c cserving.ServingV1alpha1Interface, bl Logs) Pusher {
	return &pusher{
		l:  l,
		c:  c,
		bl: bl,
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
		envs, err = envutil.ParseCliEnvVars(cfg.EnvironmentVariables)
		if err != nil {
			return kf.ConfigErr{Reason: err.Error()}
		}
	}

	d, s, err := p.deployScheme(appName, cfg.Namespace)
	if err != nil {
		return err
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

	if s == nil {
		s = p.initService(appName, cfg.Namespace, buildSpec)
	}
	s.Spec.RunLatest.Configuration.Build = buildSpec
	s.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Image = imageName
	s.Spec.RunLatest.Configuration.RevisionTemplate.Spec.ServiceAccountName = cfg.ServiceAccount

	if cfg.Grpc {
		s.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Ports = []corev1.ContainerPort{{Name: "h2c", ContainerPort: 8080}}
	}

	if len(envs) > 0 {
		s.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Env = envs
	}

	// TODO(josephlewis42): inject VCAP style environment variables here.

	resourceVersion, err := p.buildAndDeploy(
		appName,
		cfg.Namespace,
		d,
		s,
	)
	if err != nil {
		return err
	}

	if err := p.bl.Tail(cfg.Output, appName, resourceVersion, cfg.Namespace); err != nil {
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

type deployer func(*serving.Service) (*serving.Service, error)

func (p *pusher) deployScheme(appName, namespace string) (deployer, *serving.Service, error) {
	apps, err := p.l.List(WithListNamespace(namespace))
	if err != nil {
		return nil, nil, err
	}

	// TODO(poy): use WithListAppName
	// Look to see if an app with the same name exists in this namespace. If
	// so, we want to update intead of create.
	for _, app := range apps {
		if app.Name == appName {
			return p.c.Services(namespace).Update, &app, nil
		}
	}

	return p.c.Services(namespace).Create, nil, nil
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
		build.TemplateInstantiationSpec{Name: "buildpack"},
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

func (p *pusher) initService(appName, namespace string, build *serving.RawExtension) *serving.Service {
	s := &serving.Service{
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

	p.initMeta(s, appName, namespace)

	return s
}

func (p *pusher) initMeta(s *serving.Service, appName, namespace string) {
	s.Name = appName
	s.Kind = "Service"
	s.APIVersion = "serving.knative.dev/v1alpha1"
	s.Namespace = namespace
}

func (p *pusher) buildAndDeploy(
	appName string,
	namespace string,
	d deployer,
	s *serving.Service,
) (string, error) {
	p.initMeta(s, appName, namespace)
	s, err := d(s)
	if err != nil {
		return "", err
	}

	return s.ResourceVersion, nil
}
