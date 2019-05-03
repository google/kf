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

package commands

import (
	"os"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/doctor"
	pkgdoctor "github.com/GoogleCloudPlatform/kf/pkg/kf/doctor"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

// NewKfCommand creates the root kf command.
func NewKfCommand() *cobra.Command {
	p := &config.KfParams{
		Output: os.Stdout,
	}

	var rootCmd = &cobra.Command{
		Use:   "kf",
		Short: "kf is like cf for Knative",
		Long: `kf is like cf for Knative

kf supports the following sub-commands:

Apps:
  kf push
  kf delete <app>
  kf apps

Services:
  kf marketplace
  kf create-service
  kf delete-service
  kf service <instance-name>
  kf services
	kf bindings
	kf bind-service <app> <instance-name>
	kf unbind-service <app> <instance-name>

You can get more info by adding the --help flag to any sub-command.
`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	rootCmd.PersistentFlags().StringVar(&p.KubeCfgFile, "kubeconfig", "", "kubectl config file (default is $HOME/.kube/config)")
	rootCmd.PersistentFlags().StringVar(&p.Namespace, "namespace", "default", "namespace")

	// App interaction
	rootCmd.AddCommand(injectPush(p))
	rootCmd.AddCommand(injectDelete(p))
	rootCmd.AddCommand(injectApps(p))
	rootCmd.AddCommand(injectProxy(p))

	// Environment Variables
	rootCmd.AddCommand(injectEnv(p))
	rootCmd.AddCommand(injectSetEnv(p))
	rootCmd.AddCommand(injectUnsetEnv(p))

	// Services
	rootCmd.AddCommand(injectCreateService(p))
	rootCmd.AddCommand(injectDeleteService(p))
	rootCmd.AddCommand(injectGetService(p))
	rootCmd.AddCommand(injectListServices(p))
	rootCmd.AddCommand(injectMarketplace(p))

	// Service Bindings
	rootCmd.AddCommand(injectBindingService(p))
	rootCmd.AddCommand(injectListBindings(p))
	rootCmd.AddCommand(injectUnbindService(p))
	rootCmd.AddCommand(injectVcapServices(p))

	// Buildpacks
	rootCmd.AddCommand(injectBuildpacks(p))
	rootCmd.AddCommand(injectUploadBuildpacks(p))

	// DoctorTests are run in the order they're defined in this list.
	// Tests will stop as soon as one of these top-level tests fails so they
	// should be ordered in a logical way e.g. testing apps should come after
	// testing the cluster because if the cluster isn't working then all the
	// app tests will fail.
	tests := []doctor.DoctorTest{
		{Name: "cluster", Test: pkgdoctor.NewClusterDiagnostic(config.GetKubernetes(p))},
	}
	rootCmd.AddCommand(doctor.NewDoctorCommand(p, tests))

	return rootCmd
}
