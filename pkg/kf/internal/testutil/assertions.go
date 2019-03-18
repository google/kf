package testutil

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"
)

// AssertEqual causes a test to fail if the two values are not DeepEqual to
// one another.
func AssertEqual(t *testing.T, fieldName string, expected, actual interface{}) {
	t.Helper()

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected %s to be equal expected: %#v actual: %#v", fieldName, expected, actual)
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
