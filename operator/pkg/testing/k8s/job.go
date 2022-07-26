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

package k8s

import (
	mfTesting "kf-operator/pkg/testing/manifestival"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgotesting "k8s.io/client-go/testing"
)

// JobOption enables further configuration of a Job.
type JobOption func(*batchv1.Job)

// ManifestivalJob creates a Job owned by manifestival.
func ManifestivalJob(name string, jo ...JobOption) *batchv1.Job {
	obj := Job(name, jo...)
	mfTesting.SetManifestivalAnnotation(obj)
	mfTesting.SetLastApplied(obj)
	return obj
}

// Job creates a Job.
func Job(name string, jo ...JobOption) *batchv1.Job {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
	}
	for _, opt := range jo {
		opt(job)
	}
	return job
}

// WithBackoffLimit creates a JobOption with BackoffLimit limit.
func WithBackoffLimit(limit int32) JobOption {
	return func(job *batchv1.Job) {
		if job != nil {
			job.Spec.BackoffLimit = &limit
		}
	}
}

// WithActiveDeadlineSeconds creates a JobOption with ActiveDeadlineSeconds.
func WithActiveDeadlineSeconds(deadline int64) JobOption {
	return func(job *batchv1.Job) {
		if job != nil {
			job.Spec.ActiveDeadlineSeconds = &deadline
		}
	}
}

// WitJobContainer creates a DeploymentOption that updates Containers
// in Deployment.
func WitJobContainer(c *corev1.Container) JobOption {
	return func(dep *batchv1.Job) {
		dep.Spec.Template.Spec.Containers = append(dep.Spec.Template.Spec.Containers, *c)
	}
}

// DeleteJobAction creates a DeleteActionImpl that deletes jobs
// in Namespace test.
func DeleteJobAction(name string) clientgotesting.DeleteActionImpl {
	return clientgotesting.DeleteActionImpl{
		ActionImpl: clientgotesting.ActionImpl{
			Namespace: "test",
			Verb:      "delete",
			Resource: schema.GroupVersionResource{
				Group:    "batch",
				Version:  "v1",
				Resource: "job",
			},
		},
		Name: name,
	}
}
