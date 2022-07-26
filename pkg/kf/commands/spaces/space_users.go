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
	"io"
	"sort"
	"strings"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/describe"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// NewSpaceUsersCommand allows users to list users for a Space.
func NewSpaceUsersCommand(p *config.KfParams, kubeClient kubernetes.Interface) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "space-users",
		Short:        "List users and their roles in a Space.",
		Example:      `kf space-users`,
		Args:         cobra.ExactArgs(0),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			roleBindingList, err := kubeClient.RbacV1().RoleBindings(p.Space).List(cmd.Context(), metav1.ListOptions{})

			if err != nil {
				return fmt.Errorf("error listing users in space")
			}

			userToRoleMap := make(map[string]Subject)

			roleBindings := roleBindingList.Items

			// Build reverse index of user -> roles
			for _, roleBinding := range roleBindings {
				for _, s := range roleBinding.Subjects {
					key := fmt.Sprintf("%s-%s", s.Name, s.Kind)
					if _, found := userToRoleMap[key]; !found {
						userToRoleMap[key] = Subject{
							Name:  s.Name,
							Kind:  s.Kind,
							Roles: []string{},
						}
					}
					subject := userToRoleMap[key]
					subject.Roles = append(subject.Roles, roleBinding.RoleRef.Name)
					userToRoleMap[key] = subject
				}
			}
			sortedSubjects := []Subject{}
			for _, subject := range userToRoleMap {
				// Sort roles by role name
				sort.Strings(subject.Roles)
				sortedSubjects = append(sortedSubjects, subject)
			}
			// Sort users by user name
			sort.Slice(sortedSubjects, func(i, j int) bool {
				return sortedSubjects[i].Name < sortedSubjects[j].Name
			})

			columnNames := []string{"Name", "Kind", "Roles"}
			describe.TabbedWriter(cmd.OutOrStdout(), func(w io.Writer) {
				fmt.Fprintln(w, strings.Join(columnNames, "\t"))
				for _, s := range sortedSubjects {
					fmt.Fprintf(w, "%s\t%s\t%v\n", s.Name, s.Kind, s.Roles)
				}
			})
			return nil
		},
	}

	return cmd
}

// Subject is the reverse index of User to Roles mapping.
type Subject struct {
	Name  string
	Kind  string
	Roles []string
}
