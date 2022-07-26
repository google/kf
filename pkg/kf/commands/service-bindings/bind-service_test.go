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
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	appsfake "github.com/google/kf/v2/pkg/kf/apps/fake"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	configlogging "github.com/google/kf/v2/pkg/kf/commands/config/logging"
	servicebindingscmd "github.com/google/kf/v2/pkg/kf/commands/service-bindings"
	secretsfake "github.com/google/kf/v2/pkg/kf/secrets/fake"
	serviceinstancebindingsfake "github.com/google/kf/v2/pkg/kf/serviceinstancebindings/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/ptr"
)

type fakes struct {
	servicebindings *serviceinstancebindingsfake.FakeClient
	secrets         *secretsfake.FakeClient
	apps            *appsfake.FakeClient
}

type bindingTest struct {
	Args  []string
	Setup func(*testing.T, fakes)
	Space string

	ExpectedErr     error
	ExpectedStrings []string
}

func runBindingTest(t *testing.T, tc bindingTest) {
	ctrl := gomock.NewController(t)

	sbClient := serviceinstancebindingsfake.NewFakeClient(ctrl)
	secretClient := secretsfake.NewFakeClient(ctrl)
	appsClient := appsfake.NewFakeClient(ctrl)

	if tc.Setup != nil {
		tc.Setup(t, fakes{
			servicebindings: sbClient,
			secrets:         secretClient,
			apps:            appsClient,
		})
	}

	buf := new(bytes.Buffer)
	p := &config.KfParams{
		Space: tc.Space,
	}
	ctx := configlogging.SetupLogger(context.Background(), buf)

	cmd := servicebindingscmd.NewBindServiceCommand(p, sbClient, secretClient, appsClient)
	cmd.SetOutput(buf)
	cmd.SetArgs(tc.Args)
	cmd.SetContext(ctx)
	_, actualErr := cmd.ExecuteC()
	if tc.ExpectedErr != nil || actualErr != nil {
		testutil.AssertErrorsEqual(t, tc.ExpectedErr, actualErr)
		return
	}

	testutil.AssertContainsAll(t, buf.String(), tc.ExpectedStrings)
}

