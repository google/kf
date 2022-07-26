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

package transformer

import (
	"context"
	"os"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/google/go-cmp/cmp"
)

func TestGetPodAnnotationsFromEnv(t *testing.T) {
	testCases := map[string]struct {
		requiredEnvVars map[string]string
		actualEnvVars   map[string]string
		wanted          map[string]string
		wantErr         bool
	}{
		"missing": {
			requiredEnvVars: map[string]string{
				"something": "something",
			},
			actualEnvVars: map[string]string{},
			wantErr:       true,
		}, "all present": {
			requiredEnvVars: map[string]string{
				"foo": "bar",
				"qux": "gerb",
			},
			actualEnvVars: map[string]string{
				"foo": "key1",
				"qux": "keyA",
			},
			wanted: map[string]string{
				"bar":  "key1",
				"gerb": "keyA",
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			for k, v := range tc.actualEnvVars {
				orig := os.Getenv(k)
				os.Setenv(k, v)
				defer os.Setenv(k, orig)
			}
			defer TestOnlyChangeRequiredAnnotationsFromEnvVars(tc.requiredEnvVars)()
			podAnnotations, err := GetPodAnnotationsFromEnv()
			if tc.wantErr {
				if err == nil {
					t.Fatal("Expected an error, didn't receive one")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wanted, podAnnotations); diff != "" {
				t.Fatalf("Unexpected podAnnotations (-want +got): %s", diff)
			}
		})
	}
}

var annotations = map[string]string{
	"john":    "scalzi",
	"patrick": "rothfuss",
}

func assertEquivalent(t *testing.T, orig, want interface{}) {
	a := &Annotation{
		PodAnnotations: annotations,
	}
	uOrig := &unstructured.Unstructured{}
	if err := scheme.Scheme.Convert(orig, uOrig, nil); err != nil {
		t.Fatalf("Unable to convert orig to unstructured: %v", err)
	}
	if err := a.Transform(context.Background())(uOrig); err != nil {
		t.Fatalf("Error transforming: %v", err)
	}
	uWant := &unstructured.Unstructured{}
	if err := scheme.Scheme.Convert(want, uWant, nil); err != nil {
		t.Fatalf("Unable to convert orig to unstructured: %v", err)
	}
	uWant.SetCreationTimestamp(metav1.Time{})
	if diff := cmp.Diff(uWant, uOrig); diff != "" {
		t.Fatalf("Unexpected changes (-want +got): %s", diff)
	}
}

func TestTransformPod(t *testing.T) {
	pod := &v1.Pod{}
	desired := pod.DeepCopy()
	desired.Annotations = annotations
	assertEquivalent(t, pod, desired)
}

func TestTransformDeployment(t *testing.T) {
	deployment := &appsv1.Deployment{}
	desired := deployment.DeepCopy()
	desired.Spec.Template.Annotations = annotations
	assertEquivalent(t, deployment, desired)
}

func TestTransformDaemonSet(t *testing.T) {
	daemonSet := &appsv1.DaemonSet{}
	desired := daemonSet.DeepCopy()
	desired.Spec.Template.Annotations = annotations
	assertEquivalent(t, daemonSet, desired)
}

func TestTransformStatefulSet(t *testing.T) {
	statefulSet := &appsv1.StatefulSet{}
	desired := statefulSet.DeepCopy()
	desired.Spec.Template.Annotations = annotations
	assertEquivalent(t, statefulSet, desired)
}

func TestTransformJob(t *testing.T) {
	job := &batchv1.Job{}
	desired := job.DeepCopy()
	desired.Spec.Template.Annotations = annotations
	assertEquivalent(t, job, desired)
}
