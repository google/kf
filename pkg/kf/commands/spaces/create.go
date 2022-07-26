// Copyright 2019 Google LLC
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

package spaces

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/internal/envutil"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/describe"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/spaces"

	"github.com/spf13/cobra"
)

// NewCreateSpaceCommand allows users to create spaces.
func NewCreateSpaceCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	var (
		containerRegistry   string
		buildServiceAccount string
		domains             []string
		runningEnvVars      map[string]string
		stagingEnvVars      map[string]string
	)

	cmd := &cobra.Command{
		Use:   "create-space NAME",
		Short: "Create a Space with the given name.",
		Example: `
		# Create a Space with custom domains.
		kf create-space my-space --domain my-space.my-company.com

		# Create a Space that uses unique storage and service accounts.
		kf create-space my-space --container-registry gcr.io/my-project --build-service-account myserviceaccount

		# Set running and staging environment variables for Apps and Builds.
		kf create-space my-space --run-env=ENVIRONMENT=nonprod --stage-env=ENVIRONMENT=nonprod,JDK_VERSION=8
		`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			name := args[0]

			toCreate := &v1alpha1.Space{}
			toCreate.Name = name
			toCreate.Spec.BuildConfig.ContainerRegistry = containerRegistry
			toCreate.Spec.BuildConfig.ServiceAccount = buildServiceAccount
			toCreate.Spec.BuildConfig.Env = envutil.MapToEnvVars(stagingEnvVars)

			for _, domain := range domains {
				toCreate.Spec.NetworkConfig.Domains = append(toCreate.Spec.NetworkConfig.Domains, v1alpha1.SpaceDomain{Domain: domain})
			}

			toCreate.Spec.RuntimeConfig.Env = envutil.MapToEnvVars(runningEnvVars)

			if _, err := client.Create(cmd.Context(), toCreate); err != nil {
				return err
			}

			w := cmd.OutOrStdout()

			fmt.Fprintln(w, "Space requested, waiting for subcomponents to be created")
			space, err := client.WaitFor(context.Background(), name, 1*time.Second, spaces.IsStatusFinal)
			if err != nil {
				return err
			}
			fmt.Fprintln(w, "Space created")
			describe.DuckStatus(w, space.Status.Status)
			fmt.Fprintln(w)

			printAdditionalCommands(cmd.OutOrStdout(), name)
			return nil
		},
	}

	cmd.Flags().StringVar(
		&containerRegistry,
		"container-registry",
		"",
		"Container registry built Apps and source code will be stored in.",
	)

	cmd.Flags().StringVar(
		&buildServiceAccount,
		"build-service-account",
		"",
		"Service account that Builds will use.",
	)

	cmd.Flags().StringArrayVar(
		&domains,
		"domain",
		nil,
		"Sets the valid domains for the Space. The first provided domain is the default.",
	)

	cmd.Flags().StringToStringVar(
		&runningEnvVars,
		"run-env",
		nil,
		"Sets the running environment variables for all Apps in the Space.",
	)

	cmd.Flags().StringToStringVar(
		&stagingEnvVars,
		"stage-env",
		nil,
		"Sets the staging environment variables for all Builds in the Space.",
	)

	return cmd
}

func printAdditionalCommands(w io.Writer, spaceName string) {
	utils.SuggestNextAction(utils.NextAction{
		Description: "Get space info",
		Commands: []string{
			fmt.Sprintf("kf space %s", spaceName),
			fmt.Sprintf("kubectl get space %s", spaceName),
		},
	})

	utils.SuggestNextAction(utils.NextAction{
		Description: "Target space",
		Commands: []string{
			fmt.Sprintf("kf target -s %s", spaceName),
		},
	})

	utils.SuggestNextAction(utils.NextAction{
		Description: "Set space config",
		Commands: []string{
			"kf configure-space",
		},
	})

}
