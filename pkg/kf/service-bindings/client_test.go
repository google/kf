package servicebindings_test

import (
	"errors"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	servicebindings "github.com/GoogleCloudPlatform/kf/pkg/kf/service-bindings"
	"github.com/golang/mock/gomock"
	apiv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	testclient "github.com/poy/service-catalog/pkg/client/clientset_generated/clientset/fake"
	clientv1beta1 "github.com/poy/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ServiceBindingApiTestCase struct {
	FactoryErr error
	Run        func(t *testing.T, apiserver *testutil.FakeApiServer, client servicebindings.ClientInterface)
}

func (tc *ServiceBindingApiTestCase) ExecuteTest(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	cs := &testclient.Clientset{}
	fakeApiServer := testutil.AddFakeReactor(cs, controller)

	client := servicebindings.NewClient(func() (clientv1beta1.ServicecatalogV1beta1Interface, error) {
		return cs.ServicecatalogV1beta1(), tc.FactoryErr
	})

	tc.Run(t, fakeApiServer, client)
}

func TestClient_Create(t *testing.T) {
	cases := map[string]ServiceBindingApiTestCase{
		"factory error": {
			FactoryErr: errors.New("some-error"),
			Run: func(t *testing.T, api *testutil.FakeApiServer, client servicebindings.ClientInterface) {
				_, err := client.Create("mydb", "myapp")
				testutil.AssertErrorsEqual(t, errors.New("some-error"), err)
			},
		},
		"server error": {
			Run: func(t *testing.T, api *testutil.FakeApiServer, client servicebindings.ClientInterface) {
				api.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("api error"))
				_, err := client.Create("mydb", "myapp")
				testutil.AssertErrorsEqual(t, errors.New("api error"), err)
			},
		},
		"custom namespace": {
			Run: func(t *testing.T, api *testutil.FakeApiServer, client servicebindings.ClientInterface) {
				api.EXPECT().Create(gomock.Any(), "custom-ns", gomock.Any()).Return(nil, nil)

				_, err := client.Create("mydb", "myapp", servicebindings.WithCreateNamespace("custom-ns"))
				testutil.AssertNil(t, "create err", err)
			},
		},
		"call semantics": {
			Run: func(t *testing.T, api *testutil.FakeApiServer, client servicebindings.ClientInterface) {
				api.EXPECT().Create(gomock.Any(), "custom-ns", gomock.Any()).DoAndReturn(func(grv schema.GroupVersionResource, ns string, obj runtime.Object) (runtime.Object, error) {
					testutil.AssertEqual(t, "group", "servicecatalog.k8s.io", grv.Group)
					testutil.AssertEqual(t, "resource", "servicebindings", grv.Resource)

					return obj, nil
				})

				_, err := client.Create("mydb", "myapp", servicebindings.WithCreateNamespace("custom-ns"))
				testutil.AssertNil(t, "create err", err)
			},
		},
		"default values": {
			Run: func(t *testing.T, api *testutil.FakeApiServer, client servicebindings.ClientInterface) {
				api.EXPECT().Create(gomock.Any(), "default", gomock.Any()).DoAndReturn(func(grv schema.GroupVersionResource, ns string, obj runtime.Object) (runtime.Object, error) {
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
			Run: func(t *testing.T, api *testutil.FakeApiServer, client servicebindings.ClientInterface) {
				api.EXPECT().Create(gomock.Any(), "custom-ns", gomock.Any()).DoAndReturn(func(grv schema.GroupVersionResource, ns string, obj runtime.Object) (runtime.Object, error) {
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

func TestClient_Delete(t *testing.T) {
	cases := map[string]ServiceBindingApiTestCase{
		"factory error": {
			FactoryErr: errors.New("some-error"),
			Run: func(t *testing.T, api *testutil.FakeApiServer, client servicebindings.ClientInterface) {
				err := client.Delete("mydb", "myapp")
				testutil.AssertErrorsEqual(t, errors.New("some-error"), err)
			},
		},
		"api-error": {
			Run: func(t *testing.T, api *testutil.FakeApiServer, client servicebindings.ClientInterface) {
				api.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("api-error"))

				err := client.Delete("mydb", "myapp")
				testutil.AssertErrorsEqual(t, errors.New("api-error"), err)
			},
		},
		"default options": {
			Run: func(t *testing.T, api *testutil.FakeApiServer, client servicebindings.ClientInterface) {
				api.EXPECT().Delete(gomock.Any(), "default", "kf-binding-myapp-mydb").Return(nil)

				err := client.Delete("mydb", "myapp")
				testutil.AssertNil(t, "delete err", err)
			},
		},
		"full options": {
			Run: func(t *testing.T, api *testutil.FakeApiServer, client servicebindings.ClientInterface) {
				api.EXPECT().Delete(gomock.Any(), "custom-ns", "kf-binding-myapp2-mydb2").Return(nil)

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
		"factory error": {
			FactoryErr: errors.New("some-error"),
			Run: func(t *testing.T, api *testutil.FakeApiServer, client servicebindings.ClientInterface) {
				_, err := client.List()
				testutil.AssertErrorsEqual(t, errors.New("some-error"), err)
			},
		},
		"default options": {
			Run: func(t *testing.T, api *testutil.FakeApiServer, client servicebindings.ClientInterface) {
				api.EXPECT().
					List(gomock.Any(), "default", gomock.Any(), gomock.Any()).
					Return(&apiv1beta1.ServiceBindingList{}, nil)

				_, err := client.List()
				testutil.AssertNil(t, "list err", err)
			},
		},
		"service error": {
			Run: func(t *testing.T, api *testutil.FakeApiServer, client servicebindings.ClientInterface) {
				api.EXPECT().
					List(gomock.Any(), "default", gomock.Any(), gomock.Any()).
					Return(nil, errors.New("api-error"))

				_, err := client.List()
				testutil.AssertErrorsEqual(t, errors.New("api-error"), err)
			},
		},
		"different namespace": {
			Run: func(t *testing.T, api *testutil.FakeApiServer, client servicebindings.ClientInterface) {
				api.EXPECT().
					List(gomock.Any(), "custom-ns", gomock.Any(), gomock.Any()).
					Return(&apiv1beta1.ServiceBindingList{}, nil)

				_, err := client.List(servicebindings.WithListNamespace("custom-ns"))
				testutil.AssertNil(t, "list err", err)
			},
		},
		"instances get passed back": {
			Run: func(t *testing.T, api *testutil.FakeApiServer, client servicebindings.ClientInterface) {
				api.EXPECT().
					List(gomock.Any(), "default", gomock.Any(), gomock.Any()).
					Return(&apiv1beta1.ServiceBindingList{
						Items: []apiv1beta1.ServiceBinding{
							apiv1beta1.ServiceBinding{},
							apiv1beta1.ServiceBinding{},
							apiv1beta1.ServiceBinding{},
						},
					}, nil)

				list, err := client.List()
				testutil.AssertNil(t, "list err", err)
				testutil.AssertEqual(t, "item count", 3, len(list))
			},
		},
		"instances get filtered by app": {
			Run: func(t *testing.T, api *testutil.FakeApiServer, client servicebindings.ClientInterface) {
				mybinding := apiv1beta1.ServiceBinding{}
				mybinding.Name = "bound-to-my-app"
				mybinding.Labels = map[string]string{servicebindings.AppNameLabel: "my-app"}

				otherbinding := apiv1beta1.ServiceBinding{}
				otherbinding.Name = "bound-to-other-app"
				otherbinding.Labels = map[string]string{servicebindings.AppNameLabel: "other-app"}

				api.EXPECT().
					List(gomock.Any(), "default", gomock.Any(), gomock.Any()).
					Return(&apiv1beta1.ServiceBindingList{Items: []apiv1beta1.ServiceBinding{mybinding, otherbinding}}, nil)

				list, err := client.List(servicebindings.WithListAppName("my-app"))
				testutil.AssertNil(t, "list err", err)
				testutil.AssertEqual(t, "item count", 1, len(list))
				testutil.AssertEqual(t, "filtered item", mybinding, list[0])
			},
		},
		"instances get filtered by service": {
			Run: func(t *testing.T, api *testutil.FakeApiServer, client servicebindings.ClientInterface) {
				mybinding := apiv1beta1.ServiceBinding{}
				mybinding.Name = "bound-to-my-service"
				mybinding.Spec.InstanceRef.Name = "my-service"

				otherbinding := apiv1beta1.ServiceBinding{}
				otherbinding.Name = "bound-to-other-service"
				otherbinding.Spec.InstanceRef.Name = "other-service"

				api.EXPECT().
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
