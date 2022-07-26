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

package stubs

import (
	"github.com/google/kf/v2/pkg/kf/internal/genericcli"
	"github.com/spf13/cobra"
)

// NewApiCommand allows the user to logout. Temporarily stubbed.
func NewApiCommand() *cobra.Command {
	alt := `The Kf analog to the API URL is the Kubernetes cluster. Using
gcloud, after authenticating, please run:

		gcloud container clusters get-credentials CLUSTER_NAME [--region GCP_REGION | --zone GCP_REGION]

This will generate a kubeconfig entry in $HOME/.kubeconfig including an entry for the cluster
server url.
This can be manually configured with kubectl with the following command:

		kubectl config set-cluster CLUSTER_NAME --server=URL [--insecure-skip-tls-verify=true]

See kf login for more information about using kubectl to configure the api url.
If gcloud was used to set the target cluster, information about this cluster can
be viewed with:

		kubectl config current-context

The result should be something like:
gke_GCP-PROJECT-NAME_GCP-ZONE-OR-REGION_CLUSTER-NAME.

To view the actual server url, please run:

		kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}'
`
	example := `  # Opens a browser window to login
  gcloud auth login
  # Uses cluster test-cluster in us-central1-a. Project is pre-configured as test-project
  gcloud container clusters get-credentials test-cluster --zone us-central1-a
  # Prints current context: gke_test-project_us-central1-a_test-cluster
  kubectl config current-context
  # Prints current server as an IPv4 address, e.g. 12.34.56.78
  kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}'

	# Configure the server url for demo-cluster as 10.0.0.1 with no verification
  kubectl config set-cluster demo-cluster --server=http://10.0.0.1 --insecure-skip-tls-verify=true
  # Set the current-context to demo-context, which already has cluster "demo-cluster" and user "demo-user"
  kubectl config use-context demo-context
`
	return genericcli.NewStubCommand(
		"api", "Target a Kubernetes cluster API endpoint.", alt, example)
}
