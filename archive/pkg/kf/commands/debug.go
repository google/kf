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
	"fmt"
	"io"
	"runtime"

	"github.com/google/kf/pkg/dockerutil"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/describe"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "k8s.io/client-go/kubernetes"
)

// NewDebugCommand creates a command that prints debugging information.
func NewDebugCommand(p *config.KfParams, kubernetes k8sclient.Interface) *cobra.Command {
	return &cobra.Command{
		Use:     "debug",
		Short:   "Show debugging information useful for filing a bug report",
		Example: `  kf debug`,
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()

			debugRuntime(w)
			debugKfParams(w, p)
			debugVersion(w, kubernetes)
			dockerutil.DescribeDefaultConfig(w)

			return nil
		},
	}
}

func debugKfParams(w io.Writer, p *config.KfParams) {
	describe.SectionWriter(w, "KfParams", func(w io.Writer) {
		if p == nil {
			fmt.Fprintln(w, "Params are nil")
			return
		}

		fmt.Fprintf(w, "Config Path:\t%s\n", p.Config)
		fmt.Fprintf(w, "Target Space:\t%s\n", p.Namespace)
		fmt.Fprintf(w, "Kubeconfig:\t%s\n", p.KubeCfgFile)
	})
}

func debugRuntime(w io.Writer) {
	describe.SectionWriter(w, "Runtime", func(w io.Writer) {
		fmt.Fprintf(w, "Go Version:\t%s\n", runtime.Version())
		fmt.Fprintf(w, "Compiler:\t%s\n", runtime.Compiler)
		fmt.Fprintf(w, "Arch:\t%s\n", runtime.GOARCH)
		fmt.Fprintf(w, "OS:\t%s\n", runtime.GOOS)
	})
}

func debugVersion(w io.Writer, kubernetes k8sclient.Interface) {
	describe.SectionWriter(w, "Version", func(w io.Writer) {
		fmt.Fprintf(w, "Kf Client:\t%s\n", Version)

		if version, err := kubernetes.Discovery().ServerVersion(); err != nil {
			fmt.Fprintf(w, "Server version error:\t%s\n", err)
		} else {
			fmt.Fprintf(w, "Server version:\t%v\n", version)
		}

		namespaceLabel(w, kubernetes, "kf", "app.kubernetes.io/version")
		namespaceLabel(w, kubernetes, "knative-serving", "serving.knative.dev/release")
	})
}

func namespaceLabel(w io.Writer, kubernetes k8sclient.Interface, namespace, label string) {
	if ns, err := kubernetes.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{}); err != nil {
		fmt.Fprintf(w, "%s[%q]:\terror: %s\n", namespace, label, err)
	} else {
		fmt.Fprintf(w, "%s[%q]:\t%v\n", namespace, label, ns.Labels[label])
	}
}
