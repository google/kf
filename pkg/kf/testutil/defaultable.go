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

// ApisDefaultableTest is a test format for assering apis.Defaultable objects
// work correctly.
type ApisDefaultableTest struct {
	// Context is the context in which to run the default.
	Context context.Context
	// Input is the value that is going to be defaulted.
	Input apis.Defaultable
	// Want is the expected result of calling SetDefaults() on the object.
	Want apis.Defaultable
}

// Run executes the test.
func (dt *ApisDefaultableTest) Run(t *testing.T) {
	t.Helper()
	dt.Input.SetDefaults(dt.Context)

	AssertEqual(t, "defaults", dt.Want, dt.Input)
}

// ApisDefaultableTestSuite defines a whole suite of defaultable tests that
// can be run. The key in the map is the name of each test.
type ApisDefaultableTestSuite map[string]ApisDefaultableTest

// Run executes the whole suite of tests.
func (ts ApisDefaultableTestSuite) Run(t *testing.T) {
	t.Helper()
	t.Parallel()
	for k, v := range ts {
		t.Run(k, v.Run)
	}
}
