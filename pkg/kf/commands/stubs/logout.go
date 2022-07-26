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

// NewLogoutCommand allows the user to logout. Temporarily stubbed.
func NewLogoutCommand() *cobra.Command {
	alt := `To logout, you should revoke the authorization of your user account.
Using your GCP account email address as ACCOUNT, please run:

		gcloud auth revoke ACCOUNT

To revoke authorization for all accounts, please run:

		gcloud auth revoke --all

This automatically invalidates any kubeconfig credentials that were created with
gcloud, but it doesn't remove them. You can still follow the kubectl steps below
to completely remove the entries.

If using kubectl, please run:

		export USER_NAME=$(kubectl config view --minify -o jsonpath='{.contexts[0].context.user}')
		export CLUSTER_NAME=$(kubectl config view --minify -o jsonpath='{.contexts[0].context.cluster}')
		kubectl config unset users.$USER_NAME
		kubectl config delete-cluster $CLUSTER_NAME
		kubectl config delete-context $(kubectl config current-context)
		kubectl config unset current-context
`
	example := `  # Currently logged in with account test@example.com
  gcloud auth revoke test@example.com

  # View all accounts in gcloud
  gcloud auth list --format=json
  # Result: [{"account": "test@example.com", "status":ACTIVE}, {"account": "test@test-project.iam.gserviceaccount.com", "status":""}]
  # Now revoke authorization for all accounts
  gcloud auth revoke --all

  # The --minify flag lets you view only entries for the current context.
  # Using a jsonpath expression, get the first and only context, and extract the user name
  export USER_NAME=$(kubectl config view --minify -o jsonpath='{.contexts[0].context.user}')
  # This command is very similar but extracts the cluster name for the current context
  export CLUSTER_NAME=$(kubectl config view --minify -o jsonpath='{.contexts[0].context.cluster}')
  # This removes the entire user entry including credentials.
  kubectl config unset user.$USER_NAME
  # Delete the cluster entry
  kubectl config delete-cluster $CLUSTER_NAME
  # Delete the context entry
  kubectl config delete-context $(kubectl config current-context)
  # Unset the current-context field, because context it refers to no longer exists
  kubectl config unset current-context
`
	return genericcli.NewStubCommand(
		"logout",
		"Log out of a Kubernetes cluster.",
		alt,
		example,
		genericcli.WithStubAliases([]string{"lo"}))
}
