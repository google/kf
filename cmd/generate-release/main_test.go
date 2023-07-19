// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	cloudbuild "google.golang.org/api/cloudbuild/v1"
	"sigs.k8s.io/yaml"
)

func TestMain(m *testing.M) {
	if os.Getenv("RUN_FOR_TEST") == "true" {
		args := os.Getenv("ARGS")
		os.Args = []string{"cloudbuild-generator"}
		if args != "" {
			os.Args = append(os.Args, strings.Split(args, ":")...)
		}
		main()
		return
	}

	os.Exit(m.Run())
}

func RunGenerator(args ...string) (string, error) {
	e, err := os.Executable()
	if err != nil {
		return "", err
	}

	cmd := exec.Command(e, args...)
	cmd.Env = []string{
		"RUN_FOR_TEST=true",
		"ARGS=" + strings.Join(args, ":"),
	}

	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))
	if err != nil {
		return "", errors.New(err.Error() + ": " + outputStr)
	}
	return outputStr, nil
}

func ParseBuild(output string, err error) (*cloudbuild.Build, error) {
	if err != nil {
		return nil, err
	}

	b := cloudbuild.Build{}
	if err := yaml.Unmarshal([]byte(output), &b); err != nil {
		return nil, err
	}

	return &b, nil
}

func TestListTemplates(t *testing.T) {
	t.Parallel()
	output, err := RunGenerator("--list-templates")
	testutil.AssertNil(t, "err", err)
	testutil.AssertEqual(t, "types", allTemplates, strings.Split(output, "\n"))
}

func TestNoDeployerImage(t *testing.T) {
	t.Parallel()
	_, err := ParseBuild(RunGenerator("--template-type=" + TemplateTypeUninstallKf))
	testutil.AssertErrorsEqual(t, errors.New(`exit status 1: deployer-image is required`), err)
}

func TestNoTemplateType(t *testing.T) {
	t.Parallel()
	_, err := RunGenerator("--deployer-image=deployer-image")
	testutil.AssertErrorsEqual(t, errors.New(`exit status 1: invalid template-type: ""`), err)
}

func TestInvalidTemplateType(t *testing.T) {
	t.Parallel()
	_, err := RunGenerator("--deployer-image=deployer-image", "--template-type=invalid")
	testutil.AssertErrorsEqual(t, errors.New(`exit status 1: invalid template-type: "invalid"`), err)
}

func TestBuildInfo(t *testing.T) {
	t.Parallel()

	expectedSubstitutions := buildExpectedSubstitutions()

	for _, templateType := range []string{TemplateTypeFreshCluster, TemplateTypeInstallKf, TemplateTypeUninstallKf} {
		t.Run(templateType, func(t *testing.T) {
			b, err := ParseBuild(RunGenerator("--deployer-image=deployer-image", "--template-type="+templateType))
			testutil.AssertNil(t, "error", err)
			testutil.AssertEqual(t, "timeout", "3600s", b.Timeout)
			testutil.AssertEqual(
				t,
				"substitutions",
				map[string]string(expectedSubstitutions[templateType]),
				b.Substitutions,
			)

			testutil.AssertNotNil(t, "options", b.Options)
			testutil.AssertEqual(t, "env", []string{
				"CLOUDSDK_CONTAINER_CLUSTER=${_CLOUDSDK_CONTAINER_CLUSTER}",
				"CLOUDSDK_COMPUTE_ZONE=${_CLOUDSDK_COMPUTE_ZONE}",
			}, b.Options.Env)
		})
	}
}

// ExpectedSubsititions is a map of substition name to default value. We give
// this its own type so for test claritiy.
type ExpectedSubsititions map[string]string

func buildExpectedSubstitutions() map[string]ExpectedSubsititions {
	// These are expected in each template.
	standardSubs := ExpectedSubsititions{
		"_CLOUDSDK_CONTAINER_CLUSTER": "",
	}

	// These are expected whenever Kf is intalled.
	installKfSubs := ExpectedSubsititions{
		"_CLOUDSDK_COMPUTE_ZONE": "",
		"_DOMAIN":                "$(SPACE_NAME).$(CLUSTER_INGRESS_IP).nip.io",
	}

	uninstallKfSubs := ExpectedSubsititions{
		"_UNINSTALL_TEKTON": "false",
	}

	return map[string]ExpectedSubsititions{
		TemplateTypeFreshCluster: v1alpha1.UnionMaps(map[string]string{
			"_NODE_COUNT":      "3",
			"_MACHINE_TYPE":    "n1-highmem-4",
			"_NETWORK":         "default",
			"_RELEASE_CHANNEL": "REGULAR",
			"_ASM_MANAGED":     "false",
		}, installKfSubs, standardSubs),
		TemplateTypeInstallKf: v1alpha1.UnionMaps(installKfSubs, standardSubs),

		// These only require the standard
		TemplateTypeDeleteCluster: standardSubs,
		TemplateTypeUninstallKf:   v1alpha1.UnionMaps(uninstallKfSubs, standardSubs),
	}
}

