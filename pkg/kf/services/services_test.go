package services

import (
	"errors"
	"fmt"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalog "github.com/poy/service-catalog/pkg/svcat/service-catalog"
	servicecatalogfakes "github.com/poy/service-catalog/pkg/svcat/service-catalog/service-catalogfakes"
)

func TestClient_CreateService(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		InstanceName string
		ServiceName  string
		PlanName     string
		Options      []CreateServiceOption
		FactoryErr   error
		ProvisionErr error

		ExpectErr error
	}{
		"default values": {
			InstanceName: "instance-name",
			ServiceName:  "service-name",
			PlanName:     "plan-name",
			Options:      []CreateServiceOption{},
			ExpectErr:    nil,
		},
		"custom values": {
			InstanceName: "instance-name",
			ServiceName:  "service-name",
			PlanName:     "plan-name",
			Options: []CreateServiceOption{
				WithCreateServiceNamespace("custom-namespace"),
				WithCreateServiceParams(map[string]interface{}{"foo": 33}),
			},
			ExpectErr: nil,
		},
		"error in factory": {
			InstanceName: "instance-name",
			ServiceName:  "service-name",
			PlanName:     "plan-name",
			FactoryErr:   errors.New("some-err"),
			ExpectErr:    errors.New("some-err"),
		},
		"error in provision": {
			InstanceName: "instance-name",
			ServiceName:  "service-name",
			PlanName:     "plan-name",
			ProvisionErr: errors.New("provision-err"),
			ExpectErr:    errors.New("provision-err"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			fakeClient := &servicecatalogfakes.FakeSvcatClient{}

			fakeClient.ProvisionStub = func(instanceName, className, planName string, opts *servicecatalog.ProvisionOptions) (*v1beta1.ServiceInstance, error) {
				expectedCfg := CreateServiceOptionDefaults().Extend(tc.Options).toConfig()

				testutil.AssertEqual(t, "instanceName", tc.InstanceName, instanceName)
				testutil.AssertEqual(t, "className", tc.ServiceName, className)
				testutil.AssertEqual(t, "planName", tc.PlanName, planName)
				testutil.AssertEqual(t, "opts.namespace", expectedCfg.Namespace, opts.Namespace)
				testutil.AssertEqual(t, "opts.params", expectedCfg.Params, opts.Params)

				return nil, tc.ProvisionErr
			}

			client := NewClient(func(ns string) (servicecatalog.SvcatClient, error) {
				return fakeClient, tc.FactoryErr
			})

			_, actualErr := client.CreateService(tc.InstanceName, tc.ServiceName, tc.PlanName, tc.Options...)
			if tc.ExpectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectErr, actualErr)

				return
			}

			testutil.AssertEqual(t, "calls to provision", 1, fakeClient.ProvisionCallCount())
		})
	}
}

func TestClient_DeleteService(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		InstanceName string
		Options      []DeleteServiceOption
		FactoryErr   error
		ServerErr    error

		ExpectErr error
	}{
		"default values": {
			InstanceName: "instance-name",
			Options:      []DeleteServiceOption{},
			ExpectErr:    nil,
		},
		"custom values": {
			InstanceName: "instance-name",
			Options: []DeleteServiceOption{
				WithDeleteServiceNamespace("custom-namespace"),
			},
			ExpectErr: nil,
		},
		"error in factory": {
			InstanceName: "instance-name",
			FactoryErr:   errors.New("some-err"),
			ExpectErr:    errors.New("some-err"),
		},
		"error in delete": {
			InstanceName: "instance-name",
			ServerErr:    errors.New("delete-err"),
			ExpectErr:    errors.New("delete-err"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			expectedCfg := DeleteServiceOptionDefaults().Extend(tc.Options).toConfig()
			fakeClient := &servicecatalogfakes.FakeSvcatClient{}

			fakeClient.DeprovisionStub = func(namespace, instanceName string) error {
				testutil.AssertEqual(t, "instanceName", tc.InstanceName, instanceName)
				testutil.AssertEqual(t, "namespace", expectedCfg.Namespace, namespace)

				return tc.ServerErr
			}

			client := NewClient(func(ns string) (servicecatalog.SvcatClient, error) {
				testutil.AssertEqual(t, "namespace", expectedCfg.Namespace, ns)

				return fakeClient, tc.FactoryErr
			})

			actualErr := client.DeleteService(tc.InstanceName, tc.Options...)
			if tc.ExpectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectErr, actualErr)

				return
			}

			testutil.AssertEqual(t, "calls to deprovision", 1, fakeClient.DeprovisionCallCount())
		})
	}
}

