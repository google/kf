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

			if err := debugRuntime(w); err != nil {
				return err
			}
			if err := debugKfParams(w, p); err != nil {
				return err
			}
			if err := debugVersion(w, kubernetes); err != nil {
				return err
			}
			if err := dockerutil.DescribeDefaultConfig(w); err != nil {
				return err
			}

			return nil
		},
	}
}

func debugKfParams(w io.Writer, p *config.KfParams) error {
	return describe.SectionWriter(w, "KfParams", func(w io.Writer) error {
		if p == nil {
			if _, err := fmt.Fprintln(w, "Params are nil"); err != nil {
				return err
			}

			return nil
		}

		if _, err := fmt.Fprintf(w, "Config Path:\t%s\n", p.Config); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "Target Space:\t%s\n", p.Namespace); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "Kubeconfig:\t%s\n", p.KubeCfgFile); err != nil {
			return err
		}

		return nil
	})
}

func debugRuntime(w io.Writer) error {
	return describe.SectionWriter(w, "Runtime", func(w io.Writer) error {
		if _, err := fmt.Fprintf(w, "Go Version:\t%s\n", runtime.Version()); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "Compiler:\t%s\n", runtime.Compiler); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "Arch:\t%s\n", runtime.GOARCH); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "OS:\t%s\n", runtime.GOOS); err != nil {
			return err
		}

		return nil
	})
}

func debugVersion(w io.Writer, kubernetes k8sclient.Interface) error {
	return describe.SectionWriter(w, "Version", func(w io.Writer) error {
		if _, err := fmt.Fprintf(w, "Kf Client:\t%s\n", Version); err != nil {
			return err
		}

		if version, err := kubernetes.Discovery().ServerVersion(); err != nil {
			if _, err := fmt.Fprintf(w, "Server version error:\t%s\n", err); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(w, "Server version:\t%v\n", version); err != nil {
				return err
			}
		}

		if err := namespaceLabel(w, kubernetes, "kf", "app.kubernetes.io/version"); err != nil {
			return err
		}
		if err := namespaceLabel(w, kubernetes, "knative-serving", "serving.knative.dev/release"); err != nil {
			return err
		}

		return nil
	})
}

func namespaceLabel(w io.Writer, kubernetes k8sclient.Interface, namespace, label string) error {
	if ns, err := kubernetes.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{}); err != nil {
		if _, err := fmt.Fprintf(w, "%s[%q]:\terror: %s\n", namespace, label, err); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(w, "%s[%q]:\t%v\n", namespace, label, ns.Labels[label]); err != nil {
			return err
		}
	}

	return nil
}
