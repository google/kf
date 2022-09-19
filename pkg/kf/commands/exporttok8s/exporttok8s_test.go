package exporttok8s

import (
	"bytes"
	_ "embed"
	"os"
	"testing"

	"github.com/google/kf/v2/pkg/kf/manifest"
	"github.com/google/kf/v2/pkg/kf/testutil"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func TestExportsToK8sCommand_sanity(t *testing.T) {

	app := manifest.Application{Name: "test"}
	app.HealthCheckType = "process"
	params, _ := getParams("gcr.io/kf-source/testbuild", &app)
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

	deployment, _ := makeDeployment(&app)

	pipelineYaml, _ := yaml.Marshal(pipeline)

	pipelinerunYaml, _ := yaml.Marshal(pipelinerun)

	yamls := [][]byte{cloneTaskYaml, pipelineYaml, pipelinerunYaml}

	buildImageYaml := bytes.Join(yamls, []byte("---\n"))

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
			expectedValue: tektonv1beta1.NewArrayOrString("https://github.com/cloudfoundry/java-buildpack,https://github.com/cloudfoundry/ruby-buildpack"),
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
			expectedValue: tektonv1beta1.NewArrayOrString("https://github.com/cloudfoundry/java-buildpack,https://github.com/cloudfoundry/ruby-buildpack"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			if tc.appManifest != nil {
				manifestYaml, _ := yaml.Marshal(tc.appManifest)
				os.WriteFile("manifest.yml", manifestYaml, os.ModePerm)
			}

			var buildpack tektonv1beta1.ArrayOrString

			var app *manifest.Application
			if tc.appManifest == nil {
				app = &manifest.Application{
					Name: "test-app",
				}
			} else {
				app, _ = tc.appManifest.App("test-app")
			}

			params, err := getParams("", app)
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

func TestGetContainer(t *testing.T) {

	cases := map[string]struct {
		appManifest *manifest.Manifest
	}{
		"no manifest": {
			appManifest: nil,
		},
		"have manifest with the properties": {
			appManifest: &manifest.Manifest{
				Applications: []manifest.Application{
					{
						Name:                   "test-app",
						DiskQuota:              "1024M",
						Memory:                 "512M",
						KfApplicationExtension: manifest.KfApplicationExtension{CPU: "1"},
						HealthCheckTimeout:     60,
						HealthCheckType:        "http",
					},
				},
			},
		},
		"have manifest but no health check": {
			appManifest: &manifest.Manifest{
				Applications: []manifest.Application{
					{
						Name:                   "test-app",
						DiskQuota:              "1024M",
						Memory:                 "512M",
						KfApplicationExtension: manifest.KfApplicationExtension{CPU: "1"},
					},
				},
			},
		},
		"have manifest but no CPU, DiskQuota,Memory": {
			appManifest: &manifest.Manifest{
				Applications: []manifest.Application{
					{
						Name:                    "test-app",
						HealthCheckTimeout:      60,
						HealthCheckType:         "http",
						HealthCheckHTTPEndpoint: "",
					},
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			if tc.appManifest != nil {
				manifestYaml, _ := yaml.Marshal(tc.appManifest)
				os.WriteFile("manifest.yml", manifestYaml, os.ModePerm)
				defer os.Remove("manifest.yml")
			}

			var app *manifest.Application
			if tc.appManifest == nil {
				app = &manifest.Application{
					Name: "test-app",
				}
			} else {
				app, _ = tc.appManifest.App("test-app")
			}

			gotContainer, err := getContainer(app)
			if err != nil {
				t.Fatalf("wanted err: %v, got: %v", nil, err)
			} else {
				testutil.AssertGoldenJSON(t, "container", gotContainer)
			}
		})
	}
}

func getResourceQuantity(str string) resource.Quantity {
	quantity, _ := resource.ParseQuantity(str)
	return quantity
}

func TestGetReplicas(t *testing.T) {
	cases := map[string]struct {
		appManifest   *manifest.Manifest
		expectedValue int32
	}{
		"no manifest": {
			appManifest:   nil,
			expectedValue: 1,
		},
		"manifest without instances": {
			appManifest: &manifest.Manifest{
				Applications: []manifest.Application{
					{
						Name: "test-app",
					},
				},
			},
			expectedValue: 1,
		},
		"manifest with instances": {
			appManifest: &manifest.Manifest{
				Applications: []manifest.Application{
					{
						Name:      "test-app",
						Instances: GetIntPointer(2),
					},
				},
			},
			expectedValue: 2,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			if tc.appManifest != nil {
				manifestYaml, _ := yaml.Marshal(tc.appManifest)
				os.WriteFile("manifest.yml", manifestYaml, os.ModePerm)
			}

			var app *manifest.Application
			if tc.appManifest == nil {
				app = &manifest.Application{
					Name: "test-app",
				}
			} else {
				app, _ = tc.appManifest.App("test-app")
			}

			gotValue := getReplicas(app)

			if tc.appManifest != nil {
				os.Remove("manifest.yml")
			}

			testutil.AssertEqual(t, "", tc.expectedValue, *gotValue)
		})
	}
}

func GetIntPointer(value int32) *int32 {
	return &value
}
