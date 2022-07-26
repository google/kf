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

package taskschedule

import (
	"testing"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	cron "github.com/robfig/cron/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetNextScheduleTime(t *testing.T) {
	t.Parallel()
	type args struct {
		earliestTime *time.Time
		now          time.Time
		schedule     string
	}
	tests := []struct {
		name         string
		args         args
		expectedTime *time.Time
		wantErr      bool
	}{
		{
			name: "now before next schedule",
			args: args{
				earliestTime: topOfTheHour(),
				now:          topOfTheHour().Add(time.Second * 30),
				schedule:     "0 * * * *",
			},
			expectedTime: nil,
		},
		{
			name: "now just after next schedule",
			args: args{
				earliestTime: topOfTheHour(),
				now:          topOfTheHour().Add(time.Minute * 61),
				schedule:     "0 * * * *",
			},
			expectedTime: deltaTimeAfterTopOfTheHour(time.Minute * 60),
		},
		{
			name: "missed 5 schedules",
			args: args{
				earliestTime: deltaTimeAfterTopOfTheHour(time.Second * 10),
				now:          *deltaTimeAfterTopOfTheHour(time.Minute * 301),
				schedule:     "0 * * * *",
			},
			expectedTime: deltaTimeAfterTopOfTheHour(time.Minute * 300),
		},
		{
			name: "rogue cronjob",
			args: args{
				earliestTime: deltaTimeAfterTopOfTheHour(time.Second * 10),
				now:          *deltaTimeAfterTopOfTheHour(time.Hour * 1000000),
				schedule:     "59 23 31 2 *",
			},
			expectedTime: nil,
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sched, err := cron.ParseStandard(tt.args.schedule)
			if err != nil {
				t.Errorf("error setting up the test, %s", err)
			}
			gotTime, err := getNextScheduleTime(v1alpha1.TaskSchedule{
				Status: v1alpha1.TaskScheduleStatus{
					TaskScheduleStatusFields: v1alpha1.TaskScheduleStatusFields{
						LastScheduleTime: &metav1.Time{
							Time: *tt.args.earliestTime,
						},
					},
				},
			}, tt.args.now, sched)
			if tt.wantErr {
				if err == nil {
					t.Error("getNextScheduleTime() got no error when expected one")
				}
				return
			}
			if !tt.wantErr && err != nil {
				t.Error("getNextScheduleTime() got error when none expected")
			}
			if gotTime == nil && tt.expectedTime != nil {
				t.Errorf("getNextScheduleTime() got nil, want %v", tt.expectedTime)
			}
			if gotTime != nil && tt.expectedTime != nil && !gotTime.Equal(*tt.expectedTime) {
				t.Errorf("getNextScheduleTime() got = %v, want %v", gotTime, tt.expectedTime)
			}
		})
	}
}

func makeTaskSchedule(names ...string) v1alpha1.TaskSchedule {
	var references []corev1.LocalObjectReference
	for _, name := range names {
		references = append(references, corev1.LocalObjectReference{
			Name: name,
		})
	}
	return v1alpha1.TaskSchedule{
		Status: v1alpha1.TaskScheduleStatus{
			TaskScheduleStatusFields: v1alpha1.TaskScheduleStatusFields{
				Active: references,
			},
		},
	}
}

func TestInActiveList(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		taskSchedule v1alpha1.TaskSchedule
		task         v1alpha1.Task
		expected     bool
	}{
		{
			name:         "exists",
			taskSchedule: makeTaskSchedule("foo", "bar", "baz"),
			task: v1alpha1.Task{
				ObjectMeta: metav1.ObjectMeta{
					Name: "bar",
				},
			},
			expected: true,
		},
		{
			name:         "doesn't exist",
			taskSchedule: makeTaskSchedule("foo", "bar", "baz"),
			task: v1alpha1.Task{
				ObjectMeta: metav1.ObjectMeta{
					Name: "quux",
				},
			},
			expected: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutil.AssertEqual(t, "inActive", tc.expected, inActiveList(&tc.taskSchedule, &tc.task))
		})
	}
}

func TestRemoveFromActive(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		taskSchedule v1alpha1.TaskSchedule
		task         v1alpha1.Task
		expected     []string
	}{
		{
			name:         "exists",
			taskSchedule: makeTaskSchedule("foo", "bar", "baz"),
			task: v1alpha1.Task{
				ObjectMeta: metav1.ObjectMeta{
					Name: "bar",
				},
			},
			expected: []string{"foo", "baz"},
		},
		{
			name:         "only active task",
			taskSchedule: makeTaskSchedule("foo"),
			task: v1alpha1.Task{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
			expected: []string{},
		},
		{
			name:         "doesn't exist",
			taskSchedule: makeTaskSchedule("foo", "bar", "baz"),
			task: v1alpha1.Task{
				ObjectMeta: metav1.ObjectMeta{
					Name: "quux",
				},
			},
			expected: []string{"foo", "bar", "baz"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			removeFromActive(&tc.taskSchedule, &tc.task)
			var expected []corev1.LocalObjectReference
			for _, name := range tc.expected {
				expected = append(expected, corev1.LocalObjectReference{
					Name: name,
				})
			}
			testutil.AssertEqual(t, "active", expected, tc.taskSchedule.Status.Active)
		})
	}
}

func topOfTheHour() *time.Time {
	T1, err := time.Parse(time.RFC3339, "2016-05-19T10:00:00Z")
	if err != nil {
		panic("test setup error")
	}
	return &T1
}

func deltaTimeAfterTopOfTheHour(duration time.Duration) *time.Time {
	T1, err := time.Parse(time.RFC3339, "2016-05-19T10:00:00Z")
	if err != nil {
		panic("test setup error")
	}
	t := T1.Add(duration)
	return &t
}
