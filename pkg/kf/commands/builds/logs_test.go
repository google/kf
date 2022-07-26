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

package builds

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/v2/pkg/kf/builds/fake"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestNewBuildLogsCommand(t *testing.T) {
	t.Parallel()

	// Setup a context that has a value so that we can assert on the fact that
	// we are using the right context.
	someContext := context.WithValue(context.Background(), "some-key", "some-value")

	cases := map[string]struct {
		args      []string
		namespace string
		setup     func(t *testing.T, fakeSources *fake.FakeClient)

		wantErr         error
		expectedStrings []string
	}{
		"invalid number of args": {
			args:    []string{},
			wantErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"missing namespace": {
			args:    []string{"my-build"},
			wantErr: errors.New(config.EmptySpaceError),
		},
		"calls with right args": {
			args:      []string{"my-build"},
			namespace: "my-ns",
			setup: func(t *testing.T, fakeSources *fake.FakeClient) {
				fakeSources.
					EXPECT().
					Tail(someContext, "my-ns", "my-build", gomock.Any()).
					Return(nil)
			},
		},
		"writer goes to stdout": {
			args:      []string{"my-build"},
			namespace: "my-ns",
			setup: func(t *testing.T, fakeSources *fake.FakeClient) {
				fakeSources.
					EXPECT().
					Tail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(_ context.Context, ns, name string, out io.Writer) {
						fmt.Fprintln(out, "LOG STREAM")
					}).
					Return(nil)
			},
			expectedStrings: []string{"LOG STREAM"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeSources := fake.NewFakeClient(ctrl)

			if tc.setup != nil {
				tc.setup(t, fakeSources)
			}

			buffer := &bytes.Buffer{}

			c := NewBuildLogsCommand(&config.KfParams{Space: tc.namespace}, fakeSources)
			c.SetOutput(buffer)
			c.SetArgs(tc.args)
			c.SetContext(someContext)

			gotErr := c.Execute()
			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
			testutil.AssertContainsAll(t, buffer.String(), tc.expectedStrings)

		})
	}
}
