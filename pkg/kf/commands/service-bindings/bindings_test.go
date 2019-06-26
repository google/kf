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
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	servicebindingscmd "github.com/google/kf/pkg/kf/commands/service-bindings"
	servicebindings "github.com/google/kf/pkg/kf/service-bindings"
	"github.com/google/kf/pkg/kf/service-bindings/fake"
	"github.com/google/kf/pkg/kf/testutil"
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
)

func TestNewListBindingsCommand(t *testing.T) {
	cases := map[string]serviceTest{
		"wrong number of args": {
			Args:        []string{"FOO"},
			ExpectedErr: errors.New("accepts 0 arg(s), received 1"),
		},
		"command params get passed correctly": {
			Args:      []string{"--app=APP_NAME", "--service=SERVICE_INSTANCE"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().List(gomock.Any()).Do(func(opts ...servicebindings.ListOption) {
					config := servicebindings.ListOptions(opts)
					testutil.AssertEqual(t, "namespace", "custom-ns", config.Namespace())
					testutil.AssertEqual(t, "app name", "custom-ns", config.Namespace())
					testutil.AssertEqual(t, "service instance name", "custom-ns", config.Namespace())
				}).Return([]v1beta1.ServiceBinding{}, nil)
			},
		},
		"defaults config": {
			Args: []string{},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().List(gomock.Any()).Do(func(opts ...servicebindings.ListOption) {
					config := servicebindings.ListOptions(opts)
					testutil.AssertEqual(t, "namespace", "", config.Namespace())
				}).Return([]v1beta1.ServiceBinding{}, nil)
			},
		},
		"bad server call": {
			Args: []string{},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().List(gomock.Any()).Return(nil, errors.New("api-error"))
			},
			ExpectedErr: errors.New("api-error"),
		},
		"output list contains items": {
			Args: []string{},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().List(gomock.Any()).Return([]v1beta1.ServiceBinding{
					*dummyBindingInstance("app1", "instance1"),
					*dummyBindingInstance("app2", "instance2"),
				}, nil)
			},
			ExpectedStrings: []string{"app1", "instance1", "app2", "instance2"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			runTest(t, tc, servicebindingscmd.NewListBindingsCommand)
		})
	}
}
