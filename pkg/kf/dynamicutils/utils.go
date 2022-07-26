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

package dynamicutils

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"knative.dev/pkg/logging"
)

// CheckCondtions returns the followings state table:
// * If all are True, then it returns True.
// * If all are True or Unknown, then it returns Unknown.
// * If any are False, then it returns False.
func CheckCondtions(ctx context.Context, u *unstructured.Unstructured) corev1.ConditionStatus {
	logger := logging.FromContext(ctx)
	conditions, found, err := unstructured.NestedSlice(u.Object, "status", "conditions")
	if err != nil {
		logger.Warnf("failed to fetch status.conditions: %v", err)
		return corev1.ConditionFalse
	}

	if !found {
		// No conditions found, interpret as Unknown
		return corev1.ConditionUnknown
	}

	// Verify the ObservedGeneration and generation match. If not, then the
	// controllers have not finished with them.
	generation, _, err := unstructured.NestedInt64(u.Object, "metadata", "generation")
	if err != nil {
		logger.Warnf("failed to fetch metadata.generation: %v", err)
		return corev1.ConditionFalse
	}

	observedGeneration, _, err := unstructured.NestedInt64(u.Object, "status", "observedGeneration")
	if err != nil {
		logger.Warnf("failed to fetch status.observedGeneration: %v", err)
		return corev1.ConditionFalse
	}

	if generation != observedGeneration {
		return corev1.ConditionUnknown
	}

	hasUnknown := false

	for _, c := range conditions {
		cc, ok := c.(map[string]interface{})
		if !ok {
			// Mishaped, return false.
			return corev1.ConditionFalse
		}

		status, _ := cc["status"].(string)
		switch corev1.ConditionStatus(status) {
		case corev1.ConditionUnknown:
			// We can't just bail in case there is a False or invalid status.
			hasUnknown = true
		case corev1.ConditionTrue:
			// Good, keep going.
			continue
		case corev1.ConditionFalse:
			return corev1.ConditionFalse
		default:
			// This is invalid.
			return corev1.ConditionFalse
		}
	}

	if hasUnknown {
		return corev1.ConditionUnknown
	}

	return corev1.ConditionTrue
}

// NewUnstructured returns an Unstructured object that is populated by a map
// of fields with their associated values. Each field should be separated by
// periods. So for example, if we wanted to create this struct:
//
// {
//
//	"metadata": {
//	  "generation": 101,
//	},
//	"status": {
//	  "observedGeneration": 99
//	}
//
// }
//
// This would be expressed with:
// map[string]interface{} {
//
//	"metadata.generation": 101,
//	"status.observedGeneration": 99,
//
// }
func NewUnstructured(m map[string]interface{}) *unstructured.Unstructured {
	u := &unstructured.Unstructured{
		Object: make(map[string]interface{}),
	}

	for field, v := range m {
		SetNestedFieldNoCopy(
			u.Object,
			v,
			strings.Split(field, ".")...,
		)
	}

	return u
}

// SetNestedFieldNoCopy is shameless copied from
// https://github.com/kubernetes/kubernetes/blob/v1.14.10/staging/src/k8s.io/apimachinery/pkg/apis/meta/v1/unstructured/helpers.go
func SetNestedFieldNoCopy(obj map[string]interface{}, value interface{}, fields ...string) error {
	m := obj

	for i, field := range fields[:len(fields)-1] {
		if val, ok := m[field]; ok {
			if valMap, ok := val.(map[string]interface{}); ok {
				m = valMap
			} else {
				return fmt.Errorf("value cannot be set because %v is not a map[string]interface{}", jsonPath(fields[:i+1]))
			}
		} else {
			newVal := make(map[string]interface{})
			m[field] = newVal
			m = newVal
		}
	}
	m[fields[len(fields)-1]] = value
	return nil
}

func jsonPath(fields []string) string {
	return "." + strings.Join(fields, ".")
}
