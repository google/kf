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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/logging"
)

// NewSetSpaceRoleCommand allows users to assign role to a subject.
func NewSetSpaceRoleCommand(p *config.KfParams, kubeClient kubernetes.Interface) *cobra.Command {
	var subjectKind string
	cmd := &cobra.Command{
		Use:   "set-space-role SUBJECT_NAME ROLE",
		Short: "Assgin a Role to a Subject (User|Group|ServiceAccount).",
		Example: `
		# Assign a User to a Role
		kf set-space-role john@example.com SpaceDeveloper

		# Assign a Group to a Role
		kf set-space-role my-group SpaceAuditor -t Group

		# Assign a ServiceAccount to a Role
		kf set-space-role my-sa SpaceAuditor -t ServiceAccount
		`,
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logging.FromContext(ctx)

			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			username := args[0]
			role := args[1]

			roleBindingName := resources.GetRoleBindingName(role, p.Space)
			if len(roleBindingName) == 0 {
				return fmt.Errorf("Role %q does not exist", role)
			}

			roleBinding, err := kubeClient.RbacV1().RoleBindings(p.Space).Get(cmd.Context(), roleBindingName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("error getting Role %q in Space %q", roleBindingName, p.Space)
			}

			if found, _ := algorithms.Subjects(roleBinding.Subjects).Contains(username, subjectKind); found {
				logger.Warnf("%q (%q) is already assigned %q Role", username, subjectKind, role)
				return nil
			}

			roleBinding.Subjects = append(roleBinding.Subjects, rbacv1.Subject{Kind: subjectKind, Name: username})

			if _, err := kubeClient.RbacV1().RoleBindings(p.Space).Update(cmd.Context(), roleBinding, metav1.UpdateOptions{}); err != nil {
				return fmt.Errorf("error assigning %q to %q", role, username)
			}

			logger.Infof("%q (%s) is assigned %q Role", username, subjectKind, role)
			return nil
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