func TestInstallKf(t *testing.T) {
	t.Parallel()

	b, err := ParseBuild(RunGenerator("--deployer-image=deployer-image", "--template-type="+TemplateTypeInstallKf))
	testutil.AssertNil(t, "error", err)
	testutil.AssertEqual(t, "step len", 8, len(b.Steps))

	testCheckSubstitutions(t, b.Steps, TemplateTypeInstallKf)
	testInstallKf(t, b.Steps[1:])
}

func TestFreshCluster(t *testing.T) {
	t.Parallel()

	b, err := ParseBuild(RunGenerator("--deployer-image=deployer-image", "--template-type="+TemplateTypeFreshCluster))
	testutil.AssertNil(t, "error", err)
	testutil.AssertEqual(t, "step len", 11, len(b.Steps))

	testCheckSubstitutions(t, b.Steps, TemplateTypeFreshCluster)

	// Create GKE cluster
	testutil.AssertEqual(t, "create GKE cluster", &cloudbuild.BuildStep{
		Id:         "create GKE cluster",
		Name:       "deployer-image",
		Entrypoint: "/builder/create_dm_deployment.bash",
		Args: []string{
			"${PROJECT_ID}",
			"${_CLOUDSDK_CONTAINER_CLUSTER}",
			"${_CLOUDSDK_COMPUTE_ZONE}",
			"${_NODE_COUNT}",
			"${_MACHINE_TYPE}",
			"${_NETWORK}",
			"${_RELEASE_CHANNEL}",
		},
	}, b.Steps[1])

	// Setup Artifact Registry
	testutil.AssertEqual(t, "create GKE cluster", &cloudbuild.BuildStep{
		Id:         "setup Artifact Registry",
		Name:       "deployer-image",
		Entrypoint: "/builder/setup-ar.bash",
		Args: []string{
			"${PROJECT_ID}",
			"${_CLOUDSDK_CONTAINER_CLUSTER}",
			"${_CLOUDSDK_COMPUTE_ZONE}",
			"${_CLOUDSDK_CONTAINER_CLUSTER}-sa@${PROJECT_ID}.iam.gserviceaccount.com",
		},
	}, b.Steps[2])

	// Install ASM
	testutil.AssertEqual(t, "create GKE cluster", &cloudbuild.BuildStep{
		Id:         "install ASM",
		Name:       "deployer-image",
		Entrypoint: "/builder/install-asm.bash",
		Args: []string{
			"${PROJECT_ID}",
			"${_CLOUDSDK_CONTAINER_CLUSTER}",
			"${_CLOUDSDK_COMPUTE_ZONE}",
		},
		Env: []string{
			"ASM_MANAGED=${_ASM_MANAGED}",
		},
	}, b.Steps[3])

	testInstallKf(t, b.Steps[4:])
}

func TestDeleteCluster(t *testing.T) {
	t.Parallel()

	b, err := ParseBuild(RunGenerator("--deployer-image=deployer-image", "--template-type="+TemplateTypeDeleteCluster))
	testutil.AssertNil(t, "error", err)
	testutil.AssertEqual(t, "step len", 2, len(b.Steps))

	testCheckSubstitutions(t, b.Steps, TemplateTypeDeleteCluster)

	// Delete GKE cluster
	testutil.AssertEqual(t, "delete GKE cluster", &cloudbuild.BuildStep{
		Id:         "delete GKE cluster",
		Name:       "deployer-image",
		Entrypoint: "gcloud",
		Args: []string{
			"deployment-manager", "--quiet", "deployments", "delete", "${_CLOUDSDK_CONTAINER_CLUSTER}",
		},
	}, b.Steps[1])
}

func testCheckSubstitutions(t *testing.T, steps []*cloudbuild.BuildStep, templateType string) {
	t.Helper()

	lines := []string{}
	for name := range buildExpectedSubstitutions()[templateType] {
		lines = append(lines,
			fmt.Sprintf(
				`if [ -z "${%s}" ]; then echo "%s is empty" && exit 1; fi`,
				name,
				name,
			),
		)
	}

	// Sort so it's deterministic
	sort.Strings(lines)

	// Check that required substitutions are not empty
	testutil.AssertEqual(t, "check substitutions", &cloudbuild.BuildStep{
		Id:         "check substitutions",
		Name:       "deployer-image",
		Entrypoint: "bash",
		Args: []string{
			`-c`,
			multiLineCommand(lines...),
		},
	}, steps[0])
}

