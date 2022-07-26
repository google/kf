// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spaces

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	configlogging "github.com/google/kf/v2/pkg/kf/commands/config/logging"
	"github.com/google/kf/v2/pkg/kf/spaces/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestNewDomainsCommand(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		wantErr        error
		args           []string
		setup          func(t *testing.T, fakeSpaces *fake.FakeClient)
		validateOutput bool
	}{
		"invalid number of args": {
			args:    []string{"some-arg"},
			wantErr: errors.New("accepts 0 arg(s), received 1"),
		},
		"calls get": {
			args: []string{},
			setup: func(t *testing.T, fakeSpaces *fake.FakeClient) {
				space := &v1alpha1.Space{}
				space.Status.NetworkConfig.Domains = []v1alpha1.SpaceDomain{
					{Domain: "test.example.com", GatewayName: "kf/external-gateway"},
					{Domain: "kf.internal", GatewayName: "kf/internal-gateway"},
				}

				fakeSpaces.EXPECT().Get(gomock.Any(), "default").Return(space, nil)
			},
			validateOutput: true,
		},
		"server failure": {
			args: []string{},
			setup: func(t *testing.T, fakeSpaces *fake.FakeClient) {
				fakeSpaces.
					EXPECT().
					Get(gomock.Any(), "default").
					Return(nil, errors.New("some-server-error"))
			},
			wantErr: errors.New("failed to get Space: some-server-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeSpaces := fake.NewFakeClient(ctrl)

			if tc.setup != nil {
				tc.setup(t, fakeSpaces)
			}

			buffer := new(bytes.Buffer)
			ctx := configlogging.SetupLogger(context.Background(), buffer)

			c := NewDomainsCommand(&config.KfParams{Space: "default"}, fakeSpaces)
			c.SetOutput(buffer)
			c.SetArgs(tc.args)
			c.SetContext(ctx)

			gotErr := c.Execute()
			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)

			if tc.validateOutput {
				testutil.AssertGolden(t, "domains_out", buffer.Bytes())
			}

		})
	}
}
