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

package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	configlogging "github.com/google/kf/v2/pkg/kf/commands/config/logging"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/reconciler/route/resources"
	"github.com/spf13/cobra"
)

func ExampleCreateProxyDirector() {
	director := CreateProxyDirector("my-app.example.com", "kf.proxy.gateway", http.Header{
		resources.KfAppMatchHeader: []string{"my-app"},
	})

	req, err := http.NewRequest(http.MethodGet, "http://my-app.example.com/foo", nil)
	if err != nil {
		panic(err)
	}

	director(req)

	fmt.Println("URL:", req.URL.String())
	fmt.Println("Host:", req.Host)
	fmt.Println("Header:", req.Header)

	// Output: URL: http://kf.proxy.gateway/foo
	// Host: my-app.example.com
	// Header: map[X-Kf-App:[my-app]]
}

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

func TestSplitTags(t *testing.T) {
	tests := map[string]struct {
		tagStr string
		want   []string
	}{
		"blank is nil": {
			tagStr: "",
			want:   nil,
		},
		"space and commas": {
			// taken from the CF documentation about how to use tags:
			// https://cli.cloudfoundry.org/en-US/cf/create-service.html
			tagStr: "list, of, tags",
			want:   []string{"list", "of", "tags"},
		},
		"CSV": {
			tagStr: "list,of,tags",
			want:   []string{"list", "of", "tags"},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			got := SplitTags(tc.tagStr)
			testutil.AssertEqual(t, "tags", tc.want, got)
		})
	}
}

func ExampleJoinHeredoc() {
	fmt.Println(JoinHeredoc(`List:

		* a
		* b
		* c
		`, // normal documentation with indentation.
		"NOTE: single line of text.",
		`



		`, // blank lines get ignored
	))

	// Output: List:
	//
	// * a
	// * b
	// * c
	//
	// NOTE: single line of text.
}

func TestAddPreviewWarning(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		cmd         *cobra.Command
		expectedOut string
		expectedErr error
	}{
		{
			name: "no preRun",
			cmd: &cobra.Command{
				Use: "cmd",
			},
		},
		{
			name:        "with preRun",
			expectedOut: "a b c",
			cmd: &cobra.Command{
				Use: "cmd",
				PersistentPreRun: func(cmd *cobra.Command, args []string) {
					fmt.Fprint(cmd.OutOrStderr(), strings.Join(args, " "))
				},
			},
		},
		{
			name:        "with preRunE",
			expectedOut: "a b c",
			cmd: &cobra.Command{
				Use: "cmd",
				PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
					fmt.Fprint(cmd.OutOrStderr(), strings.Join(args, " "))
					return errors.New("some-error")
				},
			},
			expectedErr: errors.New("some-error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			addPreviewWarning(tc.cmd, previewWarning)
			bufErr := new(bytes.Buffer)
			buf := new(bytes.Buffer)
			tc.cmd.SetOut(buf)
			tc.cmd.SetContext(configlogging.SetupLogger(context.Background(), bufErr))
			err := tc.cmd.PersistentPreRunE(tc.cmd, []string{"a", "b", "c"})

			if !strings.Contains(bufErr.String(), previewWarning) {
				t.Fatal("missing warning message")
			}

			if actual, expected := buf.String(), tc.expectedOut; !strings.Contains(actual, expected) {
				t.Fatalf("expected to contain %q, got %q", expected, actual)
			}

			if actual, expected := fmt.Sprint(err), fmt.Sprint(tc.expectedErr); actual != expected {
				t.Fatalf("expected %s, got %s", expected, actual)
			}
		})
	}
}
