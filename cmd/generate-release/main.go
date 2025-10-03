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
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	cloudbuild "google.golang.org/api/cloudbuild/v1"
	"sigs.k8s.io/yaml"
)

const (
	TemplateTypeFreshCluster  = "fresh-cluster"
	TemplateTypeDeleteCluster = "delete-cluster"
	TemplateTypeInstallKf     = "install-kf"
	TemplateTypeUninstallKf   = "uninstall-kf"
)

var allTemplates = []string{TemplateTypeFreshCluster, TemplateTypeDeleteCluster, TemplateTypeInstallKf, TemplateTypeUninstallKf}

func main() {
	log.SetFlags(0)
	templateType := flag.String(
		"template-type",
		"",
		fmt.Sprintf(
			"Selects type of cloud build template to output. Options are %s",
			strings.Join(allTemplates, ", "),
		),
	)
	deployerImage := flag.String(
		"deployer-image",
		"",
		"The deployer image that should be referenced by the templates",
	)
	listTemplates := flag.Bool(
		"list-templates",
		false,
		"List template types. Don't generate anything.",
	)
	flag.Parse()

	if *listTemplates {
		for _, template := range allTemplates {
			fmt.Println(template)
		}
		return
	}

	if *deployerImage == "" {
		log.Fatal("deployer-image is required")
	}

	substitions := map[string]string{
		"_CLOUDSDK_CONTAINER_CLUSTER": "",
	}

	kfInstallSubstitions := map[string]string{
		"_CLOUDSDK_COMPUTE_ZONE": "",
		"_DOMAIN":                "$(SPACE_NAME).$(CLUSTER_INGRESS_IP).nip.io",
	}

	kfUninstallSubstitions := map[string]string{
		"_UNINSTALL_TEKTON": "false",
	}

	steps := []*cloudbuild.BuildStep{}
	switch *templateType {
	case TemplateTypeFreshCluster:
		steps = freshCluster(*deployerImage)
		substitions = v1alpha1.UnionMaps(substitions, kfInstallSubstitions, map[string]string{
			"_NODE_COUNT":      "3",
			"_MACHINE_TYPE":    "n1-highmem-4",
			"_NETWORK":         "default",
			"_RELEASE_CHANNEL": "REGULAR",
			"_ASM_MANAGED":     "false",
		})
	case TemplateTypeInstallKf:
		steps = installKf(*deployerImage)
		substitions = v1alpha1.UnionMaps(substitions, kfInstallSubstitions)
	case TemplateTypeUninstallKf:
		steps = uninstallKf(*deployerImage)
		substitions = v1alpha1.UnionMaps(substitions, kfUninstallSubstitions)
	case TemplateTypeDeleteCluster:
		steps = deleteCluster(*deployerImage)
	default:
		log.Fatalf("invalid template-type: %q", *templateType)
	}

	// Prepend the check substitions step.
	steps = append(checkSubstitutions(*deployerImage, substitions), steps...)

	build := cloudbuild.Build{
		Timeout:       "3600s",
		Substitutions: substitions,
		Options: &cloudbuild.BuildOptions{
			Env: []string{
				"CLOUDSDK_CONTAINER_CLUSTER=${_CLOUDSDK_CONTAINER_CLUSTER}",
				"CLOUDSDK_COMPUTE_ZONE=${_CLOUDSDK_COMPUTE_ZONE}",
			},
		},
		Steps: steps,
		Tags:  []string{"kf-cluster-operation"},
	}
	data, err := yaml.Marshal(build)
	if err != nil {
		log.Fatalf("failed to marshal to YAML: %v", err)
	}

	fmt.Println(string(data))
}

func checkSubstitutions(deployerImage string, substitions map[string]string) []*cloudbuild.BuildStep {
	lines := []string{}
	for name := range substitions {
		lines = append(
			lines,
			fmt.Sprintf(
				`if [ -z "${%s}" ]; then echo "%s is empty" && exit 1; fi`,
				name,
				name,
			),
		)
	}

	// We'll sort the lines to make it more deterministic.
	sort.Strings(lines)

	return []*cloudbuild.BuildStep{
		{
			Id:         "check substitutions",
			Name:       deployerImage,
			Entrypoint: "bash",
			Args: []string{
				"-c",
				multiLineCommand(lines...),
			},
		},
	}
}

