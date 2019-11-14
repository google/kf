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
	kfv1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	testclient "github.com/google/kf/pkg/client/servicecatalog/clientset/versioned/fake"
	servicebindings "github.com/google/kf/pkg/kf/service-bindings"
	"github.com/google/kf/pkg/kf/testutil"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
)

type fakeDependencies struct {
	apiserver *testutil.FakeApiServer
}

type ServiceBindingAPITestCase struct {
	Run func(t *testing.T, fakes fakeDependencies, client servicebindings.ClientInterface)
}

func (tc *ServiceBindingAPITestCase) ExecuteTest(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	cs := &testclient.Clientset{}
	fakeAPIServer := testutil.AddFakeReactor(cs, controller)

	appsController := gomock.NewController(t)
	defer appsController.Finish()

	client := servicebindings.NewClient(cs)
	tc.Run(t, fakeDependencies{apiserver: fakeAPIServer}, client)
}

func TestClient_List(t *testing.T) {
	cases := map[string]ServiceBindingAPITestCase{
		"default options": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.apiserver.EXPECT().
					List(gomock.Any(), "default", gomock.Any(), gomock.Any()).
					Return(&servicecatalogv1beta1.ServiceBindingList{}, nil)

				_, err := client.List()
				testutil.AssertNil(t, "list err", err)
			},
		},
		"service error": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.apiserver.EXPECT().
					List(gomock.Any(), "default", gomock.Any(), gomock.Any()).
					Return(nil, errors.New("api-error"))

				_, err := client.List()
				testutil.AssertErrorsEqual(t, errors.New("api-error"), err)
			},
		},
		"different namespace": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.apiserver.EXPECT().
					List(gomock.Any(), "custom-ns", gomock.Any(), gomock.Any()).
					Return(&servicecatalogv1beta1.ServiceBindingList{}, nil)

				_, err := client.List(servicebindings.WithListNamespace("custom-ns"))
				testutil.AssertNil(t, "list err", err)
			},
		},
		"instances get passed back": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.apiserver.EXPECT().
					List(gomock.Any(), "default", gomock.Any(), gomock.Any()).
					Return(&servicecatalogv1beta1.ServiceBindingList{
						Items: []servicecatalogv1beta1.ServiceBinding{
							{},
							{},
							{},
						},
					}, nil)

				list, err := client.List()
				testutil.AssertNil(t, "list err", err)
				testutil.AssertEqual(t, "item count", 3, len(list))
			},
		},
		"instances get filtered by app": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				mybinding := servicecatalogv1beta1.ServiceBinding{}
				mybinding.Name = "bound-to-my-app"
				mybinding.Labels = map[string]string{kfv1alpha1.NameLabel: "my-app"}

				otherbinding := servicecatalogv1beta1.ServiceBinding{}
				otherbinding.Name = "bound-to-other-app"
				otherbinding.Labels = map[string]string{kfv1alpha1.NameLabel: "other-app"}

				deps.apiserver.EXPECT().
					List(gomock.Any(), "default", gomock.Any(), gomock.Any()).
					Return(&servicecatalogv1beta1.ServiceBindingList{Items: []servicecatalogv1beta1.ServiceBinding{mybinding, otherbinding}}, nil)

				list, err := client.List(servicebindings.WithListAppName("my-app"))
				testutil.AssertNil(t, "list err", err)
				testutil.AssertEqual(t, "item count", 1, len(list))
				testutil.AssertEqual(t, "filtered item", mybinding, list[0])
			},
		},
		"instances get filtered by service": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				mybinding := servicecatalogv1beta1.ServiceBinding{}
				mybinding.Name = "bound-to-my-service"
				mybinding.Spec.InstanceRef.Name = "my-service"

				otherbinding := servicecatalogv1beta1.ServiceBinding{}
				otherbinding.Name = "bound-to-other-service"
				otherbinding.Spec.InstanceRef.Name = "other-service"

				deps.apiserver.EXPECT().
					List(gomock.Any(), "default", gomock.Any(), gomock.Any()).
					Return(&servicecatalogv1beta1.ServiceBindingList{Items: []servicecatalogv1beta1.ServiceBinding{mybinding, otherbinding}}, nil)

				list, err := client.List(servicebindings.WithListServiceInstance("my-service"))
				testutil.AssertNil(t, "list err", err)
				testutil.AssertEqual(t, "item count", 1, len(list))
				testutil.AssertEqual(t, "filtered item", mybinding, list[0])
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, tc.ExecuteTest)
	}
}
