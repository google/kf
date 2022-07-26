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
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/doctor"
	"github.com/google/kf/v2/pkg/kf/doctor/troubleshooter"
	"github.com/google/kf/v2/pkg/kf/internal/genericcli"
	"github.com/google/kf/v2/pkg/kf/testutil"
)

type failDiagnostic struct{}

func (fd failDiagnostic) Diagnose(ctx context.Context, d *doctor.Diagnostic) {
	d.Error("test-fail")
}

type okayDiagnostic struct{}

func (fd okayDiagnostic) Diagnose(ctx context.Context, d *doctor.Diagnostic) {
}

type countingDiagnostic struct {
	count int
}

func (cd *countingDiagnostic) Diagnose(ctx context.Context, d *doctor.Diagnostic) {
	cd.count++
	d.Error(fmt.Sprintf("count: %d", cd.count))
}

type fakeCloser struct {
	diagnosable doctor.Diagnosable
}

func (f *fakeCloser) Close(component troubleshooter.Component, resourceName string) doctor.Diagnosable {
	return f.diagnosable
}

func TestNewDoctorCommand(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		space          string
		wantErr        error
		args           []string
		diagnostics    []DoctorTest
		components     []troubleshooter.Component
		closer         ComponentCloser
		expectedOutput []string
	}{
		"bad diagnostic": {
			args: []string{"invalid2"},
			diagnostics: []DoctorTest{
				{Name: "failer", Test: failDiagnostic{}},
			},
			wantErr: errors.New(`unknown component "invalid2", supported components: failer`),
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
		"retries flag": {
			args: []string{"retrier", "--retries", "3", "--delay", "10ms"},
			diagnostics: []DoctorTest{
				{Name: "retrier", Test: &countingDiagnostic{}},
			},
			expectedOutput: []string{"doctor/retrier", "FAIL", "count: 3"},
			wantErr:        errors.New(`environment failed checks`),
		},
		"malformed argument": {
			args:    []string{"app/foo/bar"},
			wantErr: errors.New(`malformed argument: "app/foo/bar"`),
		},
		"unknown object type": {
			args: []string{"domain/foo"},
			components: []troubleshooter.Component{
				{
					Type: &genericcli.KubernetesType{
						KfName: "Source",
					},
				},
				{
					Type: &genericcli.KubernetesType{
						KfName: "App",
					},
				},
			},
			wantErr: errors.New(`unknown object type "domain", supported object types: app|source`),
		},
		"successful component": {
			args: []string{"app/foo"},
			components: []troubleshooter.Component{
				{
					Type: &genericcli.KubernetesType{
						KfName: "App",
					},
				},
			},
			closer: &fakeCloser{diagnosable: &okayDiagnostic{}},
		},
		"unsuccessful component": {
			args: []string{"app/foo"},
			components: []troubleshooter.Component{
				{
					Type: &genericcli.KubernetesType{
						KfName: "App",
					},
				},
			},
			closer:  &fakeCloser{diagnosable: &failDiagnostic{}},
			wantErr: errors.New(`environment failed checks`),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			buffer := &bytes.Buffer{}

			c := NewDoctorCommand(
				&config.KfParams{Space: tc.space},
				tc.diagnostics,
				tc.components,
				tc.closer,
			)

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
