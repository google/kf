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

package apps

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/google/kf/pkg/kf/describe"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

// NewAppsCommand creates a apps command.
func NewAppsCommand(p *config.KfParams, appsClient apps.Client) *cobra.Command {
	return &cobra.Command{
		Use:     "apps",
		Short:   "List pushed apps",
		Long:    ``,
		Example: `kf apps`,
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}
			cmd.SilenceUsage = true

			fmt.Fprintf(cmd.OutOrStdout(), "Getting apps in space %s\n\n", p.Namespace)

			applist, err := appsClient.List(p.Namespace)
			if err != nil {
				return err
			}

			describe.TabbedWriter(cmd.OutOrStdout(), func(w io.Writer) {
				fmt.Fprintln(w, "Name\tRequested State\tInstances\tMemory\tDisk\tURLs\tCluster URL")
				for _, app := range applist {

					// Requested State
					var requestedState string
					switch {
					case !app.DeletionTimestamp.IsZero():
						requestedState = "deleting"
					case app.Spec.Instances.Stopped:
						requestedState = "stopped"
					case !app.Status.IsReady():
						requestedState = "not ready"
					default:
						requestedState = "ready"
					}

					// Instances
					var instances string
					switch {
					case app.Spec.Instances.Exactly != nil:
						instances = strconv.FormatInt(int64(*app.Spec.Instances.Exactly), 10)
					case app.Spec.Instances.Min == nil && app.Spec.Instances.Max == nil:
						instances = "?"
					case app.Spec.Instances.Min != nil && app.Spec.Instances.Max != nil:
						instances = fmt.Sprintf(
							"%d - %d",
							*app.Spec.Instances.Min,
							*app.Spec.Instances.Max,
						)
					case app.Spec.Instances.Max != nil:
						instances = fmt.Sprintf(
							"0 - %d",
							*app.Spec.Instances.Max,
						)
					case app.Spec.Instances.Min != nil:
						instances = fmt.Sprintf(
							"%d - ∞",
							*app.Spec.Instances.Min,
						)
					}

					// Memory & Disk
					// TODO(#431): Persistent disks
					var memory, disk string
					if containers := app.Spec.Template.Spec.Containers; len(containers) > 0 {
						if mem, ok := containers[0].Resources.Requests[corev1.ResourceMemory]; ok {
							memory = mem.String()
						}

						if d, ok := containers[0].Resources.Requests[corev1.ResourceEphemeralStorage]; ok {
							disk = d.String()
						}
					}

					// URL
					var urls []string
					for _, route := range app.Spec.Routes {
						urls = append(urls, route.String())
					}

					if app.Name == "" {
						continue
					}

					kfApp := apps.NewFromApp(&app)

					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						app.Name,
						requestedState,
						instances,
						memory,
						disk,
						strings.Join(urls, ", "),
						kfApp.GetClusterURL(),
					)
				}
			})

			return nil
		},
	}
}
