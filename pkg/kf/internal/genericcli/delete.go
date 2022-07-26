// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package genericcli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	cliutil "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/injection/clients/dynamicclient"
	"knative.dev/pkg/logging"
)

// NewDeleteByNameCommand creates a delete command for Kubernetes objects that
// works by the name of the object.
func NewDeleteByNameCommand(t Type, p *config.KfParams, opts ...DeleteByNameOption) *cobra.Command {
	var (
		async cliutil.AsyncFlags
		retry cliutil.RetryFlags
	)

	friendlyType := t.FriendlyName()
	lowerFriendly := strings.ToLower(friendlyType)

	options := DeleteByNameOptions{
		WithDeleteByNameCommandName(fmt.Sprintf("delete-%s", lowerFriendly)),
	}.Extend(opts)

	short := fmt.Sprintf("Delete the %s with the given name.", friendlyType)
	if t.Namespaced() {
		short = fmt.Sprintf("Delete the %s with the given name in the targeted Space.", friendlyType)
	}

	long := fmt.Sprintf(`Deletes the %[1]s with the given name and wait for it to be deleted.

	Kubernetes will delete the %[1]s once all child resources it owns have been deleted.
	Deletion may take a long time if any of the following conditions are true:

	* There are many child objects.
	* There are finalizers on the object preventing deletion.
	* The cluster is in an unhealthy state.
	`, friendlyType)

	cmd := &cobra.Command{
		Use:               fmt.Sprintf("%s NAME", options.CommandName()),
		Aliases:           options.Aliases(),
		Short:             short,
		Long:              cliutil.JoinHeredoc(long, options.AdditionalLongText()),
		Example:           fmt.Sprintf("kf %s my-%s", options.CommandName(), lowerFriendly),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: ValidArgsFunction(t, p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if t.Namespaced() {
				if err := p.ValidateSpaceTargeted(); err != nil {
					return err
				}
			}

			resourceName := args[0]
			logger := logging.FromContext(ctx)

			// Print status messages to stderr so stdout is syntatically valid output
			// if the user wanted JSON, YAML, etc.
			if t.Namespaced() {
				logger.Infof("Deleting %s %q in Space %q", friendlyType, resourceName, p.Space)
			} else {
				logger.Infof("Deleting %s %q", friendlyType, resourceName)
			}

			client := GetResourceInterface(ctx, t, dynamicclient.Get(ctx), p.Space)

			deleteOptions := metav1.DeleteOptions{}
			policy := options.PropagationPolicy()
			if len(policy) > 0 {
				deleteOptions.PropagationPolicy = &policy
			} else {
				policy = metav1.DeletePropagationForeground
				deleteOptions.PropagationPolicy = &policy
			}

			if err := retry.Retry(func() error {
				return client.Delete(ctx, resourceName, deleteOptions)
			}); err != nil {
				return err
			}

			if async.IsAsync() {
				return nil
			}

			logger.Info("Waiting for deletion...")

			tick := time.Tick(1 * time.Second)
			for {
				_, err := client.Get(ctx, resourceName, metav1.GetOptions{})
				if err != nil {
					if apierrors.IsNotFound(err) {
						return nil // success
					}

					return fmt.Errorf("waiting for deletion failed: %s", err) // failure
				}

				select {
				case <-tick:
					// repeat instance check
				case <-ctx.Done():
					return errors.New("timed out while waiting for resource to delete")
				}
			}
		},
	}

	async.Add(cmd)
	retry.AddRetryForK8sPropagation(cmd)

	return cmd
}
