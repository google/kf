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

package servicebindings_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	servicebindingscmd "github.com/google/kf/v2/pkg/kf/commands/service-bindings"
	injection "github.com/google/kf/v2/pkg/kf/injection/fake"
	secretsfake "github.com/google/kf/v2/pkg/kf/secrets/fake"
	serviceinstancebindingsfake "github.com/google/kf/v2/pkg/kf/serviceinstancebindings/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewBindRouteServiceCommand(t *testing.T) {
	type fakes struct {
		servicebindings *serviceinstancebindingsfake.FakeClient
		secrets         *secretsfake.FakeClient
	}
	cases := map[string]struct {
		args                  []string
		namespace             string
		setup                 func(*testing.T, fakes)
		expectedErr           error
		routeServicesDisabled bool // default to false
	}{
		"wrong number of args": {
			args:        []string{},
			expectedErr: errors.New("accepts 2 arg(s), received 0"),
		},
		"command params get passed correctly": {
			args:      []string{"domain.com", "--hostname=myhost", "SERVICE_INSTANCE", "--path=somepath", `-c={"some":"json"}`},
			namespace: "custom-ns",
			setup: func(t *testing.T, fakes fakes) {
				bindingName := v1alpha1.MakeRouteServiceBindingName("myhost", "domain.com", "somepath", "SERVICE_INSTANCE")
				secretName := v1alpha1.MakeRouteServiceBindingParamsSecretName("myhost", "domain.com", "somepath", "SERVICE_INSTANCE")
				fakes.servicebindings.EXPECT().Create(gomock.Any(), "custom-ns", &v1alpha1.ServiceInstanceBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      bindingName,
						Namespace: "custom-ns",
					},
					Spec: v1alpha1.ServiceInstanceBindingSpec{
						BindingType: v1alpha1.BindingType{
							Route: &v1alpha1.RouteRef{
								Hostname: "myhost",
								Domain:   "domain.com",
								Path:     "somepath",
							},
						},
						InstanceRef: v1.LocalObjectReference{
							Name: "SERVICE_INSTANCE",
						},
						ParametersFrom: v1.LocalObjectReference{
							Name: secretName,
						},
					},
				})

				fakes.secrets.EXPECT().CreateParamsSecret(gomock.Any(), gomock.Any(), secretName, json.RawMessage(`{"some":"json"}`))
				fakes.servicebindings.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "custom-ns",
					bindingName, gomock.Any())
			},
		},
		"empty namespace": {
			args:        []string{"domain.com", "SERVICE_INSTANCE"},
			expectedErr: errors.New(config.EmptySpaceError),
		},
		"defaults config": {
			args:      []string{"domain.com", "SERVICE_INSTANCE"},
			namespace: "custom-ns",
			setup: func(t *testing.T, fakes fakes) {
				bindingName := v1alpha1.MakeRouteServiceBindingName("", "domain.com", "", "SERVICE_INSTANCE")
				secretName := v1alpha1.MakeRouteServiceBindingParamsSecretName("", "domain.com", "", "SERVICE_INSTANCE")
				fakes.servicebindings.EXPECT().Create(gomock.Any(), "custom-ns", &v1alpha1.ServiceInstanceBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      bindingName,
						Namespace: "custom-ns",
					},
					Spec: v1alpha1.ServiceInstanceBindingSpec{
						BindingType: v1alpha1.BindingType{
							Route: &v1alpha1.RouteRef{
								Domain: "domain.com",
							},
						},
						InstanceRef: v1.LocalObjectReference{
							Name: "SERVICE_INSTANCE",
						},
						ParametersFrom: v1.LocalObjectReference{
							Name: secretName,
						},
					},
				})
				fakes.secrets.EXPECT().CreateParamsSecret(gomock.Any(), gomock.Any(), secretName, json.RawMessage("{}"))
				fakes.servicebindings.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "custom-ns",
					bindingName, gomock.Any())
			},
		},
		"bad config path": {
			args:        []string{"domain.com", "SERVICE_INSTANCE", `-c=/some/bad/path`},
			namespace:   "custom-ns",
			expectedErr: errors.New("couldn't read file: open /some/bad/path: no such file or directory"),
		},
		"bad server call": {
			args:      []string{"domain.com", "SERVICE_INSTANCE"},
			namespace: "custom-ns",
			setup: func(t *testing.T, fakes fakes) {
				fakes.servicebindings.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("api-error"))
			},
			expectedErr: errors.New("api-error"),
		},
		"async": {
			args:      []string{"--async", "domain.com", "SERVICE_INSTANCE"},
			namespace: "default",
			setup: func(t *testing.T, fakes fakes) {
				fakes.secrets.EXPECT().CreateParamsSecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
				fakes.servicebindings.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"failed binding": {
			args:      []string{"domain.com", "SERVICE_INSTANCE"},
			namespace: "custom-ns",
			setup: func(t *testing.T, fakes fakes) {
				fakes.secrets.EXPECT().CreateParamsSecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
				fakes.servicebindings.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any())
				fakes.servicebindings.EXPECT().WaitForConditionReadyTrue(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("binding already exists"))
			},
			expectedErr: errors.New("bind failed: binding already exists"),
		},
		"route services disabled": {
			args:                  []string{"domain.com", "SERVICE_INSTANCE"},
			namespace:             "custom-ns",
			routeServicesDisabled: true,
			expectedErr:           errors.New(`Route services feature is toggled off. Set "enable_route_services" to true in "config-defaults" to enable route services`),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			sbClient := serviceinstancebindingsfake.NewFakeClient(ctrl)
			secretClient := secretsfake.NewFakeClient(ctrl)

			if tc.setup != nil {
				tc.setup(t, fakes{
					servicebindings: sbClient,
					secrets:         secretClient,
				})
			}

			buf := new(bytes.Buffer)
			p := &config.KfParams{
				Space: tc.namespace,
			}

			ctx := injection.WithInjection(context.Background(), t)
			ff := make(kfconfig.FeatureFlagToggles)
			ff.SetRouteServices(!tc.routeServicesDisabled)
			ctx = testutil.WithFeatureFlags(ctx, t, ff)

			cmd := servicebindingscmd.NewBindRouteServiceCommand(p, sbClient, secretClient)
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.args)
			cmd.SetContext(ctx)
			_, actualErr := cmd.ExecuteC()
			testutil.AssertErrorsEqual(t, tc.expectedErr, actualErr)
		})
	}
}
