package exporttok8s

import (
	"errors"
	"os"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/tektonutil"
	"github.com/spf13/cobra"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	//this image is taken from the image of kf stack v2 build task, used here because
	//we are using the similar logic to create the image url.
	RunAndBuildImage = "cloudfoundry/cflinuxfs3@sha256:5219e9e30000e43e5da17906581127b38fa6417f297f522e332a801e737928f5"

	//EmptyUrlError is the message returned if the user didn't give a url of the source package
	EmptyUrlError = "url of source package could not be empty"

	//EmptyDestinationError is the message returned if the user didn't give a image destination
	EmptyDestinationError = "destination of image could not be empty"
)

type pipelineYamlOptions struct {
	url              string
	buildPack        string
	skipDetect       string
	imageDestination string
}

func NewExportToK8s(cfg *config.KfParams) *cobra.Command {
	cmd := &cobra.Command{
		Hidden:  true,
		Use:     "export-to-k8s",
		Short:   "export yaml files for the app",
		Example: `kf export-to-k8s`,
		Args:    cobra.ExactArgs(5),
		Long: `
		The export-to-k8s command allows operators to export the Tekton Pipeline, PipelineRun
		and App deployment files.

		Users can edit and execute the Tekton yaml files. The pipeline would then export an
		App image URL. Users need to replace the image in the deployment file with the
		exported image URL, and execute the deployment file to deploy their App.
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			exportPath, opts, err := ValidateOptions(args)
			if err != nil {
				return err
			}

			pipelinespec := getPipelineSpec(*opts)

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

			pipelineYaml, err := yaml.Marshal(&pipeline)
			if err != nil {
				return err
			}

			if err := os.WriteFile(exportPath, pipelineYaml, os.ModePerm); err != nil {
				return err
			}

			return nil
		},
	}
	return cmd
}

func getPipelineSpec(opts pipelineYamlOptions) *tektonv1beta1.PipelineSpec {
	pipelineSpec := tektonv1beta1.PipelineSpec{
		Workspaces: []tektonv1beta1.PipelineWorkspaceDeclaration{
			{
				Name: "output",
			},
		},
		Params: []tektonv1beta1.ParamSpec{
			tektonutil.StringParam("GITHUB_URL", "The url of the source package."),
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
						Value: *tektonv1beta1.NewArrayOrString(opts.url),
					},
				},
			},
			// this task is to export the image
			{
				Name: "push",
				RunAfter: []string{
					"clone-code",
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
							tektonutil.StringParam("DESTINATION", "The destination of the image."),
						},
						Steps: []tektonv1beta1.Step{
							{
								Name:    "copy-lifecycle",
								Image:   "ko://code.cloudfoundry.org/buildpackapplifecycle/installer",
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

				Params: []tektonv1beta1.Param{
					{
						Name:  "BUILDPACKS",
						Value: *tektonv1beta1.NewArrayOrString(opts.buildPack),
					},
					{
						Name:  "RUN_IMAGE",
						Value: *tektonv1beta1.NewArrayOrString(RunAndBuildImage),
					},
					{
						Name:  "BUILDER_IMAGE",
						Value: *tektonv1beta1.NewArrayOrString(RunAndBuildImage),
					},
					{
						Name:  "SKIP_DETECT",
						Value: *tektonv1beta1.NewArrayOrString(opts.skipDetect),
					},
					{
						Name:  "IMAGE_DESTINATION",
						Value: *tektonv1beta1.NewArrayOrString(opts.imageDestination),
					},
				},
			},
		},
	}
	return &pipelineSpec
}

func ValidateOptions(args []string) (string, *pipelineYamlOptions, error) {
	exportPath, url, buildPack, skipDetect, imageDestination := args[0], args[1], args[2], args[3], args[4]
	if exportPath == "" {
		path, err := os.Getwd()
		if err != nil {
			return "", nil, err
		}
		exportPath = path + "/pipeline.yaml"
	}

	if url == "" {
		return url, nil, errors.New(EmptyUrlError)
	}

	if buildPack == "" {
		buildPack = "https://github.com/cloudfoundry/go-buildpack"
	}

	if skipDetect == "" {
		skipDetect = "true"
	}

	if imageDestination == "" {
		return url, nil, errors.New(EmptyDestinationError)
	}

	opts := pipelineYamlOptions{
		url:              url,
		buildPack:        buildPack,
		skipDetect:       skipDetect,
		imageDestination: imageDestination,
	}

	return exportPath, &opts, nil
}
