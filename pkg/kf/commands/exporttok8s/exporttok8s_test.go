package exporttok8s

import (
	"bytes"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func TestExportsToK8sCommand_sanity(t *testing.T) {

	pipelinespec := makePipelineSpec(pipelineYamlOptions{
		url:              "https://github.com/cloudfoundry-samples/test-app",
		buildPack:        "https://github.com/cloudfoundry/go-buildpack",
		skipDetect:       "true",
		imageDestination: "gcr.io/kf-source/testbuild",
	})

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

	deployment := makeDeployment()

	pipelineYaml, _ := yaml.Marshal(pipeline)

	pipelinerunYaml, _ := yaml.Marshal(pipelinerun)

	str := [][]byte{pipelineYaml, pipelinerunYaml}
	pipelineAndPipelinerunYaml := bytes.Join(str, []byte("---\n"))

	deploymentYaml, _ := yaml.Marshal(deployment)

	testutil.AssertGolden(t, "pipeline_pipelinerun yaml", pipelineAndPipelinerunYaml)
	testutil.AssertGolden(t, "deployment yaml", deploymentYaml)
}
