package exporttok8s

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"os"
	"strings"

	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/manifest"
	"github.com/google/kf/v2/pkg/kf/tektonutil"
	"github.com/google/kf/v2/pkg/reconciler/app/resources"
	"github.com/spf13/cobra"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"
)

//go:embed resource/clone-code.yaml
var cloneTaskYaml []byte

type options struct {
	appName          string
	exportPath       string
	url              string
	imageDestination string
}

func NewExportToK8s(cfg *config.KfParams) *cobra.Command {
	cmd := &cobra.Command{
		Hidden:  true,
		Use:     "export-to-k8s",
		Short:   "export yaml files for the app",
		Example: `kf export-to-k8s`,
		Args:    cobra.ExactArgs(4),
		Long: `
		The export-to-k8s command allows operators to export the Tekton Pipeline, PipelineRun
		and App deployment files.

		Users can edit and execute the Tekton yaml files. The pipeline would then export an
		App image URL. Users need to replace the image in the deployment file with the
		exported image URL, and execute the deployment file to deploy their App.
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := ValidateOptions(args)
			if err != nil {
				return err
			}

			var app *manifest.Application
			appManifest, err := manifest.CheckForManifest(context.Background(), nil)
			if err != nil {
				return err
			}

			if appManifest == nil {
				app = &manifest.Application{
					Name: opts.appName,
				}
			} else {
				app, err = appManifest.App(opts.appName)
				if err != nil {
					return err
				}
			}

			params, err := getParams(opts.imageDestination, app)
			if err != nil {
				return err
			}

			pipelinespec := makePipelineSpec(opts.url, params)

			pipeline := tektonv1beta1.Pipeline{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pipeline",
					APIVersion: "tekton.dev/v1beta1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "build-and-publish",
				},
				Spec: *pipelinespec,
			}

			pipelinerun := makePipelineRun(pipelinespec)

			deployment, err := makeDeployment(app)
			if err != nil {
				return err
			}

			pipelineYaml, err := yaml.Marshal(pipeline)
			if err != nil {
				return err
			}

			pipelinerunYaml, err := yaml.Marshal(pipelinerun)
			if err != nil {
				return err
			}

			yamls := [][]byte{cloneTaskYaml, pipelineYaml, pipelinerunYaml}
			buildImageYaml := bytes.Join(yamls, []byte("---\n"))

			deploymentYaml, err := yaml.Marshal(deployment)
			if err != nil {
				return err
			}

			if err := os.WriteFile(opts.exportPath+"/build_image.yaml", buildImageYaml, os.ModePerm); err != nil {
				return err
			}

			if err := os.WriteFile(opts.exportPath+"/deployment.yaml", deploymentYaml, os.ModePerm); err != nil {
				return err
			}

			return nil
		},
	}
	return cmd
}

func makePipelineSpec(url string, params []tektonv1beta1.Param) *tektonv1beta1.PipelineSpec {
	pipelineSpec := tektonv1beta1.PipelineSpec{
		Workspaces: []tektonv1beta1.PipelineWorkspaceDeclaration{
			{
				Name: "output",
			},
		},

		Tasks: []tektonv1beta1.PipelineTask{
			// this task is to clone the github url which users upload their source package
			{
				Name: "source-upload",
				TaskRef: &tektonv1beta1.TaskRef{
					Name: "clone-code",
				},
				Workspaces: []tektonv1beta1.WorkspacePipelineTaskBinding{
					{
						Workspace: "output",
						Name:      "output",
					},
				},
				Params: []tektonv1beta1.Param{
					{
						Name:  "url",
						Value: *tektonv1beta1.NewArrayOrString(url),
					},
				},
			},
			// this task is to export the image
			{
				Name: "push",
				RunAfter: []string{
					"source-upload",
				},
				TaskSpec: &tektonv1beta1.EmbeddedTask{
					TaskSpec: tektonv1beta1.TaskSpec{
						Workspaces: []tektonv1beta1.WorkspaceDeclaration{
							{
								Name: "output",
							},
						},
						Params: []tektonv1beta1.ParamSpec{
							tektonutil.StringParam("BUILDPACKS", "Ordered list of comma separated builtpacks to attempt."),
							tektonutil.StringParam("RUN_IMAGE", "The run image apps will use as the base for IMAGE (output)."),
							tektonutil.StringParam("BUILDER_IMAGE", "The image on which builds will run."),
							tektonutil.DefaultStringParam("SKIP_DETECT", "Skip the detect phase", "false"),
							tektonutil.StringParam("IMAGE_DESTINATION", "The destination of the image."),
						},
						Steps: []tektonv1beta1.Step{
							{
								Name:    "copy-lifecycle",
								Image:   "gcr.io/kf-releases/installer-d148684b3032e4386ff76c190d42c7d0:latest",
								Command: []string{"/ko-app/installer"},
								VolumeMounts: []corev1.VolumeMount{
									{Name: "staging-tmp-dir", MountPath: "/staging"},
								},
							},
							{
								Name:  "run-lifecycle",
								Image: "$(inputs.params.BUILDER_IMAGE)",
								// NOTE: this command shouldn't be run as root, instead it should be run as
								// vcap:vcap
								Command: []string{"bash"},
								// A /tmp directory is necessary because some buildpacks use /tmp
								// which causes cross-device links to be made because Tekton mounts
								// the /workspace directory.
								Args: []string{
									"-euc",
									`
echo "/staging/app /tmp/app" | xargs -n 1 cp -r /workspace/output/
CF_STACK=cflinuxfs3 /workspace/builder \
  -buildArtifactsCacheDir=/tmp/cache \
  -buildDir=/tmp/app \
  -buildpacksDir=/tmp/buildpacks \
  -outputBuildArtifactsCache=/tmp/output-cache \
  -outputDroplet=/tmp/droplet \
  -outputMetadata=/tmp/result.json \
  "-buildpackOrder=$(inputs.params.BUILDPACKS)" \
  "-skipDetect=$(inputs.params.SKIP_DETECT)"
cp -r /tmp/droplet /workspace/droplet

cat << 'EOF' > /workspace/entrypoint.bash
#!/usr/bin/env bash
set -e

if [[ "$@" == "" ]]; then
  exec /lifecycle/launcher "/home/vcap/app" "" ""
else
  exec /lifecycle/launcher "/home/vcap/app" "$@" ""
fi

EOF
chmod a+x /workspace/entrypoint.bash

cat << 'EOF' > /workspace/Dockerfile
FROM $(inputs.params.RUN_IMAGE)
COPY launcher /lifecycle/launcher
COPY entrypoint.bash /lifecycle/entrypoint.bash
WORKDIR /home/vcap
USER vcap:vcap
COPY droplet droplet.tar.gz
RUN tar -xzf droplet.tar.gz && rm droplet.tar.gz
ENTRYPOINT ["/lifecycle/entrypoint.bash"]
EOF
`,
								},
								VolumeMounts: []corev1.VolumeMount{
									{Name: "staging-tmp-dir", MountPath: "/staging"},
								},
							},
							{
								Name:       "build",
								WorkingDir: "/workspace",
								Command:    []string{"/kaniko/executor"},
								Image:      "gcr.io/kaniko-project/executor:latest",
								Args: []string{
									"--dockerfile",
									"/workspace/Dockerfile",
									"--context",
									"/workspace",
									"--destination",
									"$(inputs.params.IMAGE_DESTINATION)",
									"--oci-layout-path",
									"/tekton/home/image-outputs/IMAGE",
									"--single-snapshot",
									"--no-push",
									"--tarPath",
									"/workspace/image.tar",
								},
								VolumeMounts: []corev1.VolumeMount{
									{Name: "cache-dir", MountPath: "/cache"},
									{Name: "staging-tmp-dir", MountPath: "/workspace/staging"},
								},
							},
							{
								Name:       "publish",
								WorkingDir: "/workspace",
								Command:    []string{"/ko-app/build-helpers"},
								Image:      "gcr.io/kf-releases/build-helpers-58e723758a11c5e698f0be6f53cdecbc:latest",
								Args: []string{
									"publish",
									"$(inputs.params.IMAGE_DESTINATION)",
									"/workspace/image.tar",
								},
							},
						},
						Volumes: []corev1.Volume{
							tektonutil.EmptyVolume("cache-dir"),
							tektonutil.EmptyVolume("staging-tmp-dir"),
						},
					},
				},
				Workspaces: []tektonv1beta1.WorkspacePipelineTaskBinding{
					{
						Workspace: "output",
						Name:      "output",
					},
				},

				Params: params,
			},
		},
	}

	return &pipelineSpec
}

func makePipelineRun(pipelineSpec *tektonv1beta1.PipelineSpec) *tektonv1beta1.PipelineRun {
	pipelineRun := tektonv1beta1.PipelineRun{
		TypeMeta: metav1.TypeMeta{
			Kind:       "pipelinerun",
			APIVersion: "tekton.dev/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "build-and-publish-run",
		},
		Spec: tektonv1beta1.PipelineRunSpec{
			PipelineSpec: pipelineSpec,
			Workspaces: []tektonv1beta1.WorkspaceBinding{
				{
					Name: "output",
					VolumeClaimTemplate: &corev1.PersistentVolumeClaim{
						Spec: corev1.PersistentVolumeClaimSpec{
							AccessModes: []corev1.PersistentVolumeAccessMode{
								corev1.ReadWriteOnce,
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"storage": resource.MustParse("1Gi"),
								},
							},
						},
					},
				},
			},
		},
	}
	return &pipelineRun
}

func ValidateOptions(args []string) (*options, error) {
	appName, exportPath, url, imageDestination := args[0], args[1], args[2], args[3]
	if appName == "" {
		return nil, errors.New("app name should not be empty")
	}

	if exportPath == "" {
		path, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		exportPath = path
	}

	if url == "" {
		return nil, errors.New("url of source package should not be empty")
	}

	if imageDestination == "" {
		return nil, errors.New("destination of image should not be empty")
	}

	opts := options{
		appName:          appName,
		exportPath:       exportPath,
		url:              url,
		imageDestination: imageDestination,
	}

	return &opts, nil
}

func makeDeployment(app *manifest.Application) (*appsv1.Deployment, error) {
	labelMap := make(map[string]string)
	labelMap["app"] = app.Name

	container, err := getContainer(app)
	if err != nil {
		return nil, err
	}

	replicas := getReplicas(app)

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: replicas,
			Selector: metav1.SetAsLabelSelector(labelMap),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labelMap,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						*container,
					},
				},
			},
		},
	}, nil
}

func getParams(destination string, app *manifest.Application) ([]tektonv1beta1.Param, error) {
	var params []tektonv1beta1.Param
	var buildpack *tektonv1beta1.ArrayOrString

	defaultBuildspec, err := getBuildSpec()
	if err != nil {
		return nil, err
	}

	defaultParams := defaultBuildspec.Params
	for _, v := range defaultParams {
		if v.Name == "BUILDPACKS" {
			buildpack = tektonv1beta1.NewArrayOrString(v.Value)
			continue
		}

		if v.Name != "SOURCE_IMAGE" {
			params = append(params, tektonv1beta1.Param{Name: v.Name, Value: *tektonv1beta1.NewArrayOrString(v.Value)})
		}
	}

	params = append(params, tektonv1beta1.Param{Name: "IMAGE_DESTINATION", Value: *tektonv1beta1.NewArrayOrString(destination)})

	findBuildpack := app.Buildpacks
	if len(findBuildpack) > 0 {
		buildpack = tektonv1beta1.NewArrayOrString(strings.Join(findBuildpack, ","))
	}
	params = append(params, tektonv1beta1.Param{Name: "BUILDPACKS", Value: *buildpack})

	return params, nil
}

//set Skip-Detect parameter to false and get all the default Buildpacks
func getBuildSpec() (*v1alpha1.BuildSpec, error) {
	app := manifest.Application{}

	BuildSpecBuilder, _, err := app.DetectBuildType(v1alpha1.SpaceStatusBuildConfig{
		BuildpacksV2: []kfconfig.BuildpackV2Definition{
			{
				Disabled: false,
				Name:     "staticfile_buildpack",
				URL:      "https://github.com/cloudfoundry/staticfile-buildpack",
			},
			{
				Disabled: false,
				Name:     "java_buildpack",
				URL:      "https://github.com/cloudfoundry/java-buildpack",
			},
			{
				Disabled: false,
				Name:     "ruby_buildpack",
				URL:      "https://github.com/cloudfoundry/ruby-buildpack",
			},
			{
				Disabled: false,
				Name:     "dotnet_core_buildpack",
				URL:      "https://github.com/cloudfoundry/dotnet-core-buildpack",
			},
			{
				Disabled: false,
				Name:     "nodejs_buildpack",
				URL:      "https://github.com/cloudfoundry/nodejs-buildpack",
			},
			{
				Disabled: false,
				Name:     "go_buildpack",
				URL:      "https://github.com/cloudfoundry/go-buildpack",
			},
			{
				Disabled: false,
				Name:     "python_buildpack",
				URL:      "https://github.com/cloudfoundry/python-buildpack",
			},
			{
				Disabled: false,
				Name:     "php_buildpack",
				URL:      "https://github.com/cloudfoundry/php-buildpack",
			},
			{
				Disabled: false,
				Name:     "binary_buildpack",
				URL:      "https://github.com/cloudfoundry/binary-buildpack",
			},
			{
				Disabled: false,
				Name:     "nginx_buildpack",
				URL:      "https://github.com/cloudfoundry/nginx-buildpack",
			},
		},
		StacksV2: kfconfig.StackV2List{
			{
				Image: "cloudfoundry/cflinuxfs3@sha256:5219e9e30000e43e5da17906581127b38fa6417f297f522e332a801e737928f5",
				Name:  "cflinuxfs3",
			},
		},
	})
	if err != nil {
		return nil, err
	}

	buildspec, err := BuildSpecBuilder("")
	if err != nil {
		return nil, err
	}
	return buildspec, nil
}

func getContainer(app *manifest.Application) (*corev1.Container, error) {
	var container corev1.Container

	container, err := app.ToContainer()
	if err != nil {
		return nil, err
	}

	probe := container.ReadinessProbe
	if len(container.Ports) == 0 {
		rewriteProbe(probe, resources.DefaultUserPort)
		container.Ports = append(container.Ports, corev1.ContainerPort{
			Name:          resources.UserPortName,
			ContainerPort: resources.DefaultUserPort,
		})
	} else {
		rewriteProbe(probe, container.Ports[0].ContainerPort)
	}

	container.Name = app.Name
	container.Image = "placeholder"

	return &container, nil
}

func getReplicas(app *manifest.Application) *int32 {
	replicas := int32(1)

	if app.Instances != nil {
		replicas = *app.Instances
	}

	return &replicas
}

func rewriteProbe(p *corev1.Probe, defaultPort int32) {
	switch {
	case p == nil:
		return
	case p.HTTPGet != nil:
		p.HTTPGet.Port = intstr.FromInt(int(defaultPort))
	case p.TCPSocket != nil:
		p.TCPSocket.Port = intstr.FromInt(int(defaultPort))
	}
}
