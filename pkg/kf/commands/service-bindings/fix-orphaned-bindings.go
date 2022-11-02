// Copyright 2022 Google LLC
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

	"github.com/fatih/color"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/apps"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/serviceinstancebindings"
	"github.com/spf13/cobra"
	"knative.dev/pkg/kmeta"
)

// NewDeleteOrphanedRoutesCommand creates a command to delete orphaned routes.
func NewFixOrphanedBindingsCommand(
	p *config.KfParams,
	bindingsClient serviceinstancebindings.Client,
	appsClient apps.Client,
) *cobra.Command {
	var (
		dryRun = false
	)

	cmd := &cobra.Command{
		Use:   "fix-orphaned-bindings",
		Short: "Fix bindings without an app owner in a space.",
		Long:  `Fix bindings with a missing owner reference to an App.`,
		Example: `
		# Identify broken bindings in the targeted space.
		kf fix-orphaned-bindings

		# Fix bindings in the targeted space.
		kf fix-orphaned-bindings --dry-run=false
		`,
		Args:         cobra.ExactArgs(0),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			out := cmd.OutOrStdout()

			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			fmt.Fprintf(out, "Finding orphaned bindings in Space: %s\n", p.Space)

			bindings, err := bindingsClient.List(ctx, p.Space)
			if err != nil {
				return fmt.Errorf("failed to list Bindings: %v", err)
			}

			apps, err := appsClient.List(ctx, p.Space)
			if err != nil {
				return fmt.Errorf("failed to list Apps: %v", err)
			}
			appLookupTable := make(map[string]v1alpha1.App)
			for _, a := range apps {
				appLookupTable[a.Name] = a
			}

			cmdColor := color.New(color.FgCyan)

			deletedBindingsCount := 0
			ownersFixedCount := 0
			for _, b := range bindings {
				appRef := b.Spec.App
				if appRef == nil {
					// Case: binding refers to something other than an app.
					continue
				}

				app, ok := appLookupTable[appRef.Name]
				if !ok {
					// Case: App has been deleted, remove binding.
					fmt.Fprintf(out, "Deleting binding %q which refers to a missing app %q\n", b.Name, appRef.Name)
					deletedBindingsCount++
					fmt.Fprintf(out, cmdColor.Sprintf("  kubectl delete serviceinstancebindings -n %q %q", p.Space, b.Name))
					fmt.Fprintln(out)

					if !dryRun {
						if err := bindingsClient.Delete(ctx, p.Space, b.Name); err != nil {
							return fmt.Errorf("failed to delete binding: %v", err)
						}
					}

					continue
				}

				if ownedByApp(b, app) {
					// Case: Binding is valid
					continue
				}

				// Case: Missing owner
				fmt.Fprintf(out, "Binding %q is missing an owner reference to app %q\n", b.Name, app.Name)
				ownersFixedCount++
				ownerRef := *kmeta.NewControllerRef(&app)
				jsonOwnerRef, err := json.Marshal(ownerRef)
				if err != nil {
					return err
				}

				fmt.Fprintf(out, cmdColor.Sprintf(
					`  kubectl patch serviceinstancebinding -n %q %q --type=json -p='[{"op":"add", "path":"/metadata/ownerReferences", "value":[%s] }]'`,
					p.Space,
					b.Name,
					jsonOwnerRef,
				))
				fmt.Fprintln(out)
				if !dryRun {
					if _, err := bindingsClient.Transform(ctx, p.Space, b.Name, func(binding *v1alpha1.ServiceInstanceBinding) error {
						binding.OwnerReferences = append(binding.OwnerReferences, ownerRef)
						return nil
					}); err != nil {
						return fmt.Errorf("failed to update binding: %v", err)
					}
				}
			}

			fmt.Fprintf(
				out,
				"Processed %d binding(s), %d owners fixed, %d deleted\n",
				len(bindings),
				ownersFixedCount,
				deletedBindingsCount,
			)
			if dryRun {
				fmt.Fprintf(
					out, color.New(color.FgHiYellow).Sprint("Run with --dry-run=false to apply."))
				fmt.Fprintln(out)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", true, "Run the command without applying changes.")

	return cmd
}

func ownedByApp(binding v1alpha1.ServiceInstanceBinding, app v1alpha1.App) bool {
	for _, owner := range binding.OwnerReferences {
		// Use UIDs for matching because versions might be mismatched.
		if owner.UID == app.UID {
			return true
		}
	}

	return false
}
