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
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	kfv1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/apps"
	appsfake "github.com/google/kf/pkg/kf/apps/fake"
	servicebindings "github.com/google/kf/pkg/kf/service-bindings"
	"github.com/google/kf/pkg/kf/testutil"
	apiv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	testclient "github.com/poy/service-catalog/pkg/client/clientset_generated/clientset/fake"
)

type fakeDependencies struct {
	apiserver  *testutil.FakeApiServer
	appsClient *appsfake.FakeClient
}

type ServiceBindingApiTestCase struct {
	Run func(t *testing.T, fakes fakeDependencies, client servicebindings.ClientInterface)
}

func (tc *ServiceBindingApiTestCase) ExecuteTest(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	cs := &testclient.Clientset{}
	fakeApiServer := testutil.AddFakeReactor(cs, controller)

	appsController := gomock.NewController(t)
	defer appsController.Finish()
	fakeApps := appsfake.NewFakeClient(appsController)

	client := servicebindings.NewClient(fakeApps, cs.ServicecatalogV1beta1())
	tc.Run(t, fakeDependencies{apiserver: fakeApiServer, appsClient: fakeApps}, client)
}

func TestClient_Create(t *testing.T) {
	cases := map[string]ServiceBindingApiTestCase{
		"server error": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.appsClient.EXPECT().Transform(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("api error"))
				_, err := client.Create("mydb", "myapp")
				testutil.AssertErrorsEqual(t, errors.New("api error"), err)
			},
		},
		"custom namespace": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.appsClient.EXPECT().Transform("custom-ns", gomock.Any(), gomock.Any()).Return(nil)
				_, err := client.Create("mydb", "myapp", servicebindings.WithCreateNamespace("custom-ns"))
				testutil.AssertNil(t, "create err", err)
			},
		},
		"default values": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.appsClient.EXPECT().Transform("default", gomock.Any(), gomock.Any()).DoAndReturn(func(ns, appName string, transformer apps.Mutator) error {
					app := &kfv1alpha1.App{}
					err := transformer(app)
					testutil.AssertNil(t, "err", err)
					testutil.AssertEqual(t, "Spec.InstanceRef.Name", "mydb", app.Spec.ServiceBindings[0].InstanceRef.Name)
					// testutil.AssertEqual(t, "name", "kf-binding-myapp-mydb", binding.Name)
					// testutil.AssertEqual(t, "namespace", "default", binding.Namespace)
					// testutil.AssertEqual(t, "labels", map[string]string{"kf-binding-name": "mydb", "kf-app-name": "myapp"}, binding.Labels)
					// testutil.AssertEqual(t, "Spec.InstanceRef.Name", "mydb", binding.Spec.InstanceRef.Name)
					// testutil.AssertEqual(t, "Spec.SecretName", "kf-binding-myapp-mydb", binding.Spec.SecretName)
					return nil
				})

				client.Create("mydb", "myapp")
			},
		},
		"custom values": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.appsClient.EXPECT().Transform("custom-ns", "myapp", gomock.Any()).DoAndReturn(func(ns, appName string, transformer apps.Mutator) error {
					app := &kfv1alpha1.App{}
					err := transformer(app)
					testutil.AssertNil(t, "err", err)
					testutil.AssertEqual(t, "Spec.InstanceRef.Name", "mydb", app.Spec.ServiceBindings[0].InstanceRef.Name)

					// testutil.AssertEqual(t, "name", "kf-binding-myapp-mydb", app.Spec.ServiceBindings[0].BindingName)
					// testutil.AssertEqual(t, "namespace", "custom-ns", app.Namespace)
					// testutil.AssertEqual(t, "labels", map[string]string{"kf-binding-name": "binding-name", "kf-app-name": "myapp"}, binding.Labels)
					// testutil.AssertEqual(t, "Spec.InstanceRef.Name", "mydb", binding.Spec.InstanceRef.Name)
					// testutil.AssertEqual(t, "Spec.SecretName", "kf-binding-myapp-mydb", binding.Spec.SecretName)

					return nil
				})

				client.Create("mydb", "myapp",
					servicebindings.WithCreateBindingName("binding-name"),
					servicebindings.WithCreateNamespace("custom-ns"),
					servicebindings.WithCreateParams(map[string]interface{}{"username": "my-user"}))
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, tc.ExecuteTest)
	}
}