func testInstallKf(t *testing.T, steps []*cloudbuild.BuildStep) {

	stepnum := 0
	// Connect to cluster
	testutil.AssertEqual(t, "connect to cluster", &cloudbuild.BuildStep{
		Id:         "connect to cluster",
		Name:       "deployer-image",
		Entrypoint: "gcloud",
		Args: []string{
			"container",
			"clusters",
			"get-credentials",
			"${_CLOUDSDK_CONTAINER_CLUSTER}",
			"--project=${PROJECT_ID}",
			"--zone=${_CLOUDSDK_COMPUTE_ZONE}",
		},
	}, steps[stepnum])
	stepnum++

	// Install Tekton
	testutil.AssertEqual(t, "install Tekton", &cloudbuild.BuildStep{
		Id:         "install Tekton",
		Name:       "deployer-image",
		Entrypoint: "kubectl",
		Args: []string{
			"apply", "--filename", "/kf/bin/tekton.yaml",
		},
	}, steps[stepnum])
	stepnum++

	// Install KCC
	testutil.AssertEqual(t, "install KCC", &cloudbuild.BuildStep{
		Id:         "install KCC",
		Name:       "deployer-image",
		Entrypoint: "/builder/setup-kcc.bash",
		Args: []string{
			"${PROJECT_ID}",
			"${_CLOUDSDK_CONTAINER_CLUSTER}",
		},
	}, steps[stepnum])
	stepnum++

	// Install Kf Operator
	testutil.AssertEqual(t, "install Kf Operator", &cloudbuild.BuildStep{
		Id:         "install Kf Operator",
		Name:       "deployer-image",
		Entrypoint: "/builder/install-kf-operator.bash",
	}, steps[stepnum])
	stepnum++

	// Setup Kf Secrets
	testutil.AssertEqual(t, "setup Workload Identity", &cloudbuild.BuildStep{
		Id:         "setup Workload Identity",
		Name:       "deployer-image",
		Entrypoint: "/builder/setup-operator-wi.bash",
		Args: []string{
			"${PROJECT_ID}", "${_CLOUDSDK_CONTAINER_CLUSTER}",
		},
	}, steps[stepnum])
	stepnum++

	// Configure Kf Operator
	testutil.AssertEqual(t, "install Kf", &cloudbuild.BuildStep{
		Id:         "Configure Kf Operator",
		Name:       "deployer-image",
		Timeout:    "600s",
		Entrypoint: "/builder/configure-kf-operator.bash",
		Args: []string{
			"${PROJECT_ID}",
			"${_CLOUDSDK_COMPUTE_ZONE}",
			"${_CLOUDSDK_CONTAINER_CLUSTER}",
		},
	}, steps[stepnum])
	stepnum++

	// Wait for Kf
	testutil.AssertEqual(t, "wait for cluster to become healthy", &cloudbuild.BuildStep{
		Id:         "wait for cluster to become healthy",
		Name:       "deployer-image",
		Timeout:    "600s",
		Entrypoint: "/builder/kf.bash",
		Args: []string{
			"doctor", "--retries", "15", "--delay", "20s",
		},
	}, steps[stepnum])
	stepnum++
}

func TestUninstallKf(t *testing.T) {
	t.Parallel()
	b, err := ParseBuild(RunGenerator("--deployer-image=deployer-image", "--template-type="+TemplateTypeUninstallKf))
	testutil.AssertNil(t, "error", err)
	testutil.AssertEqual(t, "step len", 4, len(b.Steps))

	testCheckSubstitutions(t, b.Steps, TemplateTypeUninstallKf)

	// Kf Uninstall
	testutil.AssertEqual(t, "uninstall Kf", &cloudbuild.BuildStep{
		Id:         "uninstall Kf",
		Name:       "deployer-image",
		Entrypoint: "kubectl",
		Args: []string{
			"patch", "cloudrun", "-n", "kf-operator", "cloud-run", "--type='json'", "-p=[{'op': 'remove', 'path': '/spec/kf'}]",
		},
	}, b.Steps[1])

	// Kf Operator Uninstall
	testutil.AssertEqual(t, "uninstall Kf Operator", &cloudbuild.BuildStep{
		Id:         "uninstall Kf Operator",
		Name:       "deployer-image",
		Entrypoint: "kubectl",
		Args: []string{
			"delete", "--filename", "/kf/bin/operator.yaml",
		},
	}, b.Steps[2])

	// Tekton Uninstall
	testutil.AssertEqual(t, "uninstall Tekton", &cloudbuild.BuildStep{
		Id:         "uninstall Tekton",
		Name:       "deployer-image",
		Entrypoint: "bash",
		Args: []string{
			"-ec",
			`if [ "${_UNINSTALL_TEKTON}" = "true" ]; then kubectl delete --filename /kf/bin/tekton.yaml; else echo skipping Tekton; fi`,
		},
	}, b.Steps[3])
}
