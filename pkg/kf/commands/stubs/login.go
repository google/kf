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

// NewLoginCommand allows the user to login. Temporarily stubbed.
func NewLoginCommand() *cobra.Command {
	alt := `If using gcloud, gcloud does not support entering user credentials
through CLI options. Run this:

		gcloud auth login
		gcloud container clusters get-credentials CLUSTER_NAME [--region GCP_REGION | --zone GCP_ZONE] [--project GCP_PROJECT]

Then you will be authenticated with your GCP user credentials and pointed to the cluster specified.
Alternatively, if using only kubectl, you can generate a context from your .kubeconfig with this:

		kubectl config set-cluster CLUSTER_NAME --server=API_ENDPOINT --certificate-authority=CA_FILE_PATH [--insecure-skip-tls-verify=true]
		kubectl config set-context CONTEXT_NAME --cluster=CLUSTER_NAME --user=USER_NAME
		kubectl config set-credentials USER_NAME [options...]

Where options can be any of the following corresponding to an equivalent cf option:

		cf login              | kubectl config set-...
		----------------------+-------------------------------
		-u                    | credentials --username=...
		-p                    | credentials --password=...
		--origin              | credentials --auth_provider=...
		-a                    | cluster --server=...
		-skip-ssl-validation  | cluster --insecure-skip-tls-verify=true
`
	example := `  # Opens a browser window to login with user credentials
  gcloud auth login
  # In test-project, there is test-cluster in the region us-west1. Add this config to ~/.kubeconfig
  gcloud container clusters get-credentials test-cluster --region us-west1 --project test-project

	# There is a cluster at 10.0.0.1, called demo-cluster. There is also a cert called demo.ca.crt
  kubectl config set-cluster demo-cluster --server=https://10.0.0.1 --certificate-authority=demo.ca.crt
  # Create a context using demo-cluster and a new user called demo-user
  kubectl config set-context demo-context --cluster=demo-cluster --user=demo-user
  # Create the user entry using a username and password
  kubectl config set-credentials demo-user --username=foo --password=bar
  # Set the current context so that kf uses the correct cluster
  kubectl config use-context demo-context
`
	return genericcli.NewStubCommand(
		"login",
		"Log in to a Kubernetes cluster interactively.",
		alt,
		example,
		genericcli.WithStubAliases([]string{"l"}))
}
