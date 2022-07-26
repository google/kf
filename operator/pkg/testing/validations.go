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

package testing

import "testing"

// ValidateError validates an error against an expectation of an error
func ValidateError(t *testing.T, err error, expectError bool) {
	if err != nil && !expectError {
		t.Fatalf("expected no error, but received error: %v", err)
	} else if err == nil && expectError {
		t.Fatalf("expected error, but no error received.")
	}
}
