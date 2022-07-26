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

package spaces

import (
	"fmt"

	"github.com/google/kf/v2/pkg/kf/algorithms"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	resources "github.com/google/kf/v2/pkg/reconciler/space/resources"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/logging"
)

// NewUnsetSpaceRoleCommand allows users to un-assign role to a subject.
func NewUnsetSpaceRoleCommand(p *config.KfParams, kubeClient kubernetes.Interface) *cobra.Command {
	var subjectKind string
	cmd := &cobra.Command{
		Use:   "unset-space-role SUBJECT_NAME ROLE",
		Short: "Unassign a Role to a Subject.",
		Example: `
		# Unassign a User to a Role
		kf unset-space-role john@example.com SpaceDeveloper

		# Unassign a Group to a Role
		kf unset-space-role my-group SpaceAuditor -t Group

		# Unassign a ServiceAccount to a Role
		kf unset-space-role my-sa SpaceAuditor -t ServiceAccount
		`,
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			username := args[0]
			role := args[1]

			roleBindingName := resources.GetRoleBindingName(role, p.Space)
			if len(roleBindingName) == 0 {
				return fmt.Errorf("Role %q does not exist", role)
			}

			roleBinding, err := kubeClient.RbacV1().RoleBindings(p.Space).Get(ctx, roleBindingName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("error getting Role %q in Space %q", roleBindingName, p.Space)
			}

			if found, index := algorithms.Subjects(roleBinding.Subjects).Contains(username, subjectKind); found {
				roleBinding.Subjects = append(roleBinding.Subjects[:index], roleBinding.Subjects[index+1:]...)
				if _, err := kubeClient.RbacV1().RoleBindings(p.Space).Update(ctx, roleBinding, metav1.UpdateOptions{}); err != nil {
					return fmt.Errorf("error removing %q (%q) from Role %q", username, subjectKind, role)
				}
				logging.FromContext(ctx).Infof("%q (%s) is removed from %q Role", username, subjectKind, role)
				return nil
			}

			return fmt.Errorf("%q (%s) was not found at Role %q", username, subjectKind, role)
		},
	}

	cmd.Flags().StringVarP(
		&subjectKind,
		"type",
		"t",
		"User",
		"Type of subject, valid values are Group|ServiceAccount|User(default).",
	)

	return cmd
}
