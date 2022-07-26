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

package builds

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/fatih/color"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	tektoncli "github.com/google/kf/v2/third_party/tektoncd-cli/pkg/cli"
	"github.com/google/kf/v2/third_party/tektoncd-cli/pkg/cmd/taskrun"
	"github.com/google/kf/v2/third_party/tektoncd-cli/pkg/helper/pods"
	tektoninjection "github.com/tektoncd/pipeline/pkg/client/injection/client"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/injection/clients/dynamicclient"
)

// BuildStatus gets the status of the given build.
// Complete will be set to true if the build has completed (or doesn't exist).
// Error will be set if the build completed with an error (or doesn't exist).
// A successful result is one that completed and error is nil.
func BuildStatus(build v1alpha1.Build) (finished bool, err error) {
	condition := build.Status.GetCondition(v1alpha1.BuildConditionSucceeded)
	if condition == nil {
		// no success condition means the build hasn't propigated yet
		return false, nil
	}

	switch condition.Status {
	case corev1.ConditionTrue:
		return true, nil

	case corev1.ConditionFalse:
		return true, fmt.Errorf("build failed for reason: %s with message: %s", condition.Reason, condition.Message)

	default: // the build is in a transition state
		return false, nil
	}
}

// BuildTailerFunc converts a func into a BuildTailer.
type BuildTailerFunc func(ctx context.Context, out io.Writer, buildName, namespace string) error

// Tail implements BuildTailer.
func (f BuildTailerFunc) Tail(ctx context.Context, out io.Writer, buildName, namespace string) error {
	return f(ctx, out, buildName, namespace)
}

// TektonLoggingShim implements BuildTailer for taskrun.LogReader
// (https://godoc.org/github.com/tektoncd/cli/pkg/cmd/taskrun#LogReader).
//
// TODO: To keep the transition to Tekton small, we'll implement a function
// here that implements the interface that knative-build expected. We should
// obviously move away from this at some point.
func TektonLoggingShim(ki kubernetes.Interface) BuildTailer {
	return BuildTailerFunc(func(ctx context.Context, out io.Writer, buildName, namespace string) error {
		ti := tektoninjection.Get(ctx)

		// There is a race condition where a TaskRun might not have a Pod
		// scheduled for it yet. When this happens, the LogReader bails with
		// an error. So we have to wait for the TaskRun to have the pod on its
		// Status.
		if err := waitForTaskRunPod(ctx, namespace, buildName, out); err != nil {
			return err
		}

		reader := taskrun.LogReader{
			Ns:            namespace,
			Run:           buildName,
			AllSteps:      true,
			DenylistSteps: []string{"istio-proxy", "istio-init", "sidecar-server"},
			Follow:        true,
			Streamer:      pods.NewStream,
			Clients: &tektoncli.Clients{
				Tekton: ti,
				Kube:   ki,
			},
		}
		logs, errs, err := reader.Read(ctx)
		if err != nil {
			return err
		}

		fgGreen := color.New(color.FgGreen)

		for {
			select {
			case err := <-errs:
				return err
			case log, ok := <-logs:
				if !ok {
					return nil
				}
				prefix := fgGreen.Sprintf("[%s/%s]", log.Task, log.Step)
				fmt.Fprintln(out, prefix, log.Log)
			}
		}
	})
}

func waitForTaskRunPod(ctx context.Context, namespace, buildName string, out io.Writer) error {
	client := dynamicclient.Get(ctx).Resource(schema.GroupVersionResource{
		Group:    "tekton.dev",
		Version:  "v1beta1",
		Resource: "taskruns",
	}).Namespace(namespace)

	timer := time.NewTimer(0)
	for {
		select {
		case <-timer.C:
			tr, err := client.Get(ctx, buildName, metav1.GetOptions{})
			if apierrs.IsNotFound(err) {
				timer.Reset(100 * time.Millisecond)
				continue
			} else if err != nil {
				return fmt.Errorf("failed to get TaskRun: %v", err)
			}
			podName, _, err := unstructured.NestedString(tr.Object, "status", "podName")
			if err != nil {
				return fmt.Errorf("failed to parse TaskRun: %v", err)
			}

			if podName != "" {
				// Found a Pod, return.
				fmt.Fprintln(out, "Reading build logs from Pod", podName)
				return nil
			}

			timer.Reset(100 * time.Millisecond)
		case <-ctx.Done():
			return ctx.Err()
		}
	}

}
