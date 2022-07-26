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
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	servicebindingscmd "github.com/google/kf/v2/pkg/kf/commands/service-bindings"
	injection "github.com/google/kf/v2/pkg/kf/injection/fake"
	serviceinstancebindingsfake "github.com/google/kf/v2/pkg/kf/serviceinstancebindings/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestNewUnbindRouteServiceCommand(t *testing.T) {
	type fakes struct {
		servicebindings *serviceinstancebindingsfake.FakeClient
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
			args:      []string{"domain.com", "--hostname=myhost", "--path=mypath", "SERVICE_INSTANCE"},
			namespace: "custom-ns",
			setup: func(t *testing.T, fakes fakes) {
				bindingName := v1alpha1.MakeRouteServiceBindingName("myhost", "domain.com", "mypath", "SERVICE_INSTANCE")
				fakes.servicebindings.EXPECT().Delete(gomock.Any(), "custom-ns", bindingName)
				fakes.servicebindings.EXPECT().WaitForDeletion(gomock.Any(), "custom-ns", bindingName, gomock.Any())
			},
		},
		"some optional params empty": {
			args:      []string{"domain.com", "--hostname=myhost", "SERVICE_INSTANCE"},
			namespace: "custom-ns",
			setup: func(t *testing.T, fakes fakes) {
				bindingName := v1alpha1.MakeRouteServiceBindingName("myhost", "domain.com", "", "SERVICE_INSTANCE")
				fakes.servicebindings.EXPECT().Delete(gomock.Any(), "custom-ns", bindingName)
				fakes.servicebindings.EXPECT().WaitForDeletion(gomock.Any(), "custom-ns", bindingName, gomock.Any())
			},
		},
		"empty namespace": {
			args:        []string{"domain.com", "SERVICE_INSTANCE"},
			expectedErr: errors.New(config.EmptySpaceError),
		},
		"bad server call": {
			args:      []string{"domain.com", "SERVICE_INSTANCE"},
			namespace: "custom-ns",
			setup: func(t *testing.T, fakes fakes) {
				fakes.servicebindings.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("api-error"))
			},
			expectedErr: errors.New("api-error"),
		},
		"async": {
			args:      []string{"--async", "domain.com", "SERVICE_INSTANCE"},
			namespace: "default",
			setup: func(t *testing.T, fakes fakes) {
				fakes.servicebindings.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any())
			},
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

			if tc.setup != nil {
				tc.setup(t, fakes{
					servicebindings: sbClient,
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

			cmd := servicebindingscmd.NewUnbindRouteServiceCommand(p, sbClient)
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.args)
			cmd.SetContext(ctx)
			_, actualErr := cmd.ExecuteC()
			testutil.AssertErrorsEqual(t, tc.expectedErr, actualErr)
		})
	}
}