func TestClient_Delete(t *testing.T) {
	cases := map[string]ServiceBindingApiTestCase{
		"api-error": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.appsClient.EXPECT().Transform(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("api-error"))

				err := client.Delete("mydb", "myapp")
				testutil.AssertErrorsEqual(t, errors.New("api-error"), err)
			},
		},
		"default options": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.appsClient.EXPECT().Transform("default", "myapp", gomock.Any()).Return(nil)

				err := client.Delete("mydb", "myapp")
				testutil.AssertNil(t, "delete err", err)
			},
		},
		"full options": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.appsClient.EXPECT().Transform("custom-ns", "myapp2", gomock.Any()).Return(nil)

				err := client.Delete("mydb2", "myapp2", servicebindings.WithDeleteNamespace("custom-ns"))
				testutil.AssertNil(t, "delete err", err)
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, tc.ExecuteTest)
	}
}

func TestClient_List(t *testing.T) {
	cases := map[string]ServiceBindingApiTestCase{
		"default options": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.apiserver.EXPECT().
					List(gomock.Any(), "default", gomock.Any(), gomock.Any()).
					Return(&apiv1beta1.ServiceBindingList{}, nil)

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
					Return(&apiv1beta1.ServiceBindingList{}, nil)

				_, err := client.List(servicebindings.WithListNamespace("custom-ns"))
				testutil.AssertNil(t, "list err", err)
			},
		},
		"instances get passed back": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.apiserver.EXPECT().
					List(gomock.Any(), "default", gomock.Any(), gomock.Any()).
					Return(&apiv1beta1.ServiceBindingList{
						Items: []apiv1beta1.ServiceBinding{
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
				mybinding := apiv1beta1.ServiceBinding{}
				mybinding.Name = "bound-to-my-app"
				mybinding.Labels = map[string]string{servicebindings.AppNameLabel: "my-app"}

				otherbinding := apiv1beta1.ServiceBinding{}
				otherbinding.Name = "bound-to-other-app"
				otherbinding.Labels = map[string]string{servicebindings.AppNameLabel: "other-app"}

				deps.apiserver.EXPECT().
					List(gomock.Any(), "default", gomock.Any(), gomock.Any()).
					Return(&apiv1beta1.ServiceBindingList{Items: []apiv1beta1.ServiceBinding{mybinding, otherbinding}}, nil)

				list, err := client.List(servicebindings.WithListAppName("my-app"))
				testutil.AssertNil(t, "list err", err)
				testutil.AssertEqual(t, "item count", 1, len(list))
				testutil.AssertEqual(t, "filtered item", mybinding, list[0])
			},
		},
		"instances get filtered by service": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				mybinding := apiv1beta1.ServiceBinding{}
				mybinding.Name = "bound-to-my-service"
				mybinding.Spec.InstanceRef.Name = "my-service"

				otherbinding := apiv1beta1.ServiceBinding{}
				otherbinding.Name = "bound-to-other-service"
				otherbinding.Spec.InstanceRef.Name = "other-service"

				deps.apiserver.EXPECT().
					List(gomock.Any(), "default", gomock.Any(), gomock.Any()).
					Return(&apiv1beta1.ServiceBindingList{Items: []apiv1beta1.ServiceBinding{mybinding, otherbinding}}, nil)

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

func ExampleKfApp_BindService() {
	myApp := &kfv1alpha1.App{}
	servicebindings.BindService(myApp, &kfv1alpha1.AppSpecServiceBinding{
		InstanceRef: apiv1beta1.LocalObjectReference{
			Name: "some-service",
		},
		BindingName: "some-binding-name",
	})
	servicebindings.BindService(myApp, &kfv1alpha1.AppSpecServiceBinding{
		InstanceRef: apiv1beta1.LocalObjectReference{
			Name: "another-service",
		},
		BindingName: "some-binding-name",
	})
	servicebindings.BindService(myApp, &kfv1alpha1.AppSpecServiceBinding{
		InstanceRef: apiv1beta1.LocalObjectReference{
			Name: "third-service",
		},
		BindingName: "third",
	})
	servicebindings.BindService(myApp, &kfv1alpha1.AppSpecServiceBinding{
		InstanceRef: apiv1beta1.LocalObjectReference{
			Name: "forth-service",
		},
		BindingName: "forth",
	})
	servicebindings.UnbindService(myApp, "third")

	for _, b := range myApp.Spec.ServiceBindings {
		fmt.Println("Instance", b.InstanceRef.Name, "BindingName", b.BindingName)
	}

	// Output: Instance another-service BindingName some-binding-name
	// Instance forth-service BindingName forth
}
