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

package servicebindings_test

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/spf13/cobra"
)

type commandFactory func(p *config.KfParams) *cobra.Command

type serviceTest struct {
	Args  []string
	Space string

	ExpectedErr     error
	ExpectedStrings []string
}

func runTest(t *testing.T, tc serviceTest, newCommand commandFactory) {
	gomock.NewController(t)

	buf := new(bytes.Buffer)
	p := &config.KfParams{
		Space: tc.Space,
	}

	cmd := newCommand(p)
	cmd.SetOutput(buf)
	cmd.SetArgs(tc.Args)
	_, actualErr := cmd.ExecuteC()
	if tc.ExpectedErr != nil || actualErr != nil {
		testutil.AssertErrorsEqual(t, tc.ExpectedErr, actualErr)
		return
	}

	testutil.AssertContainsAll(t, buf.String(), tc.ExpectedStrings)
}
