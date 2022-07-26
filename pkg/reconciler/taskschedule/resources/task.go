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

package resources

import (
	"fmt"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

const (
	managedByLabel     = "app.kubernetes.io/managed-by"
	OwningTaskSchedule = "taskschedules.kf.dev/owner"
)

// getTaskName creates a name for the Task based on the TaskSchedule name and a
// hash of the scheduled time (to the minute). This lets us use the Task name as
// a lock to prevent making the same Task twice.
func getTaskName(taskSchedule *v1alpha1.TaskSchedule, scheduledTime time.Time) string {
	return fmt.Sprintf("%s-%d", taskSchedule.Name, scheduledTime.Truncate(time.Minute).Unix())
}

// MakeTask creates a Task for the given TaskSchedule and time.
func MakeTask(taskSchedule *v1alpha1.TaskSchedule, scheduledTime time.Time) *v1alpha1.Task {
	return &v1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getTaskName(taskSchedule, scheduledTime),
			Namespace: taskSchedule.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(taskSchedule),
			},
			Labels: v1alpha1.UnionMaps(
				taskSchedule.GetLabels(),
				map[string]string{
					managedByLabel: "kf",
					// Add owning TaskSchedule label as it is not possible to
					// filter lists based on metadata.ownerReferences. We use
					// this to filter Task lists by TaskSchedule.
					OwningTaskSchedule: taskSchedule.Name,
				},
			),
		},
		Spec: taskSchedule.Spec.TaskTemplate,
	}
}
