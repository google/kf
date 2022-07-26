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

package testutil

import (
	"context"
	"testing"

	"knative.dev/pkg/apis"
)

// ApisValidatableTest is a test format for assering apis.Validatable objects
// work correctly.
type ApisValidatableTest struct {
	// Context is the context in which to run the validation.
	Context context.Context
	// Input is the value that is going to be validated.
	Input apis.Validatable
	// Want is the expected result of calling SetDefaults() on the object.
	Want *apis.FieldError
}

// Run executes the test.
func (dt *ApisValidatableTest) Run(t *testing.T) {
	got := dt.Input.Validate(dt.Context)

	AssertEqual(t, "validation errors", dt.Want.Error(), got.Error())
}

// ApisValidatableTestSuite defines a whole suite of validatable tests that
// can be run. The key in the map is the name of each test.
type ApisValidatableTestSuite map[string]ApisValidatableTest

// Run executes the whole suite of tests.
func (ts ApisValidatableTestSuite) Run(t *testing.T) {
	t.Parallel()
	for k, v := range ts {
		t.Run(k, v.Run)
	}
}