func TestNewBindServiceCommand(t *testing.T) {
	sampleApp := &v1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name: "APP_NAME",
		},
		Spec: v1alpha1.AppSpec{},
	}
	ownerRefs := []metav1.OwnerReference{
		{
			APIVersion:         "kf.dev/v1alpha1",
			Kind:               "App",
			Name:               "APP_NAME",
			Controller:         ptr.Bool(true),
			BlockOwnerDeletion: ptr.Bool(true),
		},
	}

	cases := map[string]bindingTest{
		"wrong number of args": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 2 arg(s), received 0"),
		},
		"command params get passed correctly": {
			Args:  []string{"APP_NAME", "SERVICE_INSTANCE", `-c={"ram_gb":4}`, "--binding-name=BINDING_NAME", "--timeout=30s"},
			Space: "custom-ns",
			Setup: func(t *testing.T, fakes fakes) {
				bindingName := v1alpha1.MakeServiceBindingName("APP_NAME", "SERVICE_INSTANCE")
				secretName := v1alpha1.MakeServiceBindingParamsSecretName("APP_NAME", "SERVICE_INSTANCE")
				fakes.servicebindings.EXPECT().Create(gomock.Any(), "custom-ns", &v1alpha1.ServiceInstanceBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:            bindingName,
						Namespace:       "custom-ns",
						OwnerReferences: ownerRefs,
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
						ParametersFrom: v1.LocalObjectReference{
							Name: secretName,
						},
						BindingNameOverride:     "BINDING_NAME",
						ProgressDeadlineSeconds: 30,
					},
				})

				fakes.secrets.EXPECT().CreateParamsSecret(gomock.Any(), gomock.Any(), secretName, json.RawMessage(`{"ram_gb":4}`))
				fakes.servicebindings.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "custom-ns",
					bindingName, gomock.Any())
				fakes.apps.EXPECT().Get(gomock.Any(), "custom-ns", "APP_NAME").Return(sampleApp, nil)
			},
			ExpectedStrings: []string{"Success", "kf restart"},
		},
		"empty namespace": {
			Args:        []string{"APP_NAME", "SERVICE_INSTANCE", `-c={"ram_gb":4}`, "--binding-name=BINDING_NAME"},
			ExpectedErr: errors.New(config.EmptySpaceError),
		},
		"defaults config": {
			Args:  []string{"APP_NAME", "SERVICE_INSTANCE"},
			Space: "custom-ns",
			Setup: func(t *testing.T, fakes fakes) {
				bindingName := v1alpha1.MakeServiceBindingName("APP_NAME", "SERVICE_INSTANCE")
				secretName := v1alpha1.MakeServiceBindingParamsSecretName("APP_NAME", "SERVICE_INSTANCE")
				fakes.servicebindings.EXPECT().Create(gomock.Any(), "custom-ns", &v1alpha1.ServiceInstanceBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:            bindingName,
						Namespace:       "custom-ns",
						OwnerReferences: ownerRefs,
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
						ParametersFrom: v1.LocalObjectReference{
							Name: secretName,
						},
						ProgressDeadlineSeconds: v1alpha1.DefaultServiceInstanceBindingProgressDeadlineSeconds,
					},
				})
				fakes.secrets.EXPECT().CreateParamsSecret(gomock.Any(), gomock.Any(), secretName, json.RawMessage("{}"))
				fakes.servicebindings.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "custom-ns",
					bindingName, gomock.Any())
				fakes.apps.EXPECT().Get(gomock.Any(), "custom-ns", "APP_NAME").Return(sampleApp, nil)
			},
		},
		"bad config path": {
			Args:  []string{"APP_NAME", "SERVICE_INSTANCE", `-c=/some/bad/path`},
			Space: "custom-ns",
			Setup: func(t *testing.T, fakes fakes) {
				fakes.apps.EXPECT().Get(gomock.Any(), "custom-ns", "APP_NAME").Return(sampleApp, nil)
			},
			ExpectedErr: errors.New("couldn't read file: open /some/bad/path: no such file or directory"),
		},
		"bad server call": {
			Args:  []string{"APP_NAME", "SERVICE_INSTANCE"},
			Space: "custom-ns",
			Setup: func(t *testing.T, fakes fakes) {
				fakes.servicebindings.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("api-error"))
				fakes.apps.EXPECT().Get(gomock.Any(), "custom-ns", "APP_NAME").Return(sampleApp, nil)
			},
			ExpectedErr: errors.New("api-error"),
		},
		"app doesn't exist": {
			Args:  []string{"APP_NAME", "SERVICE_INSTANCE"},
			Space: "custom-ns",
			Setup: func(t *testing.T, fakes fakes) {
				fakes.apps.EXPECT().Get(gomock.Any(), "custom-ns", "APP_NAME").Return(nil, apierrs.NewNotFound(v1alpha1.Resource("apps"), "APP_NAME"))
			},
			ExpectedErr: errors.New("failed to get App for binding: apps.kf.dev \"APP_NAME\" not found"),
		},
		"async": {
			Args:  []string{"--async", "APP_NAME", "SERVICE_INSTANCE"},
			Space: "default",
			Setup: func(t *testing.T, fakes fakes) {
				fakes.secrets.EXPECT().CreateParamsSecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
				fakes.servicebindings.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any())
				fakes.apps.EXPECT().Get(gomock.Any(), "default", "APP_NAME").Return(sampleApp, nil)
			},
		},
		"failed binding": {
			Args:  []string{"APP_NAME", "SERVICE_INSTANCE"},
			Space: "custom-ns",
			Setup: func(t *testing.T, fakes fakes) {
				fakes.secrets.EXPECT().CreateParamsSecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
				fakes.servicebindings.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any())
				fakes.servicebindings.EXPECT().WaitForConditionReadyTrue(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("binding already exists"))
				fakes.apps.EXPECT().Get(gomock.Any(), "custom-ns", "APP_NAME").Return(sampleApp, nil)
			},
			ExpectedErr: errors.New("bind failed: binding already exists"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			runBindingTest(t, tc)
		})
	}
}
