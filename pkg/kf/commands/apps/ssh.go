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

package apps

import (
	"strings"

	dockerterm "github.com/docker/docker/pkg/term"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/completion"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/injection/clients/execstreamer"
	"github.com/google/kf/v2/third_party/k8s.io/kubectl/pkg/util/term"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/remotecommand"
)

const podPrefix = "pod/"

// NewSSHCommand sets up a mock SSH connection to an App that mimics cf SSH
func NewSSHCommand(p *config.KfParams) *cobra.Command {
	var (
		disableTTY bool
		command    []string
		container  string
	)

	cmd := &cobra.Command{
		Use:   "ssh APP_NAME",
		Short: "Open a shell on an App instance.",
		Example: `
		# Open a shell to a specific App
		kf ssh myapp

		# Open a shell to a specific Pod
		kf ssh pod/myapp-revhex-podhex

		# Start a different command with args
		kf ssh myapp -c /my/command -c arg1 -c arg2
		`,
		Args: cobra.ExactArgs(1),
		Long: `
		Opens a shell on an App instance using the Pod exec endpoint.

		This command mimics CF's SSH command by opening a connection to the
		Kubernetes control plane which spawns a process in a Pod.

		The command connects to an arbitrary Pod that matches the App's runtime
		labels. If you want a specific Pod, use the pod/<podname> notation.

		NOTE: Traffic is encrypted between the CLI and the control plane, and
		between the control plane and Pod. A malicious Kubernetes control plane
		could observe the traffic.
		`,
		ValidArgsFunction: completion.AppCompletionFn(p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}
			streamExec := execstreamer.Get(ctx)

			enableTTY := !disableTTY
			appName := args[0]

			podSelector := metav1.ListOptions{}
			if strings.HasPrefix(appName, podPrefix) {
				podName := strings.TrimPrefix(appName, podPrefix)
				podSelector.FieldSelector = fields.OneTermEqualSelector("metadata.name", podName).String()
			} else {
				podSelector.LabelSelector = metav1.FormatLabelSelector(labelSelectorForAppPods(appName, "app-server"))
			}

			execOpts := corev1.PodExecOptions{
				Container: container,
				Command:   command,
				Stdin:     true,
				Stdout:    true,
				Stderr:    true,
				TTY:       enableTTY,
			}

			t := term.TTY{
				Out: cmd.OutOrStdout(),
				In:  cmd.InOrStdin(),
				Raw: true,
			}

			sizeQueue := t.MonitorSize(t.GetSize())

			streamOpts := remotecommand.StreamOptions{
				Stdin:             cmd.InOrStdin(),
				Stdout:            cmd.OutOrStdout(),
				Stderr:            cmd.ErrOrStderr(),
				Tty:               enableTTY,
				TerminalSizeQueue: sizeQueue,
			}

			// Set up a TTY locally if it's enabled.
			if fd, isTerm := dockerterm.GetFdInfo(streamOpts.Stdin); isTerm && enableTTY {
				originalState, err := dockerterm.MakeRaw(fd)
				if err != nil {
					return err
				}

				defer dockerterm.RestoreTerminal(fd, originalState)
			}

			return streamExec.Stream(ctx, podSelector, execOpts, streamOpts)
		},
	}

	cmd.Flags().StringArrayVarP(
		&command,
		"command",
		"c",
		[]string{"/bin/bash"},
		"Command to run for the shell. Subsequent definitions will be used as args.",
	)

	cmd.Flags().StringVar(
		&container,
		"container",
		v1alpha1.DefaultUserContainerName,
		"Container to start the command in.",
	)

	cmd.Flags().BoolVarP(
		&disableTTY,
		"disable-pseudo-tty",
		"T",
		false,
		"Don't use a TTY when executing.",
	)

	return cmd
}

func labelSelectorForAppPods(appName, component string) *metav1.LabelSelector {
	appLabels := v1alpha1.AppComponentLabels(appName, component)
	delete(appLabels, v1alpha1.ManagedByLabel)
	labelSelector := metav1.SetAsLabelSelector(appLabels)
	// Due to https://github.com/tektoncd/pipeline/issues/8827 Kf pods wil have
	// managed by label value hardcoded to "tekton-pipelines". Previous
	// value "kf" left for backward compatibility
	managedByRequirement := metav1.LabelSelectorRequirement{Key: v1alpha1.ManagedByLabel,
		Operator: metav1.LabelSelectorOpIn, Values: []string{v1alpha1.ManagedByKfValue, v1alpha1.ManagedByTektonValue}}
	labelSelector.MatchExpressions = []metav1.LabelSelectorRequirement{managedByRequirement}
	return labelSelector
}
