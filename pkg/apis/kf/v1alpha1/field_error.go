// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"fmt"
	"sort"
	"strings"

	"knative.dev/pkg/apis"
)

// ErrInvalidEnumValue constructs a FieldError for a field that has received an
// invalid string value.
func ErrInvalidEnumValue(value interface{}, fieldPath string, acceptedValues []string) *apis.FieldError {
	var strs []string
	for _, v := range acceptedValues {
		strs = append(strs, fmt.Sprint(v))
	}

	sort.Strings(strs)

	return &apis.FieldError{
		Message: fmt.Sprintf("invalid value: %v, should be one of: %s", value, strings.Join(strs, ", ")),
		Paths:   []string{fieldPath},
	}
}
