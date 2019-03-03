package commands

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
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

const (
	buildAPIVersion = "build.knative.dev/v1alpha1"
)

// SrcImageBuilder creates and uploads a container image that contains the
// contents of the argument 'dir'.
type SrcImageBuilder func(dir, srcImage string) error

// NewPushCommand creates a push command.
func NewPushCommand(p *KfParams, builder SrcImageBuilder) *cobra.Command {
	var (
		containerRegistry string
		serviceAccount    string
	)

	var pushCmd = &cobra.Command{
		Use:   "push",
		Short: "Push a new app or sync changes to an existing app",
		Args:  cobra.ExactArgs(1),
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Cobra ensures we are only called with a single argument.
			appName := args[0]
			if containerRegistry == "" {
				return errors.New("container registry is not set")
			}
			if serviceAccount == "" {
				return errors.New("service account is not set")
			}

			cmd.SilenceUsage = true
			client, err := p.ServingFactory()
			if err != nil {
				return err
			}

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			srcImage := path.Join(
				containerRegistry,
				imageName(appName, true),
			)
			if err := builder(cwd, srcImage); err != nil {
				return err
			}

			return buildAndDeploy(
				appName,
				srcImage,
				containerRegistry,
				serviceAccount,
				p,
				client,
			)
		},
	}

	pushCmd.Flags().StringVar(
		&containerRegistry,
		"container-registry",
		"",
		"The container registry to push containers (REQUIRED)",
	)

	pushCmd.Flags().StringVar(
		&serviceAccount,
		"service-account",
		"",
		"The service account to enable access to the container registry (REQUIRED)",
	)

	return pushCmd
}

func imageName(appName string, srcCodeImage bool) string {
	var prefix string
	if srcCodeImage {
		prefix = "src-"
	}
	return fmt.Sprintf("%s%s-%d:latest", prefix, appName, time.Now().UnixNano())
}

func buildAndDeploy(
	appName string,
	srcImage string,
	containerRegistry string,
	serviceAccount string,
	p *KfParams,
	client cserving.ServingV1alpha1Interface,
) error {
	imageName := path.Join(
		containerRegistry,
		imageName(appName, false),
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
	cfg.Namespace = p.Namespace

	if _, err = client.Services(p.Namespace).Create(cfg); err != nil {
		return err
	}

	return nil
}
