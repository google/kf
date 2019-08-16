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

package install

import (
	"github.com/google/kf/pkg/kf/commands/install/gke"
	"github.com/spf13/cobra"
)

// NewInstallCommand creates a command that can install kf to various
// environments.
func NewInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [subcommand]",
		Short: "Install kf",
		Long: `Installs kf into a new Kubernetes cluster, optionally creating the
		cluster.

		WARNING: No checks are done on a cluster before installing a new version
		of kf. This means that if you target a cluster with a later version of kf
		then you can downgrade the system.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	installers := []*cobra.Command{
		// Add new installers below
		gke.NewGKECommand(),
	}

	for _, kfi := range installers {
		cmd.AddCommand(kfi)
	}

	return cmd
}
