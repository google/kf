package kf

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/kf"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

//go:generate go run internal/tools/option-builder/option-builder.go options.yml

// Pusher deploys source code to Knative. It should be created via NewPusher.
type Pusher struct {
	f   ServingFactory
	b   SrcImageBuilder
	l   AppLister
	bl  Logs
	out io.Writer
}

// AppLister lists the deployed apps.
type AppLister interface {
	// List lists the deployed apps.
	List(opts ...ListOption) ([]serving.Service, error)
}

// Logs handles build and deploy logs.
type Logs interface {
	// Tail writes the logs for the build and deploy stage to the given out.
	// The method exits once the logs are done streaming.
	Tail(out io.Writer, appName, resourceVersion, namespace string) error
}

// SrcImageBuilder creates and uploads a container image that contains the
// contents of the argument 'dir'.
type SrcImageBuilder func(dir, srcImage string) error

// NewPusher creates a new Pusher.
func NewPusher(l AppLister, f ServingFactory, b SrcImageBuilder, bl Logs) *Pusher {
	return &Pusher{
		l:  l,
		f:  f,
		b:  b,
		bl: bl,
	}
}

// Push deploys an application to Knative. It can be configured via
// Options.
func (p *Pusher) Push(appName string, opts ...PushOption) error {
	cfg, err := p.setupConfig(appName, opts)
	if err != nil {
		return err
	}

	var envs map[string]string
	if len(cfg.EnvironmentVariables) > 0 {
		var err error
		envs, err = p.parseEnvs(cfg.EnvironmentVariables)
		if err != nil {
			return kf.ConfigErr{Reason: err.Error()}
		}
	}

	client, err := p.f()
	if err != nil {
		return err
	}

	d, s, err := p.deployScheme(appName, cfg.Namespace, client)
	if err != nil {
		return err
	}

	var buildSpec *serving.RawExtension
	imageName := cfg.DockerImage

	if imageName == "" {
		// Uploading source code writes to `log.Print`. Prefix this to indicate
		// the corresponding logs are for uploading the source code.
		log.SetPrefix("\033[32m[upload-source-code]\033[0m ")
		srcImage, err := p.uploadSrc(appName, cfg)
		// Remove the prefix.
		log.SetPrefix("")
		if err != nil {
			return err
		}

		buildSpec, imageName, err = p.buildSpec(
			appName,
			srcImage,
			cfg.ContainerRegistry,
			cfg.ServiceAccount,
		)
		if err != nil {
			return err
		}
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
		var result []corev1.EnvVar
		for name, value := range envs {
			result = append(result, corev1.EnvVar{
				Name:  name,
				Value: value,
			})
		}
		s.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Env = result
	}

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

func (p *Pusher) setupConfig(appName string, opts []PushOption) (pushConfig, error) {
	cfg := PushOptionDefaults().Extend(opts).toConfig()

	if cfg.Path == "" && cfg.DockerImage == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return pushConfig{}, err
		}
		cfg.Path = cwd
	}

	if appName == "" {
		return pushConfig{}, kf.ConfigErr{"invalid app name"}
	}
	if (cfg.ContainerRegistry == "" && cfg.DockerImage == "") ||
		(cfg.ContainerRegistry != "" && cfg.DockerImage != "") {
		return pushConfig{}, kf.ConfigErr{"container registry or docker image must be set (not both)"}
	}
	if cfg.Path != "" && cfg.DockerImage != "" {
		return pushConfig{}, kf.ConfigErr{"path flag is not valid with docker image flag"}
	}

	return cfg, nil
}

type deployer func(*serving.Service) (*serving.Service, error)

func (p *Pusher) deployScheme(appName, namespace string, client cserving.ServingV1alpha1Interface) (deployer, *serving.Service, error) {
	apps, err := p.l.List(WithListNamespace(namespace))
	if err != nil {
		return nil, nil, err
	}

	// TODO: use WithListAppName
	// Look to see if an app with the same name exists in this namespace. If
	// so, we want to update intead of create.
	for _, app := range apps {
		if app.Name == appName {
			return client.Services(namespace).Update, &app, nil
		}
	}

	return client.Services(namespace).Create, nil, nil
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
	return fmt.Sprintf("%s%s:%d", prefix, appName, time.Now().UnixNano())
}

const (
	buildAPIVersion = "build.knative.dev/v1alpha1"
)

func (p *Pusher) buildSpec(
	appName string,
	srcImage string,
	containerRegistry string,
	serviceAccount string,
) (*serving.RawExtension, string, error) {
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
		return nil, "", err
	}

	return &serving.RawExtension{
		Raw: buildSpecRaw,
	}, imageName, nil
}

func (p *Pusher) initService(appName, namespace string, build *serving.RawExtension) *serving.Service {
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

func (p *Pusher) initMeta(s *serving.Service, appName, namespace string) {
	s.Name = appName
	s.Kind = "Service"
	s.APIVersion = "serving.knative.dev/v1alpha1"
	s.Namespace = namespace
}

func (p *Pusher) buildAndDeploy(
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

// parseEnvs turns a slice of strings formatted as NAME=VALUE into a map. The
// logic is taken from os/exec.dedupEnvCase with a few differences:
// malformed strings create an error, and case insensitivity is always assumed
// false.
func (p *Pusher) parseEnvs(envs []string) (map[string]string, error) {
	m := map[string]string{}
	for _, kv := range envs {
		eq := strings.Index(kv, "=")
		if eq < 0 {
			return nil, fmt.Errorf("malformed environment variable: %s", kv)
		}
		k := kv[:eq]
		v := kv[eq+1:]
		m[k] = v
	}
	return m, nil
}

func (p *Pusher) dedupEnvs(m map[string]string, envs []corev1.EnvVar) []corev1.EnvVar {
	mEnvs := map[string]string{}

	for _, env := range envs {
		mEnvs[env.Name] = env.Value
	}
	for name, value := range m {
		mEnvs[name] = value
	}

	var result []corev1.EnvVar
	for name, value := range mEnvs {
		result = append(result, corev1.EnvVar{
			Name:  name,
			Value: value,
		})
	}
	return result
}
