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
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	servicescmd "github.com/google/kf/v2/pkg/kf/commands/services"
	injection "github.com/google/kf/v2/pkg/kf/injection/fake"
	secretsfake "github.com/google/kf/v2/pkg/kf/secrets/fake"
	serviceinstancesfake "github.com/google/kf/v2/pkg/kf/serviceinstances/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewCreateUserProvidedServiceCommand(t *testing.T) {
	type fakes struct {
		services *serviceinstancesfake.FakeClient
		secrets  *secretsfake.FakeClient
	}

	cases := map[string]struct {
		args                 []string
		namespace            string
		setup                func(*testing.T, fakes)
		expectErr            error
		routeServicesEnabled bool
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

		// good results
		"default": {
			namespace: "test-ns",
			args:      []string{"mydb", "-p", `{"username":"fake-user"}`},
			setup: func(t *testing.T, fakes fakes) {
				fakes.services.EXPECT().Create(gomock.Any(), "test-ns", &v1alpha1.ServiceInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mydb",
						Namespace: "test-ns",
					},
					Spec: v1alpha1.ServiceInstanceSpec{
						ServiceType: v1alpha1.ServiceType{
							UPS: &v1alpha1.UPSInstance{},
						},
						ParametersFrom: corev1.LocalObjectReference{
							Name: v1alpha1.GenerateName("serviceinstance", "mydb", "params"),
						},
						Tags: []string{},
					},
				})

				fakes.secrets.EXPECT().CreateParamsSecret(gomock.Any(), gomock.Any(), gomock.Any(), json.RawMessage(`{"username":"fake-user"}`))
				fakes.services.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "test-ns", "mydb", gomock.Any())
			},
		},

		"with tags": {
			namespace: "test-ns",
			args:      []string{"mydb", "-p", `{"username":"fake-user"}`, "-t", "tag1, tag2"},
			setup: func(t *testing.T, fakes fakes) {
				fakes.services.EXPECT().Create(gomock.Any(), "test-ns", &v1alpha1.ServiceInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mydb",
						Namespace: "test-ns",
					},
					Spec: v1alpha1.ServiceInstanceSpec{
						ServiceType: v1alpha1.ServiceType{
							UPS: &v1alpha1.UPSInstance{},
						},
						ParametersFrom: corev1.LocalObjectReference{
							Name: v1alpha1.GenerateName("serviceinstance", "mydb", "params"),
						},
						Tags: []string{"tag1", "tag2"},
					},
				})

				fakes.secrets.EXPECT().CreateParamsSecret(gomock.Any(), gomock.Any(), gomock.Any(), json.RawMessage(`{"username":"fake-user"}`))
				fakes.services.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "test-ns", "mydb", gomock.Any())
			},
		},

		"route services not enabled": {
			namespace:            "test-ns",
			args:                 []string{"some-rs", "-r", "http://example-rs.com"},
			routeServicesEnabled: false,
			expectErr:            errors.New(`Route services feature is toggled off. Set "enable_route_services" to true in "config-defaults" to enable route services`),
		},

		"route services enabled": {
			namespace:            "test-ns",
			args:                 []string{"some-rs", "-r", "http://example-rs.com"},
			routeServicesEnabled: true,
			setup: func(t *testing.T, fakes fakes) {
				parsedURL, _ := v1alpha1.ParseURL("http://example-rs.com")
				expected := &v1alpha1.ServiceInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-rs",
						Namespace: "test-ns",
					},
					Spec: v1alpha1.ServiceInstanceSpec{
						ServiceType: v1alpha1.ServiceType{
							UPS: &v1alpha1.UPSInstance{
								RouteServiceURL: parsedURL,
							},
						},
						ParametersFrom: corev1.LocalObjectReference{
							Name: v1alpha1.GenerateName("serviceinstance", "some-rs", "params"),
						},
						Tags: []string{},
					},
				}

				fakes.services.EXPECT().Create(gomock.Any(), "test-ns", gomock.Any()).Do(func(ctx context.Context, ns string, si *v1alpha1.ServiceInstance) {
					testutil.AssertEqual(t, "route service", expected, si)
				})

				fakes.secrets.EXPECT().CreateParamsSecret(gomock.Any(), gomock.Any(), gomock.Any(), json.RawMessage(`{}`))
				fakes.services.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "test-ns", "some-rs", gomock.Any())
			},
		},

		"async": {
			namespace: "test-ns",
			args:      []string{"mydb", "-p", `{"username":"fake-user"}`, "--async"},
			setup: func(t *testing.T, fakes fakes) {
				fakes.services.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any())
				fakes.secrets.EXPECT().CreateParamsSecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
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

			ctx := injection.WithInjection(context.Background(), t)
			ff := kfconfig.FeatureFlagToggles{}
			ff.SetRouteServices(tc.routeServicesEnabled)
			ctx = testutil.WithFeatureFlags(ctx, t, ff)

			cmd := servicescmd.NewCreateUserProvidedServiceCommand(p, sClient, secretClient)
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.args)
			cmd.SetContext(ctx)
			_, actualErr := cmd.ExecuteC()
			if tc.expectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.expectErr, actualErr)
			}
		})
	}
}
