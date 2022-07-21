package exporttok8s

import (
	"fmt"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func TestExportsToK8sCommand_sanity(t *testing.T) {

	pipelinespec := getPipelineSpec("https://github.com/cloudfoundry-samples/test-app")
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
		fmt.Printf("Error while Marshaling pipelineYaml. %v", err)
	}

	testutil.AssertGolden(t, "yaml is correct", pipelineYaml)
}
