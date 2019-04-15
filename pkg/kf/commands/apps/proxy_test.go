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

package apps

import (
	"bytes"
	"errors"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/fake"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	"github.com/golang/mock/gomock"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

func TestNewProxyCommand(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		Namespace       string
		Args            []string
		ExpectedStrings []string
		ExpectedErr     error
		Setup           func(t *testing.T, lister *fake.FakeLister, istio *fake.FakeIstioClient)
	}{
		"no app name": {
			Namespace:   "default",
			Args:        []string{},
			ExpectedErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"minimal configuration": {
			Namespace:   "default",
			Args:        []string{"my-app", "--no-start=true"},
			ExpectedErr: nil,
			Setup: func(t *testing.T, lister *fake.FakeLister, istio *fake.FakeIstioClient) {
				istio.EXPECT().ListIngresses(gomock.Any()).Return([]corev1.LoadBalancerIngress{{IP: "8.8.8.8"}}, nil)
				lister.EXPECT().List(gomock.Any()).Return([]serving.Service{{}}, nil)
			},
		},
		"autodetect failure": {
			Namespace:   "default",
			Args:        []string{"my-app"},
			ExpectedErr: errors.New("istio-failure"),
			Setup: func(t *testing.T, lister *fake.FakeLister, istio *fake.FakeIstioClient) {
				istio.EXPECT().ListIngresses(gomock.Any()).Return(nil, errors.New("istio-failure"))
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeLister := fake.NewFakeLister(ctrl)
			fakeIstio := fake.NewFakeIstioClient(ctrl)

			if tc.Setup != nil {
				tc.Setup(t, fakeLister, fakeIstio)
			}

			buf := new(bytes.Buffer)
			p := &config.KfParams{
				Output:    buf,
				Namespace: tc.Namespace,
			}

			cmd := NewProxyCommand(p, fakeLister, fakeIstio)
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.Args)
			_, actualErr := cmd.ExecuteC()
			if tc.ExpectedErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedErr, actualErr)
				return
			}

			testutil.AssertContainsAll(t, buf.String(), tc.ExpectedStrings)
			testutil.AssertEqual(t, "SilenceUsage", true, cmd.SilenceUsage)

			ctrl.Finish()
		})
	}
}