func TestClient_GetService(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		InstanceName string
		Options      []GetServiceOption
		FactoryErr   error
		ServerErr    error

		ExpectErr error
	}{
		"default values": {
			InstanceName: "instance-name",
			Options:      []GetServiceOption{},
			ExpectErr:    nil,
		},
		"custom values": {
			InstanceName: "instance-name",
			Options: []GetServiceOption{
				WithGetServiceNamespace("custom-namespace"),
			},
			ExpectErr: nil,
		},
		"error in factory": {
			InstanceName: "instance-name",
			FactoryErr:   errors.New("some-err"),
			ExpectErr:    errors.New("some-err"),
		},
		"error in get": {
			InstanceName: "instance-name",
			ServerErr:    errors.New("delete-err"),
			ExpectErr:    errors.New("delete-err"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			expectedCfg := GetServiceOptionDefaults().Extend(tc.Options).toConfig()
			fakeClient := &servicecatalogfakes.FakeSvcatClient{}

			fakeClient.RetrieveInstanceStub = func(namespace, instanceName string) (*v1beta1.ServiceInstance, error) {
				testutil.AssertEqual(t, "instanceName", tc.InstanceName, instanceName)
				testutil.AssertEqual(t, "namespace", expectedCfg.Namespace, namespace)

				return nil, tc.ServerErr
			}

			client := NewClient(func(ns string) (servicecatalog.SvcatClient, error) {
				testutil.AssertEqual(t, "namespace", expectedCfg.Namespace, ns)

				return fakeClient, tc.FactoryErr
			})

			_, actualErr := client.GetService(tc.InstanceName, tc.Options...)
			if tc.ExpectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectErr, actualErr)

				return
			}

			testutil.AssertEqual(t, "calls to RetrieveInstance", 1, fakeClient.RetrieveInstanceCallCount())
		})
	}
}

func TestClient_ListServices(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		InstanceName string
		Options      []ListServicesOption
		FactoryErr   error
		ServerErr    error

		ExpectErr error
	}{
		"default values": {
			InstanceName: "instance-name",
			Options:      []ListServicesOption{},
			ExpectErr:    nil,
		},
		"custom values": {
			InstanceName: "instance-name",
			Options: []ListServicesOption{
				WithListServicesNamespace("custom-namespace"),
			},
			ExpectErr: nil,
		},
		"error in factory": {
			InstanceName: "instance-name",
			FactoryErr:   errors.New("some-err"),
			ExpectErr:    errors.New("some-err"),
		},
		"error in get": {
			InstanceName: "instance-name",
			ServerErr:    errors.New("server-err"),
			ExpectErr:    errors.New("server-err"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			expectedCfg := ListServicesOptionDefaults().Extend(tc.Options).toConfig()
			fakeClient := &servicecatalogfakes.FakeSvcatClient{}

			fakeClient.RetrieveInstancesStub = func(namespace, classFilter, planFilter string) (*v1beta1.ServiceInstanceList, error) {
				testutil.AssertEqual(t, "namespace", expectedCfg.Namespace, namespace)
				testutil.AssertEqual(t, "classFilter", "", classFilter)
				testutil.AssertEqual(t, "planFilter", "", planFilter)

				return nil, tc.ServerErr
			}

			client := NewClient(func(ns string) (servicecatalog.SvcatClient, error) {
				testutil.AssertEqual(t, "namespace", expectedCfg.Namespace, ns)

				return fakeClient, tc.FactoryErr
			})

			_, actualErr := client.ListServices(tc.Options...)
			if tc.ExpectErr != nil || actualErr != nil {
				if fmt.Sprint(tc.ExpectErr) != fmt.Sprint(actualErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.ExpectErr, actualErr)
				}

				return
			}

			testutil.AssertEqual(t, "calls to RetrieveInstances", 1, fakeClient.RetrieveInstancesCallCount())
		})
	}
}
