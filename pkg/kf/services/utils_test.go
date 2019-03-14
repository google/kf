package services

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

func TestParseJSONString(t *testing.T) {
	cases := map[string]struct {
		Value         string
		ExpectedMap   map[string]interface{}
		ExpectedError error
	}{
		"bad JSON": {
			Value:         "}}{{",
			ExpectedError: errors.New(`invalid JSON provided: "}}{{"`),
		},
		"empty JSON": {
			Value:       "{}",
			ExpectedMap: make(map[string]interface{}),
		},
		"JSON with contents": {
			Value:       `{"foo": "bar"}`,
			ExpectedMap: map[string]interface{}{"foo": "bar"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actualMap, actualErr := ParseJSONString(tc.Value)

			if tc.ExpectedError != nil || actualErr != nil {
				if fmt.Sprint(tc.ExpectedError) != fmt.Sprint(actualErr) {
					t.Fatalf("expected err: %v, got: %v", tc.ExpectedError, actualErr)
				}

				return
			}

			if !reflect.DeepEqual(tc.ExpectedMap, actualMap) {
				t.Errorf("expected map: %v, got: %v", tc.ExpectedMap, actualMap)
			}
		})
	}
}

func TestParseJSONOrFile(t *testing.T) {
	tmp, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp) // clean up

	badFile := path.Join(tmp, "bad.json")
	goodFile := path.Join(tmp, "good.json")

	if err := ioutil.WriteFile(badFile, []byte("}}{{"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(goodFile, []byte(`{"foo":"bar"}`), 0644); err != nil {
		t.Fatal(err)
	}

	cases := map[string]struct {
		Value         string
		ExpectedMap   map[string]interface{}
		ExpectedError error
	}{
		"empty JSON": {
			Value:       "{}",
			ExpectedMap: make(map[string]interface{}),
		},
		"JSON with contents": {
			Value:       `{"foo": "bar"}`,
			ExpectedMap: map[string]interface{}{"foo": "bar"},
		},
		"missing file": {
			Value:         `/path/does/not/exist`,
			ExpectedError: errors.New("couldn't read file: open /path/does/not/exist: no such file or directory"),
		},
		"bad file": {
			Value:         badFile,
			ExpectedError: fmt.Errorf("couldn't parse %s as JSON: invalid JSON provided: \"}}{{\"", badFile),
		},
		"good file": {
			Value:       goodFile,
			ExpectedMap: map[string]interface{}{"foo": "bar"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actualMap, actualErr := ParseJSONOrFile(tc.Value)

			if tc.ExpectedError != nil || actualErr != nil {
				if fmt.Sprint(tc.ExpectedError) != fmt.Sprint(actualErr) {
					t.Fatalf("expected err: %v, got: %v", tc.ExpectedError, actualErr)
				}

				return
			}

			if !reflect.DeepEqual(tc.ExpectedMap, actualMap) {
				t.Errorf("expected map: %v, got: %v", tc.ExpectedMap, actualMap)
			}
		})
	}

}

//
// func ParseJSONString(jsonString string) (map[string]interface{}, error) {
// 	p := make(map[string]interface{})
// 	if err := json.Unmarshal([]byte(jsonString), &p); err != nil {
// 		return nil, fmt.Errorf("invalid JSON provided: %q", jsonString)
// 	}
// 	return p, nil
// }
