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

package dynamicutils_test

import (
	"context"
	"testing"

	"github.com/google/kf/v2/pkg/kf/dynamicutils"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestConditionsTrue(t *testing.T) {
	t.Parallel()

	build := func(c ...string) *unstructured.Unstructured {
		conditions := make([]interface{}, 0)
		for _, cc := range c {
			conditions = append(conditions, map[string]interface{}{
				"status": cc,
			})
		}

		return &unstructured.Unstructured{
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{},
				"status": map[string]interface{}{
					"conditions": conditions,
				},
			},
		}
	}

	buildWithGenerations := func(observedGeneration, generation interface{}, u *unstructured.Unstructured) *unstructured.Unstructured {
		u.Object["metadata"] = map[string]interface{}{
			"generation": generation,
		}
		u.Object["status"].(map[string]interface{})["observedGeneration"] = observedGeneration
		return u
	}

	testCases := []struct {
		name     string
		expected corev1.ConditionStatus
		u        *unstructured.Unstructured
	}{
		{
			name:     "all True without generations",
			expected: corev1.ConditionTrue,
			u:        build("True", "True", "True"),
		},
		{
			name:     "all True with generations",
			expected: corev1.ConditionTrue,
			u:        buildWithGenerations(int64(99), int64(99), build("True", "True", "True")),
		},
		{
			name:     "all True with invalid generation",
			expected: corev1.ConditionFalse,
			u:        buildWithGenerations(int64(99), "invalid", build("True", "True", "True")),
		},
		{
			name:     "all True with invalid observedGeneration",
			expected: corev1.ConditionFalse,
			u:        buildWithGenerations("invalid", int64(99), build("True", "True", "True")),
		},
		{
			name:     "all True and an Unknown",
			expected: corev1.ConditionUnknown,
			u:        build("True", "Unknown", "True"),
		},
		{
			name:     "all False",
			expected: corev1.ConditionFalse,
			u:        build("False", "False", "False"),
		},
		{
			name:     "all Unknown",
			expected: corev1.ConditionUnknown,
			u:        build("Unknown", "Unknown", "Unknown"),
		},
		{
			name:     "all False and an Unknown",
			expected: corev1.ConditionFalse,
			u:        build("False", "Unknown", "False"),
		},
		{
			name:     "one False",
			expected: corev1.ConditionFalse,
			u:        build("True", "False", "True"),
		},
		{
			name:     "empty conditions",
			expected: corev1.ConditionTrue,
			u:        build(),
		},
		{
			name:     "no status field",
			expected: corev1.ConditionUnknown,
			u:        &unstructured.Unstructured{},
		},
		{
			name:     "one invalid",
			expected: corev1.ConditionFalse,
			u:        build("True", "invalid", "True"),
		},
		{
			name:     "mismatch generations",
			expected: corev1.ConditionUnknown,
			u:        buildWithGenerations(int64(1), int64(2), build("True")),
		},
		{
			name:     "invalid status field",
			expected: corev1.ConditionFalse,
			u: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": []string{"a", "b"},
				},
			},
		},
		{
			name:     "invalid status.condition field",
			expected: corev1.ConditionFalse,
			u: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{"a", "b"},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testutil.AssertEqual(t, "comp", tc.expected, dynamicutils.CheckCondtions(context.Background(), tc.u))
		})
	}
}

func TestNewUnstructured(t *testing.T) {
	t.Parallel()

	m := []map[string]interface{}{
		{
			"baz": "asdf",
			"xyz": []string{"a", "b", "c"},
		},
	}

	u := dynamicutils.NewUnstructured(
		map[string]interface{}{
			"a.b.c": "foo",
			"a.b.d": m,
		},
	)

	testutil.AssertUnstructuredEqual(t, "a.b.c", "foo", u)
	testutil.AssertUnstructuredEqual(t, "a.b.d", m, u)
}

func TestSetNestedFieldNoCopy(t *testing.T) {
	t.Parallel()

	u := &unstructured.Unstructured{
		Object: make(map[string]interface{}),
	}
	dynamicutils.SetNestedFieldNoCopy(u.Object, "foo", "a", "b", "c")
	testutil.AssertUnstructuredEqual(t, "a.b.c", "foo", u)
}
