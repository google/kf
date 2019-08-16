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
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// NewVcapServicesCommand allows users to bind apps to service instances.
func NewVcapServicesCommand(
	p *config.KfParams,
	client kubernetes.Interface,
) *cobra.Command {

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

			secret, err := client.
				CoreV1().
				Secrets(p.Namespace).
				Get(fmt.Sprintf("kf-injected-envs-%s", appName), metav1.GetOptions{})
			if err != nil {
				return err
			}

			vcapServices, ok := secret.Data["VCAP_SERVICES"]
			if !ok {
				return errors.New("VCAP_SERVICES does not exist")
			}

			decoded, err := base64.StdEncoding.DecodeString(string(vcapServices))
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(decoded))
			return nil
		},
	}

	return cmd
}
