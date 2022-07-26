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
	"context"
	"fmt"
	"io"
	"runtime"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/dockerutil"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/describe"
	"github.com/google/kf/v2/pkg/kf/doctor"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	k8sclient "k8s.io/client-go/kubernetes"
)

// NewDebugCommand creates a command that prints debugging information.
func NewDebugCommand(p *config.KfParams, kubernetes k8sclient.Interface) *cobra.Command {
	return &cobra.Command{
		Use:     "debug",
		Short:   "Print debugging information useful for filing a bug report.",
		Example: `  kf debug`,
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()

			debugRuntime(w)
			debugKfParams(w, p)
			ns := getKfNamespace(cmd.Context(), kubernetes)
			debugVersion(w, kubernetes, ns)
			dockerutil.DescribeDefaultConfig(w)
			debugServerComponents(cmd.Context(), w, kubernetes)

			return nil
		},
		SilenceUsage: true,
	}
}

func debugKfParams(w io.Writer, p *config.KfParams) {
	describe.SectionWriter(w, "KfParams", func(w io.Writer) {
		if p == nil {
			fmt.Fprintln(w, "Params are nil")
			return
		}

		fmt.Fprintf(w, "Config Path:\t%s\n", p.Config)
		fmt.Fprintf(w, "Target Space:\t%s\n", p.Space)
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

func debugVersion(w io.Writer, kubernetes k8sclient.Interface, ns *v1.Namespace) {
	describe.SectionWriter(w, "Version", func(w io.Writer) {
		fmt.Fprintf(w, "Kf Client:\t%s\n", Version)

		if ns != nil {
			fmt.Fprintf(w, "Kf Server:\t%s\n", ns.Labels[v1alpha1.VersionLabel])
		}

		if version, err := kubernetes.Discovery().ServerVersion(); err != nil {
			fmt.Fprintf(w, "Kubernetes version error:\t%s\n", err)
		} else {
			fmt.Fprintf(w, "Kubernetes version:\t%v\n", version)
		}
	})
}

func debugServerComponents(ctx context.Context, w io.Writer, kubernetes k8sclient.Interface) {
	// namespaces contains the namespaces we're interested in, we bootstrap
	// with two namespaces we know should exist so if the user doesn't
	// have permissions to autodetect we'll still try something.
	namespaces := sets.NewString("kf", "kube-system")
	namespaces.Insert(doctor.DiscoverControllerNamespaces(ctx, kubernetes)...)

	describe.SectionWriter(w, "Cluster Components", func(w io.Writer) {
		describe.TabbedWriter(w, func(w io.Writer) {
			fmt.Fprintln(w, "Namespace\tResource\tImages")

			for _, ns := range namespaces.List() {
				// Deployments
				{
					resourceList, err := kubernetes.AppsV1().
						Deployments(ns).
						List(ctx, metav1.ListOptions{})

					if err != nil {
						fmt.Fprintf(w, "%s\tDeployment\tERROR: %v\n", ns, err)
					} else {
						for _, resource := range resourceList.Items {
							fmt.Fprintf(
								w,
								"%s\tDeployment/%s\t%v\n",
								ns,
								resource.Name,
								gatherContainerImages(resource.Spec.Template),
							)
						}
					}
				}

				// DaemonSets
				{
					resourceList, err := kubernetes.AppsV1().
						DaemonSets(ns).
						List(ctx, metav1.ListOptions{})

					if err != nil {
						fmt.Fprintf(w, "%s\tDaemonSet\tERROR: %v\n", ns, err)
					} else {
						for _, resource := range resourceList.Items {
							fmt.Fprintf(
								w,
								"%s\tDaemonSet/%s\t%v\n",
								ns,
								resource.Name,
								gatherContainerImages(resource.Spec.Template),
							)
						}
					}
				}
			}
		})
	})
}

// gatherContainerImages gets the set of images for each container on a PodTemplateSpec.
func gatherContainerImages(podTemplate corev1.PodTemplateSpec) []string {
	images := sets.NewString()

	for _, container := range podTemplate.Spec.Containers {
		images.Insert(container.Image)
	}

	return images.List()
}
