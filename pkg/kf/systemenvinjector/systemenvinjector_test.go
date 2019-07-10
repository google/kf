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

package systemenvinjector

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/internal/envutil"
	servicebindings "github.com/google/kf/pkg/kf/service-bindings"
	fakebindings "github.com/google/kf/pkg/kf/service-bindings/fake"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestSystemEnvInjector(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		setup     func(app *v1alpha1.App, fake *fakebindings.FakeClientInterface)
		expectErr error
		validate  func(t *testing.T, env map[string]string)
	}{
		"new-service": {
			setup: func(app *v1alpha1.App, fake *fakebindings.FakeClientInterface) {
				app.Name = "foo"
				app.Namespace = "ns"

				fake.EXPECT().GetVcapServices("foo", gomock.Any()).Return(servicebindings.VcapServicesMap{}, nil)

			},
			validate: func(t *testing.T, env map[string]string) {
				testutil.AssertEqual(t, "env count", 2, len(env))
				if _, ok := env["VCAP_SERVICES"]; !ok {
					t.Fatal("Expected map to contain VCAP_SERVICES")
				}

				if _, ok := env["VCAP_APPLICATION"]; !ok {
					t.Fatal("Expected map to contain VCAP_APPLICATION")
				}
			},
		},
		"lookup failure": {
			setup: func(app *v1alpha1.App, fake *fakebindings.FakeClientInterface) {
				app.Name = "foo"
				app.Namespace = "ns"

				fake.EXPECT().GetVcapServices("foo", gomock.Any()).Return(nil, errors.New("test"))
			},
			expectErr: errors.New("test"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeClient := fakebindings.NewFakeClientInterface(ctrl)
			svc := &v1alpha1.App{}

			if tc.setup != nil {
				tc.setup(svc, fakeClient)
			}

			injector := NewSystemEnvInjector(fakeClient)
			actualErr := injector.InjectSystemEnv(svc)

			if tc.expectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.expectErr, actualErr)
				return
			}

			tc.validate(t, envutil.EnvVarsToMap(envutil.GetAppEnvVars(svc)))
		})
	}
}
