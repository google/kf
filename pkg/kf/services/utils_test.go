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

package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/google/kf/pkg/kf/testutil"
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAssertJSONMap(t *testing.T) {
	cases := map[string]struct {
		Value         string
		ExpectedJSON  json.RawMessage
		ExpectedError error
	}{
		"bad JSON": {
			Value:         "}}{{",
			ExpectedError: errors.New(`value must be a JSON map, got: "}}{{"`),
		},
		"empty JSON": {
			Value:        "{}",
			ExpectedJSON: json.RawMessage("{}"),
		},
		"JSON with contents": {
			Value:        `{"foo": "bar"}`,
			ExpectedJSON: json.RawMessage(`{"foo": "bar"}`),
		},
		"not a map": {
			Value:         `1234`,
			ExpectedError: errors.New(`value must be a JSON map, got: "1234"`),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actualJSON, actualErr := AssertJSONMap(json.RawMessage(tc.Value))

			if tc.ExpectedError != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedError, actualErr)

				return
			}

			if !reflect.DeepEqual(tc.ExpectedJSON, actualJSON) {
				t.Errorf("expected JSON: %v, got: %v", string(tc.ExpectedJSON), string(actualJSON))
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
		ExpectedJSON  json.RawMessage
		ExpectedError error
	}{
		"empty JSON": {
			Value:        "{}",
			ExpectedJSON: json.RawMessage(`{}`),
		},
		"JSON with contents": {
			Value:        `{"foo": "bar"}`,
			ExpectedJSON: json.RawMessage(`{"foo": "bar"}`),
		},
		"missing file": {
			Value:         `/path/does/not/exist`,
			ExpectedError: errors.New("couldn't read file: open /path/does/not/exist: no such file or directory"),
		},
		"bad file": {
			Value:         badFile,
			ExpectedError: fmt.Errorf("couldn't parse %s as JSON: value must be a JSON map, got: \"}}{{\"", badFile),
		},
		"good file": {
			Value:        goodFile,
			ExpectedJSON: json.RawMessage(`{"foo":"bar"}`),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actualJSON, actualErr := ParseJSONOrFile(tc.Value)

			if tc.ExpectedError != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedError, actualErr)

				return
			}

			if !reflect.DeepEqual(tc.ExpectedJSON, actualJSON) {
				t.Errorf("expected JSON: %v, got: %v", string(tc.ExpectedJSON), string(actualJSON))
			}
		})
	}
}

func ExampleLastStatusCondition() {
	si := v1beta1.ServiceInstance{
		Status: v1beta1.ServiceInstanceStatus{
			Conditions: []v1beta1.ServiceInstanceCondition{
				{Status: "Wrong"},
				{LastTransitionTime: metav1.Time{Time: time.Now()}, Status: "Ready"},
			},
		},
	}

	c := LastStatusCondition(si)
	if _, err := fmt.Println(c.Status); err != nil {
		panic(err)
	}

	// Output: Ready
}
