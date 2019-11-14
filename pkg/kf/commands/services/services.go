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

package services

import (
	"fmt"
	"io"
	"strings"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/describe"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/marketplace"
	"github.com/google/kf/pkg/kf/services"
	"github.com/spf13/cobra"
)

// NewListServicesCommand allows users to list service instances.
func NewListServicesCommand(
	p *config.KfParams,
	client services.Client,
	appsClient apps.Client,
	marketplaceClient marketplace.ClientInterface,
) *cobra.Command {
	servicesCommand := &cobra.Command{
		Use:     "services",
		Aliases: []string{"s"},
		Short:   "List service instances",
		Long:    `Lists all service instances in the target space.`,
		Example: `kf services`,
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			instances, err := client.List(p.Namespace)
			if err != nil {
				return err
			}

			apps, err := appsClient.List(p.Namespace)
			if err != nil {
				return err
			}
			ma := mapAppToServices(apps)

			if err := describe.TabbedWriter(cmd.OutOrStdout(), func(w io.Writer) error {
				if _, err := fmt.Fprintln(w, "Name\tService\tPlan\tBound Apps\tLast Operation\tBroker"); err != nil {
					return err
				}
				for _, instance := range instances {
					lastCond := services.LastStatusCondition(instance)
					var brokerInfo string
					brokerInfo, err = marketplaceClient.BrokerName(instance)
					if err != nil {
						brokerInfo = fmt.Sprintf("error finding broker: %s", err)
					}

					className := instance.Spec.ClusterServiceClassExternalName
					planName := instance.Spec.ClusterServicePlanExternalName

					if instance.Spec.ServiceClassRef != nil {
						className = instance.Spec.ServiceClassExternalName
						planName = instance.Spec.ServicePlanExternalName
					}

					if _, err := fmt.Fprintf(
						w,
						"%s\t%s\t%s\t%s\t%s\t%s\n",
						instance.Name,                         // Name
						className,                             // Service
						planName,                              // Plan
						strings.Join(ma[instance.Name], ", "), // Bound Apps
						lastCond.Reason,                       // Last Operation
						brokerInfo,                            // Broker
					); err != nil {
						return err
					}
				}

				return nil
			}); err != nil {
				return err
			}

			return nil
		},
	}

	return servicesCommand
}

func mapAppToServices(apps []v1alpha1.App) map[string][]string {
	m := map[string][]string{}
	for _, app := range apps {
		for _, binding := range app.Spec.ServiceBindings {
			m[binding.BindingName] = append(m[binding.BindingName], app.Name)
		}
	}
	return m
}
