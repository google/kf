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

package garbagecollector

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck/v1beta1"
)

var referenceTime = time.Date(1992, 7, 22, 12, 0, 0, 0, time.UTC)

func makeTask(t time.Time, c v1.ConditionStatus) v1alpha1.Task {
	return v1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:              fmt.Sprintf("%s-%s", c, t),
			CreationTimestamp: metav1.NewTime(t),
		},
		Status: v1alpha1.TaskStatus{
			Status: v1beta1.Status{
				Conditions: v1beta1.Conditions{
					apis.Condition{
						Type:   apis.ConditionSucceeded,
						Status: c,
					},
				},
			},
		},
	}

}

func TestTasksToGC(t *testing.T) {
	cases := map[string]struct {
		tasks     []v1alpha1.Task
		maxTasks  int
		wantTasks []v1alpha1.Task
	}{
		"under limit": {
			tasks: []v1alpha1.Task{
				makeTask(referenceTime, v1.ConditionTrue),
			},
			maxTasks:  2,
			wantTasks: nil,
		},
		"gc's succeeded tasks": {
			tasks: []v1alpha1.Task{
				makeTask(referenceTime, v1.ConditionTrue),
				makeTask(referenceTime.Add(time.Minute), v1.ConditionTrue),
				makeTask(referenceTime.Add(time.Hour), v1.ConditionTrue),
			},
			maxTasks: 2,
			wantTasks: []v1alpha1.Task{
				makeTask(referenceTime, v1.ConditionTrue),
			},
		},
		"gc's failed tasks": {
			tasks: []v1alpha1.Task{
				makeTask(referenceTime, v1.ConditionFalse),
				makeTask(referenceTime.Add(time.Minute), v1.ConditionTrue),
				makeTask(referenceTime.Add(time.Hour), v1.ConditionTrue),
			},
			maxTasks: 2,
			wantTasks: []v1alpha1.Task{
				makeTask(referenceTime, v1.ConditionFalse),
			},
		},
		"doesn't gc in progress tasks": {
			tasks: []v1alpha1.Task{
				makeTask(referenceTime, v1.ConditionUnknown),
				makeTask(referenceTime.Add(time.Minute), v1.ConditionTrue),
				makeTask(referenceTime.Add(time.Hour), v1.ConditionTrue),
			},
			maxTasks: 2,
			wantTasks: []v1alpha1.Task{
				makeTask(referenceTime.Add(time.Minute), v1.ConditionTrue),
			},
		},
		"gc's oldest task when unordered": {
			tasks: []v1alpha1.Task{
				makeTask(referenceTime.Add(time.Minute), v1.ConditionTrue),
				makeTask(referenceTime.Add(time.Hour), v1.ConditionTrue),
				makeTask(referenceTime, v1.ConditionTrue),
			},
			maxTasks: 2,
			wantTasks: []v1alpha1.Task{
				makeTask(referenceTime, v1.ConditionTrue),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := tasksToGC(tc.tasks, tc.maxTasks)
			testutil.AssertEqual(t, "tasks", tc.wantTasks, actual)
		})
	}
}

func makeBuild(t time.Time, c v1.ConditionStatus) v1alpha1.Build {
	return v1alpha1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:              fmt.Sprintf("%s-%s", c, t),
			CreationTimestamp: metav1.NewTime(t),
		},
		Status: v1alpha1.BuildStatus{
			Status: v1beta1.Status{
				Conditions: v1beta1.Conditions{
					apis.Condition{
						Type:   apis.ConditionSucceeded,
						Status: c,
					},
				},
			},
		},
	}

}

func TestBuildsToGC(t *testing.T) {
	cases := map[string]struct {
		builds     []v1alpha1.Build
		maxBuilds  int
		wantBuilds []v1alpha1.Build
	}{
		"under limit without nonfinal builds": {
			builds: []v1alpha1.Build{
				makeBuild(referenceTime, v1.ConditionTrue),
			},
			maxBuilds:  2,
			wantBuilds: nil,
		},
		"under limit with nonfinal builds": {
			builds: []v1alpha1.Build{
				makeBuild(referenceTime, v1.ConditionUnknown),
				makeBuild(referenceTime.Add(time.Minute), v1.ConditionUnknown),
				makeBuild(referenceTime, v1.ConditionTrue),
			},
			maxBuilds:  1,
			wantBuilds: nil,
		},
		"gc's succeeded builds": {
			builds: []v1alpha1.Build{
				makeBuild(referenceTime, v1.ConditionTrue),
				makeBuild(referenceTime.Add(time.Minute), v1.ConditionTrue),
				makeBuild(referenceTime.Add(time.Hour), v1.ConditionTrue),
			},
			maxBuilds: 2,
			wantBuilds: []v1alpha1.Build{
				makeBuild(referenceTime, v1.ConditionTrue),
			},
		},
		"gc's failed builds": {
			builds: []v1alpha1.Build{
				makeBuild(referenceTime, v1.ConditionFalse),
				makeBuild(referenceTime.Add(time.Minute), v1.ConditionTrue),
				makeBuild(referenceTime.Add(time.Hour), v1.ConditionTrue),
			},
			maxBuilds: 2,
			wantBuilds: []v1alpha1.Build{
				makeBuild(referenceTime, v1.ConditionFalse),
			},
		},
		"doesn't gc in progress builds": {
			builds: []v1alpha1.Build{
				makeBuild(referenceTime, v1.ConditionUnknown),
				makeBuild(referenceTime.Add(time.Minute), v1.ConditionTrue),
				makeBuild(referenceTime.Add(time.Hour), v1.ConditionTrue),
				makeBuild(referenceTime.Add(2*time.Hour), v1.ConditionTrue),
			},
			maxBuilds: 2,
			wantBuilds: []v1alpha1.Build{
				makeBuild(referenceTime.Add(time.Minute), v1.ConditionTrue),
			},
		},
		"gc's oldest build when unordered": {
			builds: []v1alpha1.Build{
				makeBuild(referenceTime, v1.ConditionTrue),
				makeBuild(referenceTime.Add(time.Minute), v1.ConditionTrue),
				makeBuild(referenceTime.Add(time.Hour), v1.ConditionTrue),
			},
			maxBuilds: 2,
			wantBuilds: []v1alpha1.Build{
				makeBuild(referenceTime, v1.ConditionTrue),
			},
		},
		"gc multiple older builds": {
			builds: []v1alpha1.Build{
				makeBuild(referenceTime, v1.ConditionTrue),
				makeBuild(referenceTime.Add(time.Minute), v1.ConditionTrue),
				makeBuild(referenceTime.Add(time.Hour), v1.ConditionTrue),
			},
			maxBuilds: 1,
			wantBuilds: []v1alpha1.Build{
				makeBuild(referenceTime.Add(time.Minute), v1.ConditionTrue),
				makeBuild(referenceTime, v1.ConditionTrue),
			},
		},
		"retain at minimum one build - single build": {
			builds: []v1alpha1.Build{
				makeBuild(referenceTime, v1.ConditionTrue),
			},
			maxBuilds:  0,
			wantBuilds: nil,
		},
		"retain at minimum one build - multiple builds": {
			builds: []v1alpha1.Build{
				makeBuild(referenceTime, v1.ConditionTrue),
				makeBuild(referenceTime.Add(time.Minute), v1.ConditionTrue),
			},
			maxBuilds: 0,
			wantBuilds: []v1alpha1.Build{
				makeBuild(referenceTime, v1.ConditionTrue),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := buildsToGC(tc.builds, tc.maxBuilds)
			testutil.AssertEqual(t, "builds", tc.wantBuilds, actual)
		})
	}
}
