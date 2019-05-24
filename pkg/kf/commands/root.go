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
	"github.com/spf13/pflag"
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
	rootCmd.PersistentFlags().StringVar(&p.Namespace, "namespace", "default", "kubernetes namespace")
	rootCmd.PersistentFlags().BoolVarP(&p.Verbose, "verbose", "v", false, "make the operation more talkative")
	rootCmd.PersistentFlags().Bool(builtInLong, false, "do not use remote override commands from CRD")

	commands := map[string]*cobra.Command{
		// App interaction
		"push":   InjectPush(p),
		"delete": InjectDelete(p),
		"apps":   InjectApps(p),
		"proxy":  InjectProxy(p),

		// Environment Variables
		"env":       InjectEnv(p),
		"set-env":   InjectSetEnv(p),
		"unset-env": InjectUnsetEnv(p),

		// Services
		"create-service": InjectCreateService(p),
		"delete-service": InjectDeleteService(p),
		"service":        InjectGetService(p),
		"services":       InjectListServices(p),
		"marketplace":    InjectMarketplace(p),

		// Service Bindings
		"bind-service":  InjectBindingService(p),
		"bindings":      InjectListBindings(p),
		"ubind-service": InjectUnbindService(p),
		"vcap-services": InjectVcapServices(p),

		// Buildpacks
		"buildpacks":        InjectBuildpacks(p),
		"upload-buildpacks": InjectUploadBuildpacks(p),

		// Spaces
		"spaces":       InjectSpaces(p),
		"create-space": InjectCreateSpace(p),
		"delete-space": InjectDeleteSpace(p),

		// DoctorTests are run in the order they're defined in this list.
		// Tests will stop as soon as one of these top-level tests fails so they
		// should be ordered in a logical way e.g. testing apps should come after
		// testing the cluster because if the cluster isn't working then all the
		// app tests will fail.
		"doctor": doctor.NewDoctorCommand(p, []doctor.DoctorTest{
			{Name: "cluster", Test: pkgdoctor.NewClusterDiagnostic(config.GetKubernetes(p))},
		}),
	}

	if !builtIn() {
		// Override the base commands with ones from the CRD.
		overrides, err := InjectOverrider(p).FetchCommandOverrides()
		if err != nil {
			// nop
			// Proceed without any overrides
		}

		for k, v := range overrides {
			commands[k] = v
		}
	}

	for _, cmd := range commands {
		rootCmd.AddCommand(cmd)
	}

	return rootCmd
}

const (
	builtInLong = "built-in"
)

// builtIn reads a FlagSet and looks for the "built-in" flag and returns its
// value if found. It is necessary to parse the flag ahead of time as it
// determines which commands are loaded and therefore needs to be parsed
// earlier than pflags normally does.
func builtIn() bool {
	flags := pflag.NewFlagSet("built-in", pflag.ContinueOnError)
	flags.Usage = func() {
		// NOP - We don't want a --help to show anything for this FlagSet. The
		// main FlagSet will take care of it.
	}
	result := flags.Bool(builtInLong, false, "")

	// We are only configured to look for --built-in. So when we encounter
	// other flags, we want to keep going.
	flags.ParseErrorsWhitelist.UnknownFlags = true
	flags.Parse(os.Args)
	return *result
}
