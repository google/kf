package exporttok8s

import (
	"bytes"
	"os"
	"testing"

	"github.com/google/kf/v2/pkg/kf/manifest"
	"github.com/google/kf/v2/pkg/kf/testutil"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func TestExportsToK8sCommand_sanity(t *testing.T) {

	params, _ := getParams("gcr.io/kf-source/testbuild", "test-app")
	pipelinespec := makePipelineSpec("https://github.com/cloudfoundry-samples/test-app", params)

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

	deployment := makeDeployment("test")

	pipelineYaml, _ := yaml.Marshal(pipeline)

	pipelinerunYaml, _ := yaml.Marshal(pipelinerun)

	str := [][]byte{pipelineYaml, pipelinerunYaml}
	buildImageYaml := bytes.Join(str, []byte("---\n"))

	deploymentYaml, _ := yaml.Marshal(deployment)

	testutil.AssertGolden(t, "build_image yaml", buildImageYaml)
	testutil.AssertGolden(t, "deployment yaml", deploymentYaml)
}

func TestGetParams(t *testing.T) {

	cases := map[string]struct {
		appManifest   *manifest.Manifest
		expectedValue *tektonv1beta1.ArrayOrString
	}{
		"no manifest": {
			appManifest:   nil,
			expectedValue: tektonv1beta1.NewArrayOrString("https://github.com/cloudfoundry/staticfile-buildpack,https://github.com/cloudfoundry/java-buildpack,https://github.com/cloudfoundry/ruby-buildpack,https://github.com/cloudfoundry/dotnet-core-buildpack,https://github.com/cloudfoundry/nodejs-buildpack,https://github.com/cloudfoundry/go-buildpack,https://github.com/cloudfoundry/python-buildpack,https://github.com/cloudfoundry/php-buildpack,https://github.com/cloudfoundry/binary-buildpack,https://github.com/cloudfoundry/nginx-buildpack"),
		},

		"have manifest, correct appName and no buildpack": {
			appManifest: &manifest.Manifest{
				Applications: []manifest.Application{
					{
						Name: "test-app",
					},
				},
			},
			expectedValue: tektonv1beta1.NewArrayOrString("https://github.com/cloudfoundry/staticfile-buildpack,https://github.com/cloudfoundry/java-buildpack,https://github.com/cloudfoundry/ruby-buildpack,https://github.com/cloudfoundry/dotnet-core-buildpack,https://github.com/cloudfoundry/nodejs-buildpack,https://github.com/cloudfoundry/go-buildpack,https://github.com/cloudfoundry/python-buildpack,https://github.com/cloudfoundry/php-buildpack,https://github.com/cloudfoundry/binary-buildpack,https://github.com/cloudfoundry/nginx-buildpack"),
		},

		"have manifest, correct appName and one buildpack": {
			appManifest: &manifest.Manifest{
				Applications: []manifest.Application{
					{
						Name:       "test-app",
						Buildpacks: []string{"https://github.com/cloudfoundry/java-buildpack"},
					},
				},
			},
			expectedValue: tektonv1beta1.NewArrayOrString("https://github.com/cloudfoundry/java-buildpack"),
		},

		"have manifest, correct appName and multiple buildpacks": {
			appManifest: &manifest.Manifest{
				Applications: []manifest.Application{
					{
						Name:       "test-app",
						Buildpacks: []string{"https://github.com/cloudfoundry/java-buildpack", "https://github.com/cloudfoundry/ruby-buildpack"},
					},
				},
			},
			expectedValue: tektonv1beta1.NewArrayOrString("https://github.com/cloudfoundry/java-buildpack", "https://github.com/cloudfoundry/ruby-buildpack"),
		},

		"have manifest, wrong appName and multiple buildpacks": {
			appManifest: &manifest.Manifest{
				Applications: []manifest.Application{
					{
						Name:       "test",
						Buildpacks: []string{"https://github.com/cloudfoundry/java-buildpack", "https://github.com/cloudfoundry/ruby-buildpack"},
					},
				},
			},
			expectedValue: tektonv1beta1.NewArrayOrString("https://github.com/cloudfoundry/staticfile-buildpack,https://github.com/cloudfoundry/java-buildpack,https://github.com/cloudfoundry/ruby-buildpack,https://github.com/cloudfoundry/dotnet-core-buildpack,https://github.com/cloudfoundry/nodejs-buildpack,https://github.com/cloudfoundry/go-buildpack,https://github.com/cloudfoundry/python-buildpack,https://github.com/cloudfoundry/php-buildpack,https://github.com/cloudfoundry/binary-buildpack,https://github.com/cloudfoundry/nginx-buildpack"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			if tc.appManifest != nil {
				manifestYaml, _ := yaml.Marshal(tc.appManifest)
				os.WriteFile("manifest.yml", manifestYaml, os.ModePerm)
			}

			var buildpack tektonv1beta1.ArrayOrString

			params, err := getParams("", "test-app")
			if err != nil {
				t.Fatalf("wanted err: %v, got: %v", nil, err)
			}

			for _, v := range params {
				if v.Name == "BUILDPACKS" {
					buildpack = v.Value
				}
			}

			if tc.appManifest != nil {
				os.Remove("manifest.yml")
			}

			testutil.AssertEqual(t, tn, tc.expectedValue, &buildpack)
		})
	}
}
