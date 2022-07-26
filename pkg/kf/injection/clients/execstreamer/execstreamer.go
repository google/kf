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

package execstreamer

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	scheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/injection"
)

func init() {
	injection.Default.RegisterClient(withExecStreamer)
}

// Key is the key used to store the ExecStreamer on a context. It should not
// be used directly. It is exported for the fake package.
type Key struct{}

// ExecStreamer streams the result of running exec() in a pod.
type ExecStreamer interface {
	// Stream the results of running exec() in a pod.
	Stream(
		ctx context.Context,
		podSelector metav1.ListOptions,
		execOpts corev1.PodExecOptions,
		streamOpts remotecommand.StreamOptions,
	) error
}

// Get returns the ExecStreamer that is injected into the context.
func Get(ctx context.Context) ExecStreamer {
	return ctx.Value(Key{}).(ExecStreamer)
}

func withExecStreamer(ctx context.Context, cfg *rest.Config) context.Context {
	return context.WithValue(ctx, Key{}, &execStreamer{cfg: cfg})
}

type execStreamer struct {
	cfg *rest.Config
}

func (s *execStreamer) Stream(
	ctx context.Context,
	podSelector metav1.ListOptions,
	execOpts corev1.PodExecOptions,
	streamOpts remotecommand.StreamOptions,
) error {
	kubernetesClient := kubeclient.Get(ctx)
	space := injection.GetNamespaceScope(ctx)

	podList, err := kubernetesClient.
		CoreV1().
		Pods(space).
		List(ctx, podSelector)
	if err != nil {
		return err
	}

	podName, err := findExecablePod(podList.Items, execOpts.Container)
	if err != nil {
		return err
	}

	if streamOpts.Stderr != nil {
		fmt.Fprintf(
			streamOpts.Stderr,
			"Running %v on pod %s in space %s\n",
			execOpts.Command, podName, space,
		)
	}

	req := kubernetesClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(space).
		SubResource("exec").
		VersionedParams(&execOpts, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(s.cfg, "POST", req.URL())
	if err != nil {
		return err
	}

	return exec.Stream(streamOpts)
}

// findExecablePod gets a Pod that appears like it should handle an exec()
// this function prevents terminating Pods or Pods that are still coming up
// from running exec()
func findExecablePod(podList []corev1.Pod, containerName string) (name string, err error) {
	if len(podList) == 0 {
		return "", errors.New("no matching Pods found")
	}

	for _, pod := range podList {
		// Don't try to exec into deleting Pods.
		if pod.GetDeletionTimestamp() != nil {
			continue
		}

		if pod.Status.Phase == corev1.PodRunning {
			for _, container := range pod.Status.ContainerStatuses {
				if container.Ready &&
					container.Name == containerName &&
					container.State.Running != nil {
					return pod.Name, nil
				}
			}
		}
	}

	return "", errors.New("No running Pods found")
}
