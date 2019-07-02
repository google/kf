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

package servicebindings

import (
	"encoding/json"
	"fmt"

	"github.com/google/kf/pkg/kf/commands/config"
	servicebindings "github.com/google/kf/pkg/kf/service-bindings"
	"github.com/spf13/cobra"
)

// NewVcapServicesCommand allows users to bind apps to service instances.
func NewVcapServicesCommand(p *config.KfParams, client servicebindings.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vcap-services APP_NAME",
		Short: "Print the VCAP_SERVICES environment variable for an app",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appName := args[0]

			cmd.SilenceUsage = true

			output, err := client.GetVcapServices(appName,
				servicebindings.WithGetVcapServicesNamespace(p.Namespace))
			if err != nil {
				return err
			}

			out, err := json.Marshal(output)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(out))

			return nil
		},
	}

	return cmd
}
