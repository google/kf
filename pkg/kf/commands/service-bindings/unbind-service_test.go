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
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	servicebindingscmd "github.com/google/kf/v2/pkg/kf/commands/service-bindings"
	"github.com/google/kf/v2/pkg/kf/serviceinstancebindings"
	serviceinstancebindingsfake "github.com/google/kf/v2/pkg/kf/serviceinstancebindings/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func runUnbindTest(t *testing.T, tc bindingTest) {
	ctrl := gomock.NewController(t)

	sbClient := serviceinstancebindingsfake.NewFakeClient(ctrl)

	if tc.Setup != nil {
		tc.Setup(t, fakes{
			servicebindings: sbClient,
		})
	}

	buf := new(bytes.Buffer)
	p := &config.KfParams{
		Space: tc.Space,
	}

	cmd := servicebindingscmd.NewUnbindServiceCommand(p, sbClient)
	cmd.SetOutput(buf)
	cmd.SetArgs(tc.Args)
	_, actualErr := cmd.ExecuteC()
	if tc.ExpectedErr != nil || actualErr != nil {
		testutil.AssertErrorsEqual(t, tc.ExpectedErr, actualErr)
		return
	}

	testutil.AssertContainsAll(t, buf.String(), tc.ExpectedStrings)
}

func TestNewUnbindServiceCommand(t *testing.T) {
	cases := map[string]bindingTest{
		"wrong number of args": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 2 arg(s), received 0"),
		},
		"command params get passed correctly": {
			Args:  []string{"APP_NAME", "SERVICE_INSTANCE"},
			Space: "custom-ns",
			Setup: func(t *testing.T, fakes fakes) {
				bindingName := v1alpha1.MakeServiceBindingName("APP_NAME", "SERVICE_INSTANCE")
				binding := &v1alpha1.ServiceInstanceBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      bindingName,
						Namespace: "custom-ns",
					},
					Spec: v1alpha1.ServiceInstanceBindingSpec{
						BindingType: v1alpha1.BindingType{
							App: &v1alpha1.AppRef{
								Name: "APP_NAME",
							},
						},
						InstanceRef: v1.LocalObjectReference{
							Name: "SERVICE_INSTANCE",
						},
					},
					Status: v1alpha1.ServiceInstanceBindingStatus{
						OSBStatus: v1alpha1.BindingOSBStatus{
							UnbindFailed: &v1alpha1.OSBState{},
						},
					},
				}
				fakes.servicebindings.EXPECT().
					Transform(gomock.Any(), "custom-ns", bindingName, gomock.Any()).
					Do(func(_ context.Context, _, _ string, m serviceinstancebindings.Mutator) {
						testutil.AssertNil(t, "mutator error", m(binding))
						testutil.AssertTrue(t, "binding.spec.UnbindRequests", binding.Spec.UnbindRequests == 1)
					})
				fakes.servicebindings.EXPECT().Delete(gomock.Any(), "custom-ns", bindingName)
				fakes.servicebindings.EXPECT().WaitForDeletion(gomock.Any(), "custom-ns", bindingName, gomock.Any())
			},
		},
		"empty namespace": {
			Args:        []string{"APP_NAME", "SERVICE_INSTANCE"},
			ExpectedErr: errors.New(config.EmptySpaceError),
		},
		"bad server call": {
			Args:  []string{"APP_NAME", "SERVICE_INSTANCE"},
			Space: "custom-ns",
			Setup: func(t *testing.T, fakes fakes) {
				fakes.servicebindings.EXPECT().Transform(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
				fakes.servicebindings.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("api-error"))
			},
			ExpectedErr: errors.New("api-error"),
		},
		"async": {
			Args:  []string{"--async", "APP_NAME", "SERVICE_INSTANCE"},
			Space: "custom-ns",
			Setup: func(t *testing.T, fakes fakes) {
				bindingName := v1alpha1.MakeServiceBindingName("APP_NAME", "SERVICE_INSTANCE")
				binding := &v1alpha1.ServiceInstanceBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      bindingName,
						Namespace: "custom-ns",
					},
					Spec: v1alpha1.ServiceInstanceBindingSpec{
						BindingType: v1alpha1.BindingType{
							App: &v1alpha1.AppRef{
								Name: "APP_NAME",
							},
						},
						InstanceRef: v1.LocalObjectReference{
							Name: "SERVICE_INSTANCE",
						},
					},
					Status: v1alpha1.ServiceInstanceBindingStatus{
						OSBStatus: v1alpha1.BindingOSBStatus{
							UnbindFailed: &v1alpha1.OSBState{},
						},
					},
				}
				fakes.servicebindings.EXPECT().
					Transform(gomock.Any(), "custom-ns", bindingName, gomock.Any()).
					Do(func(_ context.Context, _, _ string, m serviceinstancebindings.Mutator) {
						testutil.AssertNil(t, "mutator error", m(binding))
						testutil.AssertTrue(t, "binding.spec.UnbindRequests", binding.Spec.UnbindRequests == 1)
					})
				fakes.servicebindings.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			runUnbindTest(t, tc)
		})
	}
}
