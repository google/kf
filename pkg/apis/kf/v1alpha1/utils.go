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

package v1alpha1

import "knative.dev/pkg/apis"

// PropagateCondition copies the condition of a sub-resource (source) to a
// destination on the given manager.
// It returns true if the condition is true, otherwise false.
func PropagateCondition(manager apis.ConditionManager, destination apis.ConditionType, source *apis.Condition) bool {
	switch {
	case source == nil:
		return false
	case source.IsFalse():
		manager.MarkFalse(destination, source.Reason, source.Message)
	case source.IsTrue():
		manager.MarkTrue(destination)
	case source.IsUnknown():
		manager.MarkUnknown(destination, source.Reason, source.Message)
	}

	return source.IsTrue()
}
