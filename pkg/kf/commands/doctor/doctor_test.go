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

package doctor

import (
	"bytes"
	"errors"
	"testing"

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/doctor"
	"github.com/google/kf/pkg/kf/testutil"
)

type failDiagnostic struct{}

func (fd failDiagnostic) Diagnose(d *doctor.Diagnostic) {
	d.Error("test-fail")
}

type okayDiagnostic struct{}

func (fd okayDiagnostic) Diagnose(d *doctor.Diagnostic) {
}

func TestNewDoctorCommand(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		namespace      string
		wantErr        error
		args           []string
		diagnostics    []DoctorTest
		expectedOutput []string
	}{
		"bad diagnostic": {
			args: []string{"invalid2"},
			diagnostics: []DoctorTest{
				{Name: "failer", Test: failDiagnostic{}},
			},
			wantErr: errors.New(`invalid argument "invalid2" for "doctor"`),
		},
		"runs only chosen tests": {
			args: []string{"passer"},
			diagnostics: []DoctorTest{
				{Name: "failer", Test: failDiagnostic{}},
				{Name: "passer", Test: okayDiagnostic{}},
			},
		},
		"runs all tests for no args": {
			args: []string{},
			diagnostics: []DoctorTest{
				{Name: "first", Test: okayDiagnostic{}},
				{Name: "second", Test: okayDiagnostic{}},
				{Name: "third", Test: okayDiagnostic{}},
			},
			expectedOutput: []string{"doctor/first", "doctor/second", "doctor/third", "PASS"},
		},
		"failure state": {
			args: []string{},
			diagnostics: []DoctorTest{
				{Name: "failer", Test: failDiagnostic{}},
			},
			expectedOutput: []string{"doctor/failer", "FAIL"},
			wantErr:        errors.New(`environment failed checks`),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			buffer := &bytes.Buffer{}

			c := NewDoctorCommand(&config.KfParams{
				Namespace: tc.namespace,
			}, tc.diagnostics)

			c.SetOutput(buffer)
			c.SetArgs(tc.args)
			gotErr := c.Execute()
			if tc.wantErr != nil || gotErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
			}

			testutil.AssertContainsAll(t, buffer.String(), tc.expectedOutput)
		})
	}
}
