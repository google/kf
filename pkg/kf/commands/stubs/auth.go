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

// NewAuthCommand allows the user to authenticate. Temporarily stubbed.
func NewAuthCommand() *cobra.Command {
	alt := `If using gcloud, it is not possible to authenticate with user
account credentials in a non-interactive way. Instead, a service account should
be configured. Then run:

	gcloud auth activate-service-account --key-file=KEY_FILE
	gcloud container clusters get-credentials CLUSTER_NAME [--region GCP_REGION | --zone GCP_ZONE]

Here, KEY_FILE is the path to a json key file for your service account.

Using kubectl, it is possible authenticate non-interactively by manually setting
up your kubeconfig context. Please run:

		kubectl config set-cluster CLUSTER_NAME --server=API_ENDPOINT [--certificate-authority=CA_FILE_PATH] [--insecure-skip-tls-verify=true]
		kubectl config set-context CONTEXT_NAME --cluster=CLUSTER_NAME --user=USER_NAME
		kubectl config set-credentials USER_NAME [options...]
		kubectl config use-context CONTEXT_NAME

Where options can be any of the following corresponding to an equivalent cf option:

		cf auth               | kubectl config set-...
		----------------------+-------------------------------
		USERNAME              | credentials: --username=...
		PASSWORD              | credentials: --password=...
		--origin              | credentials: --auth_provider=...
		--client-credentials  | credentials: --auth_provider=oidc
		CLIENT_ID             | credentials: --auth-provider-arg=client-id=...
		CLIENT_SECRET         | credentials: --auth-provider-arg=client-secret=...
`
	example := `  # Using a key file called service-account.json, authenticate with gcloud
  gcloud auth activate-service-account --key-file=service-account.json
  # Service account is in test-project, which has cluster test-cluster and zone us-central1-a
  gcloud container clusters get-credentials test-cluster --zone us-central1-a

  # Using cluster demo-cluster at 10.0.0.1, with certificate authority file test-cert.ca.crt
  kubectl config set-cluster demo-cluster --server=https://10.0.0.1 --certificate-authority=test-cert.ca.crt
  # Creating a new context containing demo-cluster and a user called test-service-account
  kubectl config set-context demo-context --cluster=demo-cluster --user=test-service-account
  # For test-service-account, there is a client id and client secret with OpenID Connect
  kubectl config set-credentials test-service-account --auth-provider=oidc \
    --auth-provider-arg=client-id=foo --auth-provider-arg=client-secret=bar
	# Set the current context so that kf uses the correct cluster
  kubectl config use-context demo-context
`
	return genericcli.NewStubCommand(
		"auth", "Manually authenticate to a Kubernetes cluster.", alt, example)
}
