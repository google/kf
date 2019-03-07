package kf

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	v1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

//go:generate go run internal/tools/option-builder/option-builder.go

// Pusher deploys source code to Knative. It should be created via NewPusher.
type Pusher struct {
	f ServingFactory
	b SrcImageBuilder
	l AppLister
}

// AppLister lists the deployed apps.
type AppLister interface {
	// List lists the deployed apps.
	List(opts ...ListOption) ([]serving.Service, error)
}

// SrcImageBuilder creates and uploads a container image that contains the
// contents of the argument 'dir'.
type SrcImageBuilder func(dir, srcImage string) error

// NewPusher creates a new Pusher.
func NewPusher(l AppLister, f ServingFactory, b SrcImageBuilder) *Pusher {
	return &Pusher{
		l: l,
		f: f,
		b: b,
	}
}

// Push deploys an application to Knative. It can be configured via
// Options.
func (p *Pusher) Push(appName string, opts ...PushOption) error {
	cfg := PushOptions(opts).toConfig()

	if cfg.Path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		cfg.Path = cwd
	}
	if cfg.Namespace == ""{
		cfg.Namespace="default"
	}

	if appName == "" {
		return errors.New("invalid app name")
	}
	if cfg.ContainerRegistry == "" {
		return errors.New("container registry is not set")
	}
	if cfg.ServiceAccount == "" {
		return errors.New("service account is not set")
	}

	client, err := p.f()
	if err != nil {
		return err
	}

	srcImage, err := p.uploadSrc(appName, cfg)
	if err != nil {
		return err
	}

	d, err := p.deployScheme(appName, cfg.Namespace, client)
	if err != nil {
		return err
	}

	return p.buildAndDeploy(
		appName,
		srcImage,
		cfg.Namespace,
		cfg.ContainerRegistry,
		cfg.ServiceAccount,
		d,
	)
}

type deployer func(*v1alpha1.Service) (*v1alpha1.Service, error)

func (p *Pusher) deployScheme(appName, namespace string, client cserving.ServingV1alpha1Interface) (deployer, error) {
	apps, err := p.l.List(WithListNamespace(namespace))
	if err != nil {
		return nil, err
	}

	// Look to see if an app with the same name exists in this namespace. If
	// so, we want to update intead of create.
	for _, app := range apps {
		if app.Name == appName {
			return func(cfg *v1alpha1.Service) (*v1alpha1.Service, error) {
				// Modify the ResourceVersion. This is required for updates.
				cfg.ResourceVersion = app.ResourceVersion
				return client.Services(namespace).Update(cfg)
			}, nil
		}
	}

	return client.Services(namespace).Create, nil
}

func (p *Pusher) uploadSrc(appName string, cfg pushConfig) (string, error) {
	srcImage := path.Join(
		cfg.ContainerRegistry,
		p.imageName(appName, true),
	)
	if err := p.b(cfg.Path, srcImage); err != nil {
		return "", err
	}

	return srcImage, nil
}

func (p *Pusher) imageName(appName string, srcCodeImage bool) string {
	var prefix string
	if srcCodeImage {
		prefix = "src-"
	}
	return fmt.Sprintf("%s%s-%d:latest", prefix, appName, time.Now().UnixNano())
}

const (
	buildAPIVersion = "build.knative.dev/v1alpha1"
)

func (p *Pusher) buildAndDeploy(
	appName string,
	srcImage string,
	namespace string,
	containerRegistry string,
	serviceAccount string,
	d deployer,
) error {
	imageName := path.Join(
		containerRegistry,
		p.imageName(appName, false),
	)

	// Knative Build wants a Build, but the RawExtension (used by the
	// Configuration object) wants a BuildSpec. Therefore, we have to manually
	// create the required JSON.
	buildSpec := build.Build{
		Spec: build.BuildSpec{
			ServiceAccountName: serviceAccount,
			Source: &build.SourceSpec{
				Custom: &corev1.Container{
					Image: srcImage,
				},
			},
			Template: &build.TemplateInstantiationSpec{
				Name: "buildpack",
				Arguments: []build.ArgumentSpec{
					{
						Name:  "IMAGE",
						Value: imageName,
					},
				},
			},
		},
	}
	buildSpec.Kind = "Build"
	buildSpec.APIVersion = buildAPIVersion
	buildSpecRaw, err := json.Marshal(buildSpec)
	if err != nil {
		return err
	}

	cfg := &serving.Service{
		Spec: serving.ServiceSpec{
			RunLatest: &serving.RunLatestType{
				Configuration: serving.ConfigurationSpec{
					Build: &serving.RawExtension{
						Raw: buildSpecRaw,
					},

					RevisionTemplate: serving.RevisionTemplateSpec{
						Spec: serving.RevisionSpec{
							Container: corev1.Container{
								Image:           imageName,
								ImagePullPolicy: "Always",
							},
						},
					},
				},
			},
		},
	}
	cfg.Name = appName
	cfg.Kind = "Service"
	cfg.APIVersion = "serving.knative.dev/v1alpha1"
	cfg.Namespace = namespace

	if _, err := d(cfg); err != nil {
		return err
	}

	return nil
}
