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
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	gomock "github.com/golang/mock/gomock"
	"knative.dev/pkg/kmp"
)

// Failable is an interface for testing.T like things. We can't use
// testing.TB which would be preferable because they explicitly set a
// private field on it to prevent others from implementing it.
type Failable interface {
	Helper()
	Fatalf(format string, args ...interface{})
}

// Assert causes a test to fail if the two values are not DeepEqual to
// one another.
func Assert(t Failable, m gomock.Matcher, actual interface{}) {
	t.Helper()
	if !m.Matches(actual) {
		t.Fatalf(m.String())
	}
}

// AssertEqual causes a test to fail if the two values are not DeepEqual to
// one another.
func AssertEqual(t Failable, fieldName string, expected, actual interface{}) {
	t.Helper()

	fail := func() {
		t.Helper()

		// Ignore diff error
		diff, _ := kmp.SafeDiff(expected, actual)
		t.Fatalf("expected %s to be equal expected: %#v actual: %#v\ndiff: %s", fieldName, expected, actual, diff)
	}

	v1, v2 := reflect.ValueOf(expected), reflect.ValueOf(actual)
	if v1.Kind() != v2.Kind() {
		fail()
	}

	// If the type has a length and can be nil, ensure one isn't nil.
	// DeepEqual would return a false positive.
	switch v1.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		if v1.Len() == 0 && v2.Len() == 0 {
			return
		}
	}

	if !reflect.DeepEqual(expected, actual) {
		fail()
	}
}

// AssertRegexp causes a test to fail if the value does not match a pattern.
func AssertRegexp(t Failable, fieldName, pattern, actual string) {
	t.Helper()

	if !regexp.MustCompile(pattern).MatchString(actual) {
		t.Fatalf("expected %s to match pattern: %s actual: %s", fieldName, pattern, actual)
	}
}

// AssertErrorsEqual checks that the values of the two errors match.
func AssertErrorsEqual(t Failable, expected, actual error) {
	t.Helper()

	if fmt.Sprint(expected) != fmt.Sprint(actual) {
		t.Fatalf("wanted err: %v, got: %v", expected, actual)
	}
}

// AssertErrorContainsAll checks that all of the search strings
// were in the error message. Case-insensitive.
func AssertErrorContainsAll(t Failable, errorMsg error, expected []string) {
	t.Helper()

	AssertContainsAll(t, errorMsg.Error(), expected)
}

// AssertContainsAll validates that all of the search strings were in the
// main string. Case-insensitive.
func AssertContainsAll(t Failable, haystack string, needles []string) {
	t.Helper()

	var missing []string
	for _, needle := range needles {
		if !strings.Contains(strings.ToLower(haystack), strings.ToLower(needle)) {
			missing = append(missing, needle)
		}
	}

	if len(missing) > 0 {
		t.Fatalf("expected the values %v to be in %q but %v were missing", needles, haystack, missing)
	}
}

// AssertNil fails the test if the value is not nil.
func AssertNil(t Failable, name string, value interface{}) {
	t.Helper()

	if value != nil {
		t.Fatalf("expected %s to be nil but got: %v", name, value)
	}
}

// AssertNotNil fails the test if the value is nil.
func AssertNotNil(t Failable, name string, value interface{}) {
	t.Helper()

	if value == nil {
		t.Fatalf("expected %s not to be nil", name)
	}
}

// AssertKeyWithValue ensures the key is present in the map and has the
// given value.
func AssertKeyWithValue(t Failable, m map[interface{}]interface{}, key, value interface{}) {
	t.Helper()

	a, ok := m[key]
	if !ok {
		t.Fatalf("expected %v to have key %v", m, key)
	}

	if !reflect.DeepEqual(value, a) {
		t.Fatalf("expected %v to have key %v with value %v", m, key, value)
	}
}

// AssertJSONEqual ensures the expexted JSON matches the actual JSON.
func AssertJSONEqual(t Failable, expected, actual string) {
	t.Helper()
	var em, am interface{}
	AssertNil(t, "expected JSON", json.Unmarshal([]byte(expected), &em))
	AssertNil(t, "actual JSON", json.Unmarshal([]byte(expected), &am))

	if !reflect.DeepEqual(em, am) {
		t.Fatalf("expected %q to equal %q", actual, expected)
	}
}
