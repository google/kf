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
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"knative.dev/pkg/kmp"
)

// Failable is an interface for testing.T like things. We can't use
// testing.TB which would be preferable because they explicitly set a
// private field on it to prevent others from implementing it.
type Failable interface {
	Helper()
	Fatalf(format string, args ...interface{})
	Name() string
}

// Assert causes a test to fail if the two values are not DeepEqual to
// one another.
func Assert(t Failable, m gomock.Matcher, actual interface{}) {
	t.Helper()
	if !m.Matches(actual) {
		t.Fatalf(m.String())
	}
}

// AssertTrue causes a test to fail if the asserted statement is not true.
func AssertTrue(t Failable, fieldName string, predicate bool) {
	t.Helper()

	if !predicate {
		t.Fatalf("%s: expected %s to be true, actual: %t", t.Name(), fieldName, predicate)
	}
}

// AssertFalse causes a test to fail if the asserted statement is true.
func AssertFalse(t Failable, fieldName string, predicate bool) {
	t.Helper()

	if predicate {
		t.Fatalf("%s: expected %s to be false, actual: %t", t.Name(), fieldName, predicate)
	}
}

// AssertEqual causes a test to fail if the two values are not DeepEqual to
// one another.
func AssertEqual(t Failable, fieldName string, expected, actual interface{}) {
	t.Helper()

	fail := func() {
		t.Helper()

		diff, err := kmp.SafeDiff(expected, actual)
		if err == nil {
			t.Fatalf("%s: expected %s to be equal expected: %#v actual: %#v diff (-expected, +actual): %s", t.Name(), fieldName, expected, actual, diff)
		} else {
			t.Fatalf("%s: expected %s to be equal expected: %#v actual: %#v", t.Name(), fieldName, expected, actual)
		}
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
		t.Fatalf("%s: expected %s to match pattern: %s actual: %s", t.Name(), fieldName, pattern, actual)
	}
}

// AssertErrorsEqual checks that the values of the two errors match.
func AssertErrorsEqual(t Failable, expected, actual error) {
	t.Helper()

	if fmt.Sprint(expected) != fmt.Sprint(actual) {
		t.Fatalf("%s: wanted err: %v, got: %v", t.Name(), expected, actual)
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
		t.Fatalf("%s: expected the values %v to be in %q but %v were missing", t.Name(), needles, haystack, missing)
	}
}

// AssertNil fails the test if the value is not nil.
func AssertNil(t Failable, name string, value interface{}) {
	t.Helper()

	if value != nil {
		t.Fatalf("%s: expected %s to be nil but got: %v", t.Name(), name, value)
	}
}

// AssertNotNil fails the test if the value is nil.
func AssertNotNil(t Failable, name string, value interface{}) {
	t.Helper()

	if value == nil {
		t.Fatalf("%s: expected %s not to be nil", t.Name(), name)
	}
}

// AssertNotBlank is used to validate that strings are not just made of
// whitespace.
func AssertNotBlank(t Failable, name string, value string) {
	t.Helper()

	if out := strings.TrimSpace(value); out == "" {
		t.Fatalf("%s: expected %s not to be blank", t.Name(), name)
	}
}

// AssertKeyWithValue ensures the key is present in the map and has the
// given value.
func AssertKeyWithValue(t Failable, m map[interface{}]interface{}, key, value interface{}) {
	t.Helper()

	a, ok := m[key]
	if !ok {
		t.Fatalf("%s: expected %v to have key %v", t.Name(), m, key)
	}

	if !reflect.DeepEqual(value, a) {
		t.Fatalf("%s: expected %v to have key %v with value %v", t.Name(), m, key, value)
	}
}

// AssertJSONEqual ensures the expexted JSON matches the actual JSON.
func AssertJSONEqual(t Failable, expected, actual string) {
	t.Helper()
	var em, am interface{}
	AssertNil(t, "expected JSON", json.Unmarshal([]byte(expected), &em))
	AssertNil(t, "actual JSON", json.Unmarshal([]byte(actual), &am))

	if !reflect.DeepEqual(em, am) {
		t.Fatalf("%s: expected %q to equal %q", t.Name(), actual, expected)
	}
}

// AssertRetrySucceeds asserts that a Retry call succeeds at least once.
func AssertRetrySucceeds(ctx context.Context, t Failable, times int, delay time.Duration, callback func() error) {
	t.Helper()

	err := Retry(ctx, times, delay, callback)
	AssertNil(t, "Retry error", err)
}

// Retry retries the callback for as long as the context does not have an
// error up to `times` times. If `times` is < 0, then it will try until the
// context has an error. The last error is returned.
func Retry(ctx context.Context, times int, delay time.Duration, callback func() error) error {
	// Retries n-1 times and ignores failures. Only return early on success.
	for i := 1; (i < times || times < 0) && ctx.Err() == nil; i++ {
		if err := callback(); err == nil {
			return nil
		}

		time.Sleep(delay)
	}

	// The last error (if any) is the one to keep.
	return callback()
}

// AssertUnstructuredEqual asserts that the actual value is equal to the value
// at the given path in the Unstructured object.
func AssertUnstructuredEqual(t *testing.T, field string, expected interface{}, u *unstructured.Unstructured) {
	t.Helper()

	actual, _, err := unstructured.NestedFieldNoCopy(
		u.Object,
		strings.Split(field, ".")...,
	)
	AssertErrorsEqual(t, nil, err)
	AssertEqual(t, field, expected, actual)
}