func freshCluster(deployerImage string) []*cloudbuild.BuildStep {
	// XXX: This uses bash instead of gcloud directly so we can use the right
	// subcommand (create vs update).
	return append([]*cloudbuild.BuildStep{
		{
			Id:         "create GKE cluster",
			Name:       deployerImage,
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
		},
		{
			Id:         "setup Artifact Registry",
			Name:       deployerImage,
			Entrypoint: "/builder/setup-ar.bash",
			Args: []string{
				"${PROJECT_ID}",
				"${_CLOUDSDK_CONTAINER_CLUSTER}",
				"${_CLOUDSDK_COMPUTE_ZONE}",
				"${_CLOUDSDK_CONTAINER_CLUSTER}-sa@${PROJECT_ID}.iam.gserviceaccount.com",
			},
		},
		{
			Id:         "install ASM",
			Name:       deployerImage,
			Entrypoint: "/builder/install-asm.bash",
			Args: []string{
				"${PROJECT_ID}",
				"${_CLOUDSDK_CONTAINER_CLUSTER}",
				"${_CLOUDSDK_COMPUTE_ZONE}",
			},
			Env: []string{
				"ASM_MANAGED=${_ASM_MANAGED}",
			},
		},
	}, installKf(deployerImage)...)
}

func deleteCluster(deployerImage string) []*cloudbuild.BuildStep {
	return []*cloudbuild.BuildStep{
		{
			Id:         "delete GKE cluster",
			Name:       deployerImage,
			Entrypoint: "/builder/delete-cluster.bash",
			Args: []string{
				"${_CLOUDSDK_CONTAINER_CLUSTER}",
			},
		},
	}
}

func installKf(deployerImage string) []*cloudbuild.BuildStep {
	return []*cloudbuild.BuildStep{
		{
			Id:         "connect to cluster",
			Name:       deployerImage,
			Entrypoint: "gcloud",
			Args: []string{
				"container",
				"clusters",
				"get-credentials",
				"${_CLOUDSDK_CONTAINER_CLUSTER}",
				"--project=${PROJECT_ID}",
				"--zone=${_CLOUDSDK_COMPUTE_ZONE}",
			},
		}, {
			Id:         "install Tekton",
			Name:       deployerImage,
			Entrypoint: "kubectl",
			Args: []string{
				"apply", "--filename", "/kf/bin/tekton.yaml",
			},
		}, {
			Id:         "install KCC",
			Name:       deployerImage,
			Entrypoint: "/builder/setup-kcc.bash",
			Args: []string{
				"${PROJECT_ID}",
				"${_CLOUDSDK_CONTAINER_CLUSTER}",
			},
		}, {
			Id:         "install Kf Operator",
			Name:       deployerImage,
			Entrypoint: "/builder/install-kf-operator.bash",
		}, {
			Id:         "setup Workload Identity",
			Name:       deployerImage,
			Entrypoint: "/builder/setup-operator-wi.bash",
			Args: []string{
				"${PROJECT_ID}",
				"${_CLOUDSDK_CONTAINER_CLUSTER}",
			},
		}, {
			Id:         "Configure Kf Operator",
			Name:       deployerImage,
			Timeout:    "600s",
			Entrypoint: "/builder/configure-kf-operator.bash",
			Args: []string{
				"${PROJECT_ID}",
				"${_CLOUDSDK_COMPUTE_ZONE}",
				"${_CLOUDSDK_CONTAINER_CLUSTER}",
			},
		}, {
			Id:         "wait for cluster to become healthy",
			Name:       deployerImage,
			Timeout:    "600s",
			Entrypoint: "/builder/kf.bash",
			Args: []string{
				"doctor",
				"--retries", "15",
				"--delay", "20s",
			},
		},
	}
}

func uninstallKf(deployerImage string) []*cloudbuild.BuildStep {
	return []*cloudbuild.BuildStep{
		{
			Id:         "uninstall Kf",
			Name:       deployerImage,
			Entrypoint: "kubectl",
			Args: []string{
				"patch", "cloudrun", "-n", "kf-operator", "cloud-run", "--type='json'", "-p=[{'op': 'remove', 'path': '/spec/kf'}]",
			},
		},
		{
			Id:         "uninstall Kf Operator",
			Name:       deployerImage,
			Entrypoint: "kubectl",
			Args: []string{
				"delete", "--filename", "/kf/bin/operator.yaml",
			},
		}, {
			Id:         "uninstall Tekton",
			Name:       deployerImage,
			Entrypoint: "bash",
			Args: []string{
				"-ec",
				`if [ "${_UNINSTALL_TEKTON}" = "true" ]; then kubectl delete --filename /kf/bin/tekton.yaml; else echo skipping Tekton; fi`,
			},
		},
	}
}

// multiLineCommand takes several commands that would ideally be on multiple
// lines and combines them via &&
//
// XXX: This is necessary because there isn't a way to utilize the normal
// pattern for bash scripts:
//
// entrypoint: bash
// args:
// - "-c"
// - |
//
//	echo line 1
//	echo line 2
//
// The YAML marshaller doesn't let the pipe ('|') not have quotes. Therefore
// instead of treating the pipe as an indication the following string is all a
// single string, it instead makes it look like the pipe is the first
// argument.
func multiLineCommand(commands ...string) string {
	return strings.Join(commands, " && ")
}
