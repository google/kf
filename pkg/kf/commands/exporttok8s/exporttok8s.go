package exporttok8s

import (
	"os"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/tektonutil"
	"github.com/spf13/cobra"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func NewExportToK8s(cfg *config.KfParams) *cobra.Command {
	cmd := &cobra.Command{
		Hidden:  true,
		Use:     "export-to-k8s",
		Short:   "export yaml files for the app",
		Example: `kf export-to-k8s`,
		Args:    cobra.ExactArgs(2),
		Long: `
		The export-to-k8s command allows operators to export the Tekton Pipeline, PipelineRun
		and App deployment files.

		Users can edit and execute the Tekton yaml files. The pipeline would then export an
		App image URL. Users need to replace the image in the deployment file with the
		exported image URL, and execute the deployment file to deploy their App.
		`,
		RunE: func(cmd *cobra.Command, args []string) error {

			pipelinespec := getPipelineSpec(args[0])
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

			if args[1] != "" {
				if err := os.WriteFile(args[1], pipelineYaml, os.ModePerm); err != nil {
					return err
				}
			} else {
				path, err := os.Getwd()
				if err != nil {
					return err
				}

				if err := os.WriteFile(path+"/pipeline.yaml", pipelineYaml, os.ModePerm); err != nil {
					return err
				}
			}
			return nil
		},
	}
	return cmd
}

func getPipelineSpec(url string) *tektonv1beta1.PipelineSpec {
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
					Name: "git-clone",
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
		},
	}
	return &pipelineSpec
}
