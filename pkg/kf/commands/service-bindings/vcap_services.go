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

	"github.com/google/kf/pkg/kf/cfutil"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	servicebindings "github.com/google/kf/pkg/kf/service-bindings"
	scv1beta1 "github.com/poy/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/spf13/cobra"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// NewVcapServicesCommand allows users to bind apps to service instances.
func NewVcapServicesCommand(
	p *config.KfParams,
	client servicebindings.ClientInterface,
	coreClient v1.CoreV1Interface,
	svcatClient scv1beta1.ServicecatalogV1beta1Interface) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "vcap-services APP_NAME",
		Short: "Print the VCAP_SERVICES environment variable for an app",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appName := args[0]

			cmd.SilenceUsage = true

			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			bindings, err := client.List(servicebindings.WithListAppName(appName), servicebindings.WithListNamespace(p.Namespace))
			if err != nil {
				return err
			}
			systemEnvInjector := cfutil.NewSystemEnvInjector(svcatClient.ServiceInstances(p.Namespace), coreClient.Secrets(p.Namespace))

			services, err := systemEnvInjector.GetVcapServices(appName, bindings)
			if err != nil {
				return err
			}

			output, err := cfutil.GetVcapServicesMap(appName, services)
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
