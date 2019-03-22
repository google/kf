package testutil

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// AssertEqual causes a test to fail if the two values are not DeepEqual to
// one another.
func AssertEqual(t *testing.T, fieldName string, expected, actual interface{}) {
	t.Helper()

	fail := func() {
		t.Helper()
		t.Fatalf("expected %s to be equal expected: %#v actual: %#v", fieldName, expected, actual)
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
func AssertRegexp(t *testing.T, fieldName, pattern, actual string) {
	t.Helper()

	if !regexp.MustCompile(pattern).MatchString(actual) {
		t.Fatalf("expected %s to match pattern: %s actual: %s", fieldName, pattern, actual)
	}
}

// AssertErrorsEqual checks that the values of the two errors match.
func AssertErrorsEqual(t *testing.T, expected, actual error) {
	t.Helper()

	if fmt.Sprint(expected) != fmt.Sprint(actual) {
		t.Fatalf("wanted err: %v, got: %v", expected, actual)
	}
}

// AssertContainsAll validates that all of the search strings were in the
// main string.
func AssertContainsAll(t *testing.T, haystack string, needles []string) {
	t.Helper()

	var missing []string
	for _, needle := range needles {
		if !strings.Contains(haystack, needle) {
			missing = append(missing, needle)
		}
	}

	if len(missing) > 0 {
		t.Fatalf("expected the values %v to be in %q but %v were missing", needles, haystack, missing)
	}
}

// AssertNil fails the test if the value is not nil.
func AssertNil(t *testing.T, name string, value interface{}) {
	t.Helper()

	if value != nil {
		t.Fatalf("expected %s to be nil but got: %v", name, value)
	}
}

// AssertKeyWithValue ensures the key is present in the map and has the
// given value.
func AssertKeyWithValue(t *testing.T, m map[interface{}]interface{}, key, value interface{}) {
	t.Helper()

	a, ok := m[key]
	if !ok {
		t.Fatalf("expected %v to have key %v", m, key)
	}

	if !reflect.DeepEqual(value, a) {
		t.Fatalf("expected %v to have key %v with value %v", m, key, value)
	}
}
