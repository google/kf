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
	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// Pusher deploys source code to Knative. It should be created via NewPusher.
type Pusher struct {
	f ServingFactory
	b SrcImageBuilder
}

// SrcImageBuilder creates and uploads a container image that contains the
// contents of the argument 'dir'.
type SrcImageBuilder func(dir, srcImage string) error

// NewPusher creates a new Pusher.
func NewPusher(f ServingFactory, b SrcImageBuilder) *Pusher {
	return &Pusher{
		f: f,
		b: b,
	}
}

// PushOption overrides the defaults when pushing.
type PushOption interface {
	configure(*pushConfig)
}

type pushConfig struct {
	namespace         string
	path              string
	containerRegistry string
	serviceAccount    string
}

//PushOptionFunc The HandlerFunc type is an adapter to allow the use of
//ordinary functions as a PushOption.
type PushOptionFunc func(*pushConfig)

func (f PushOptionFunc) configure(c *pushConfig) {
	f(c)
}

// PushOptions is a slice of PushOption. It has methods that are helpful when
// analyzing given PushOptions.
type PushOptions []PushOption

// Namespace returns the set namespace. Will return empty if the it's not set.
func (o PushOptions) Namespace() string {
	cfg := pushConfig{}
	for _, opt := range o {
		opt.configure(&cfg)
	}
	return cfg.namespace
}

// Path returns the set path . Will return empty if the it's not set.
func (o PushOptions) Path() string {
	cfg := pushConfig{}
	for _, opt := range o {
		opt.configure(&cfg)
	}
	return cfg.path
}

// ContainerRegistry returns the set container registry. Will return empty if
// the it's not set.
func (o PushOptions) ContainerRegistry() string {
	cfg := pushConfig{}
	for _, opt := range o {
		opt.configure(&cfg)
	}
	return cfg.containerRegistry
}

// ServiceAccount returns the set service account. Will return empty if
// the it's not set.
func (o PushOptions) ServiceAccount() string {
	cfg := pushConfig{}
	for _, opt := range o {
		opt.configure(&cfg)
	}
	return cfg.serviceAccount
}

// WithPushNamespace configures the namespace. Defaults to "default".
func WithPushNamespace(namespace string) PushOption {
	return PushOptionFunc(func(c *pushConfig) {
		c.namespace = namespace
	})
}

// WithPushPath configures the path to find the source code. Defaults to
// current working directory.
func WithPushPath(path string) PushOption {
	return PushOptionFunc(func(c *pushConfig) {
		c.path = path
	})
}

// WithPushContainerRegistry configures the container registry to use while
// pushing. This is currently required.
func WithPushContainerRegistry(containerRegistry string) PushOption {
	return PushOptionFunc(func(c *pushConfig) {
		c.containerRegistry = containerRegistry
	})
}

// WithPushServiceAccount configures the container registry to use while
// pushing. This is currently required.
func WithPushServiceAccount(serviceAccount string) PushOption {
	return PushOptionFunc(func(c *pushConfig) {
		c.serviceAccount = serviceAccount
	})
}

// Push deploys an application to Knative. It can be configured via
// PushOptions.
func (p *Pusher) Push(appName string, opts ...PushOption) error {
	cfg := pushConfig{
		namespace: "default",
	}

	for _, o := range opts {
		o.configure(&cfg)
	}

	if cfg.path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		cfg.path = cwd
	}

	if appName == "" {
		return errors.New("invalid app name")
	}
	if cfg.containerRegistry == "" {
		return errors.New("container registry is not set")
	}
	if cfg.serviceAccount == "" {
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

	return p.buildAndDeploy(
		appName,
		srcImage,
		cfg.namespace,
		cfg.containerRegistry,
		cfg.serviceAccount,
		client,
	)
}

func (p *Pusher) uploadSrc(appName string, cfg pushConfig) (string, error) {
	srcImage := path.Join(
		cfg.containerRegistry,
		p.imageName(appName, true),
	)
	if err := p.b(cfg.path, srcImage); err != nil {
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
	client cserving.ServingV1alpha1Interface,
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

	if _, err = client.Services(namespace).Create(cfg); err != nil {
		return err
	}

	return nil
}
