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

package services_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	servicescmd "github.com/google/kf/v2/pkg/kf/commands/services"
	secretsfake "github.com/google/kf/v2/pkg/kf/secrets/fake"
	serviceinstancesfake "github.com/google/kf/v2/pkg/kf/serviceinstances/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	v1 "k8s.io/api/core/v1"
)

func TestNewUpdateUserProvidedServiceCommand(t *testing.T) {
	type fakes struct {
		services *serviceinstancesfake.FakeClient
		secrets  *secretsfake.FakeClient
	}

	validUserProvidedService := &v1alpha1.ServiceInstance{
		Spec: v1alpha1.ServiceInstanceSpec{
			ServiceType: v1alpha1.ServiceType{
				UPS: &v1alpha1.UPSInstance{},
			},
			ParametersFrom: v1.LocalObjectReference{
				Name: "params-secret",
			},
		},
		Status: v1alpha1.ServiceInstanceStatus{
			SecretName: "params-secret",
		},
	}

	validBrokeredService := &v1alpha1.ServiceInstance{
		Spec: v1alpha1.ServiceInstanceSpec{
			ServiceType: v1alpha1.ServiceType{
				Brokered: &v1alpha1.BrokeredInstance{
					ClassName:  "some-class",
					PlanName:   "some-plan",
					Namespaced: true,
				},
			},
		},
	}

	cases := map[string]struct {
		args      []string
		namespace string
		setup     func(*testing.T, fakes)
		expectErr error
	}{
		// user errors
		"bad number of args": {
			expectErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"bad namespace": {
			args:      []string{"mydb"},
			expectErr: errors.New(config.EmptySpaceError),
		},
		"bad path": {
			namespace: "test-ns",
			args:      []string{"mydb", "-p", "/some/bad/path"},
			expectErr: errors.New("couldn't read file: open /some/bad/path: no such file or directory"),
		},
		"wrong service type": {
			namespace: "test-ns",
			args:      []string{"mydb", "-p", `{"username":"fake-user"}`},
			setup: func(t *testing.T, fakes fakes) {
				fakes.services.EXPECT().Get(gomock.Any(), "test-ns", "mydb").Return(validBrokeredService, nil)
			},
			expectErr: errors.New("Service instance is not user-provided"),
		},

		// good results
		"only credentials": {
			namespace: "test-ns",
			args:      []string{"mydb", "-p", `{"username":"fake-user"}`},
			setup: func(t *testing.T, fakes fakes) {
				fakes.services.EXPECT().Get(gomock.Any(), "test-ns", "mydb").Return(validUserProvidedService, nil)
				fakes.secrets.EXPECT().UpdateParamsSecret(gomock.Any(), gomock.Any(), gomock.Any(), json.RawMessage(`{"username":"fake-user"}`))
				fakes.services.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "test-ns", "mydb", gomock.Any())
			},
		},
		"only tags": {
			namespace: "test-ns",
			args:      []string{"mydb", "-t", "tag1, tag2"},
			setup: func(t *testing.T, fakes fakes) {
				fakes.services.EXPECT().Get(gomock.Any(), "test-ns", "mydb").Return(validUserProvidedService, nil)
				fakes.services.EXPECT().Transform(gomock.Any(), "test-ns", "mydb", gomock.Any())
				fakes.services.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "test-ns", "mydb", gomock.Any())
			},
		},
		"credentials and tags": {
			namespace: "test-ns",
			args:      []string{"mydb", "-p", `{"username":"fake-user"}`, "-t", "tag1, tag2"},
			setup: func(t *testing.T, fakes fakes) {
				fakes.services.EXPECT().Get(gomock.Any(), "test-ns", "mydb").Return(validUserProvidedService, nil)
				fakes.services.EXPECT().Transform(gomock.Any(), "test-ns", "mydb", gomock.Any())
				fakes.secrets.EXPECT().UpdateParamsSecret(gomock.Any(), gomock.Any(), gomock.Any(), json.RawMessage(`{"username":"fake-user"}`))
				fakes.services.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "test-ns", "mydb", gomock.Any())
			},
		},

		"async": {
			namespace: "test-ns",
			args:      []string{"mydb", "-p", `{"username":"fake-user"}`, "--async"},
			setup: func(t *testing.T, fakes fakes) {
				fakes.services.EXPECT().Get(gomock.Any(), "test-ns", "mydb").Return(validUserProvidedService, nil)
				fakes.secrets.EXPECT().UpdateParamsSecret(gomock.Any(), gomock.Any(), gomock.Any(), json.RawMessage(`{"username":"fake-user"}`))
				// expect WaitForConditionReadyTrue not to be called
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			sClient := serviceinstancesfake.NewFakeClient(ctrl)
			secretClient := secretsfake.NewFakeClient(ctrl)

			if tc.setup != nil {
				tc.setup(t, fakes{
					services: sClient,
					secrets:  secretClient,
				})
			}

			buf := new(bytes.Buffer)
			p := &config.KfParams{
				Space: tc.namespace,
			}

			cmd := servicescmd.NewUpdateUserProvidedServiceCommand(p, sClient, secretClient)
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.args)
			_, actualErr := cmd.ExecuteC()
			if tc.expectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.expectErr, actualErr)
			}
		})
	}
}
