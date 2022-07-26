// Copyright 2020 Google LLC
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

package genericcli

import (
	"bytes"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestNewStubCommand_docs(t *testing.T) {
	cases := map[string]struct {
		name    string
		short   string
		alt     string
		example string
		opts    []StubOption

		wantUse     string
		wantShort   string
		wantLong    string
		wantAliases []string
		wantExample string
	}{
		"With Alias": {
			name:    "login",
			short:   "Allows a user to login",
			alt:     "TLDR: gcloud auth login",
			example: "  gcloud auth login",
			opts: []StubOption{
				WithStubAliases([]string{"l"}),
			},
			wantUse:     "login",
			wantShort:   "Allows a user to login",
			wantLong:    "kf login is currently unsupported.\nTLDR: gcloud auth login\n",
			wantAliases: []string{"l"},
			wantExample: "  gcloud auth login",
		},
		"WithoutAlias": {
			name:    "api",
			short:   "Set or view target api url",
			alt:     "TLDR: kubectl config current-context",
			example: "  kubectl config current-context",

			wantUse:     "api",
			wantShort:   "Set or view target api url",
			wantLong:    "kf api is currently unsupported.\nTLDR: kubectl config current-context\n",
			wantAliases: []string{},
			wantExample: "  kubectl config current-context",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			cmd := NewStubCommand(tc.name, tc.short, tc.alt, tc.example, tc.opts...)
			testutil.AssertEqual(t, "use", tc.wantUse, cmd.Use)
			testutil.AssertEqual(t, "short", tc.wantShort, cmd.Short)
			testutil.AssertEqual(t, "long", tc.wantLong, cmd.Long)
			testutil.AssertEqual(t, "aliases", tc.wantAliases, cmd.Aliases)
			testutil.AssertEqual(t, "example", tc.wantExample, cmd.Example)
		})
	}
}

func TestNewStubCommand(t *testing.T) {
	cases := map[string]struct {
		name string
		alt  string

		args    []string
		wantOut string
		wantErr error
	}{
		"no args": {
			name: "logout",
			alt:  "TLDR: gcloud auth revoke --all",
			args: []string{},
			wantOut: `kf logout is currently unsupported.
TLDR: gcloud auth revoke --all

`,
		},
		"with args": {
			name: "auth",
			alt:  "TLDR: gcloud auth activate-service-account",
			args: []string{"admin", "password", "--origin", "oicd"},
			wantOut: `kf auth is currently unsupported.
TLDR: gcloud auth activate-service-account

`,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			buf := new(bytes.Buffer)
			cmd := NewStubCommand(tc.name, "short", tc.alt, "example")
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.args)

			_, actualErr := cmd.ExecuteC()
			testutil.AssertErrorsEqual(t, tc.wantErr, actualErr)
			testutil.AssertEqual(t, "output", buf.String(), tc.wantOut)

			// This Cobra classification is necessary to prevent rendering CLI docs
			// on these topics we don't support. Instead they're "Additional Topics"
			testutil.AssertTrue(t, "IsAdditionalHelpTopicCommand", cmd.IsAdditionalHelpTopicCommand())
		})
	}
}
