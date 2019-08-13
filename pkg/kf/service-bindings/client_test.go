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
	secretsfake "github.com/google/kf/pkg/kf/secrets/fake"
	servicebindings "github.com/google/kf/pkg/kf/service-bindings"
	"github.com/google/kf/pkg/kf/testutil"
	apiv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	testclient "github.com/poy/service-catalog/pkg/client/clientset_generated/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type fakeDependencies struct {
	apiserver *testutil.FakeApiServer
	secrets   *secretsfake.FakeClientInterface
}

type ServiceBindingApiTestCase struct {
	Run func(t *testing.T, fakes fakeDependencies, client servicebindings.ClientInterface)
}

func (tc *ServiceBindingApiTestCase) ExecuteTest(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	cs := &testclient.Clientset{}
	fakeApiServer := testutil.AddFakeReactor(cs, controller)

	secretsController := gomock.NewController(t)
	defer secretsController.Finish()
	fakeSecrets := secretsfake.NewFakeClientInterface(secretsController)

	client := servicebindings.NewClient(cs.ServicecatalogV1beta1(), fakeSecrets)
	tc.Run(t, fakeDependencies{apiserver: fakeApiServer, secrets: fakeSecrets}, client)
}

func TestClient_Create(t *testing.T) {
	cases := map[string]ServiceBindingApiTestCase{
		"server error": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.apiserver.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("api error"))
				_, err := client.Create("mydb", "myapp")
				testutil.AssertErrorsEqual(t, errors.New("api error"), err)
			},
		},
		"custom namespace": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.apiserver.EXPECT().Create(gomock.Any(), "custom-ns", gomock.Any()).Return(nil, nil)

				_, err := client.Create("mydb", "myapp", servicebindings.WithCreateNamespace("custom-ns"))
				testutil.AssertNil(t, "create err", err)
			},
		},
		"call semantics": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.apiserver.EXPECT().Create(gomock.Any(), "custom-ns", gomock.Any()).DoAndReturn(func(grv schema.GroupVersionResource, ns string, obj runtime.Object) (runtime.Object, error) {
					testutil.AssertEqual(t, "group", "servicecatalog.k8s.io", grv.Group)
					testutil.AssertEqual(t, "resource", "servicebindings", grv.Resource)

					return obj, nil
				})

				_, err := client.Create("mydb", "myapp", servicebindings.WithCreateNamespace("custom-ns"))
				testutil.AssertNil(t, "create err", err)
			},
		},
		"default values": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.apiserver.EXPECT().Create(gomock.Any(), "default", gomock.Any()).DoAndReturn(func(grv schema.GroupVersionResource, ns string, obj runtime.Object) (runtime.Object, error) {
					binding := obj.(*apiv1beta1.ServiceBinding)
					testutil.AssertEqual(t, "name", "kf-binding-myapp-mydb", binding.Name)
					testutil.AssertEqual(t, "namespace", "default", binding.Namespace)
					testutil.AssertEqual(t, "labels", map[string]string{"kf-binding-name": "mydb", "kf-app-name": "myapp"}, binding.Labels)
					testutil.AssertEqual(t, "Spec.InstanceRef.Name", "mydb", binding.Spec.InstanceRef.Name)
					testutil.AssertEqual(t, "Spec.SecretName", "kf-binding-myapp-mydb", binding.Spec.SecretName)
					return obj, nil
				})

				client.Create("mydb", "myapp")
			},
		},
		"custom values": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.apiserver.EXPECT().Create(gomock.Any(), "custom-ns", gomock.Any()).DoAndReturn(func(grv schema.GroupVersionResource, ns string, obj runtime.Object) (runtime.Object, error) {
					binding := obj.(*apiv1beta1.ServiceBinding)
					testutil.AssertEqual(t, "name", "kf-binding-myapp-mydb", binding.Name)
					testutil.AssertEqual(t, "namespace", "custom-ns", binding.Namespace)
					testutil.AssertEqual(t, "labels", map[string]string{"kf-binding-name": "binding-name", "kf-app-name": "myapp"}, binding.Labels)
					testutil.AssertEqual(t, "Spec.InstanceRef.Name", "mydb", binding.Spec.InstanceRef.Name)
					testutil.AssertEqual(t, "Spec.SecretName", "kf-binding-myapp-mydb", binding.Spec.SecretName)

					return obj, nil
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

func TestClient_GetOrCreate(t *testing.T) {
	emptyBindingList := &apiv1beta1.ServiceBindingList{
		Items: []apiv1beta1.ServiceBinding{},
	}
	cases := map[string]ServiceBindingApiTestCase{
		"first time create": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {

				deps.apiserver.EXPECT().
					List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(emptyBindingList, nil)

				deps.apiserver.EXPECT().Create(gomock.Any(), "custom-ns", gomock.Any()).DoAndReturn(func(grv schema.GroupVersionResource, ns string, obj runtime.Object) (runtime.Object, error) {
					binding := obj.(*apiv1beta1.ServiceBinding)
					testutil.AssertEqual(t, "name", "kf-binding-myapp-mydb", binding.Name)
					testutil.AssertEqual(t, "namespace", "custom-ns", binding.Namespace)
					testutil.AssertEqual(t, "labels", map[string]string{"kf-binding-name": "mydb", "kf-app-name": "myapp"}, binding.Labels)

					return obj, nil
				})

				_, created, err := client.GetOrCreate("mydb", "myapp", servicebindings.WithCreateNamespace("custom-ns"))
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "created", true, created)
			},
		},
		"already exists": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {

				deps.apiserver.EXPECT().
					List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&apiv1beta1.ServiceBindingList{
						Items: []apiv1beta1.ServiceBinding{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "kf-binding-myapp-mydb",
									Namespace: "custom-ns",
									Labels:    map[string]string{"kf-binding-name": "mydb", "kf-app-name": "myapp"},
								},
								Spec: apiv1beta1.ServiceBindingSpec{
									InstanceRef: apiv1beta1.LocalObjectReference{
										Name: "mydb",
									},
								},
							},
						},
					}, nil)

				binding, created, err := client.GetOrCreate("mydb", "myapp", servicebindings.WithCreateNamespace("custom-ns"))
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "created", false, created)
				testutil.AssertEqual(t, "name", "kf-binding-myapp-mydb", binding.Name)
				testutil.AssertEqual(t, "namespace", "custom-ns", binding.Namespace)
				testutil.AssertEqual(t, "labels", map[string]string{"kf-binding-name": "mydb", "kf-app-name": "myapp"}, binding.Labels)
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
				deps.apiserver.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("api-error"))

				err := client.Delete("mydb", "myapp")
				testutil.AssertErrorsEqual(t, errors.New("api-error"), err)
			},
		},
		"default options": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.apiserver.EXPECT().Delete(gomock.Any(), "default", "kf-binding-myapp-mydb").Return(nil)

				err := client.Delete("mydb", "myapp")
				testutil.AssertNil(t, "delete err", err)
			},
		},
		"full options": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.apiserver.EXPECT().Delete(gomock.Any(), "custom-ns", "kf-binding-myapp2-mydb2").Return(nil)

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
