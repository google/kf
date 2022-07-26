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

import (
	"context"
	"strings"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

func TestTaskSchedule_Validate_statusUpdate(t *testing.T) {
	t.Parallel()

	badMeta := metav1.ObjectMeta{
		Name: strings.Repeat("A", 64), // Too long
	}

	ctx := context.Background()
	ctx = apis.WithinSubResourceUpdate(ctx, &TaskSchedule{}, "status")

	// Validation should be skipped while updating the status.
	f := &TaskSchedule{ObjectMeta: badMeta}
	err := f.Validate(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTaskSchedule_Validate(t *testing.T) {
	goodTaskTemplate := TaskSpec{
		AppRef: corev1.LocalObjectReference{
			Name: "appRef",
		},
	}

	goodSpec := TaskScheduleSpec{
		ConcurrencyPolicy: "Always",
		Schedule:          "* * * * *",
		TaskTemplate:      goodTaskTemplate,
	}

	badMeta := metav1.ObjectMeta{
		Name: strings.Repeat("A", 64), // Too long
	}

	goodMeta := metav1.ObjectMeta{
		Name: "valid",
	}

	cases := map[string]struct {
		spec TaskSchedule
		want *apis.FieldError
	}{
		"valid spec": {
			spec: TaskSchedule{
				ObjectMeta: goodMeta,
				Spec:       goodSpec,
			},
		},
		"invalid ObjectMeta": {
			spec: TaskSchedule{
				ObjectMeta: badMeta,
				Spec:       goodSpec,
			},
			want: apis.ValidateObjectMetadata(badMeta.GetObjectMeta()).ViaField("metadata"),
		},
		"invalid schedule cron expression": {
			spec: TaskSchedule{
				ObjectMeta: goodMeta,
				Spec: TaskScheduleSpec{
					ConcurrencyPolicy: "Always",
					Schedule:          "foo",
					TaskTemplate:      goodTaskTemplate,
				},
			},
			want: apis.ErrInvalidValue("foo", "spec.schedule"),
		},
		"invalid task template": {
			spec: TaskSchedule{
				ObjectMeta: goodMeta,
				Spec: TaskScheduleSpec{
					ConcurrencyPolicy: "Always",
					Schedule:          "* * * * *",
					TaskTemplate:      TaskSpec{},
				},
			},
			want: apis.ErrMissingField("spec.taskTemplate.appRef"),
		},
		"invalid concurrency policy": {
			spec: TaskSchedule{
				ObjectMeta: goodMeta,
				Spec: TaskScheduleSpec{
					ConcurrencyPolicy: "badPolicy",
					Schedule:          "* * * * *",
					TaskTemplate:      goodTaskTemplate,
				},
			},
			want: apis.ErrInvalidValue("badPolicy", "spec.concurrencyPolicy"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := tc.spec.Validate(context.Background())

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}
