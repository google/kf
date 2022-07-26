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

package v1alpha1

// MarkSpaceHealthy notes that the Space was able to be retrieved and
// defaults can be applied from it.
func (status *TaskScheduleStatus) MarkSpaceHealthy() {
	status.SpaceCondition().MarkSuccess()
}

// MarkSpaceUnhealthy notes that the Space was could not be retrieved.
func (status *TaskScheduleStatus) MarkSpaceUnhealthy(reason, message string) {
	status.SpaceCondition().MarkFalse(reason, message)
}

// MarkScheduleError notes that the TaskSchedule is not Ready due to an error
// with the cron schedule.
func (status *TaskScheduleStatus) MarkScheduleError(err error) {
	status.manage().MarkFalse(TaskScheduleConditionReady, "ScheduleError", err.Error())
}
