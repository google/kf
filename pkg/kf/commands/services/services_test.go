package services_test

import (
	"errors"
	"testing"

	servicescmd "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/services"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/services"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/services/fake"
	"github.com/golang/mock/gomock"
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
)

func TestNewServicesCommand(t *testing.T) {
	cases := map[string]serviceTest{
		"too many params": {
			Args:        []string{"foo", "bar"},
			ExpectedErr: errors.New("accepts 0 arg(s), received 2"),
		},
		"custom namespace": {
			Namespace: "test-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().ListServices(gomock.Any()).
					DoAndReturn(func(opts ...services.ListServicesOption) (*v1beta1.ServiceInstanceList, error) {
						options := services.ListServicesOptions(opts)
						testutil.AssertEqual(t, "namespace", "test-ns", options.Namespace())

						return &v1beta1.ServiceInstanceList{}, nil
					})
			},
		},
		"empty result": {
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				emptyList := &v1beta1.ServiceInstanceList{Items: []v1beta1.ServiceInstance{}}
				f.EXPECT().ListServices(gomock.Any()).Return(emptyList, nil)
			},
			ExpectedErr: nil, // explicitly expecting no failure with zero length list
		},
		"full result": {
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				serviceList := &v1beta1.ServiceInstanceList{Items: []v1beta1.ServiceInstance{
					*dummyServerInstance("service-1"),
					*dummyServerInstance("service-2"),
				}}
				f.EXPECT().ListServices(gomock.Any()).Return(serviceList, nil)
			},
			ExpectedStrings: []string{"service-1", "service-2"},
		},
		"bad server call": {
			ExpectedErr: errors.New("server-call-error"),
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().ListServices(gomock.Any()).Return(nil, errors.New("server-call-error"))
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			runTest(t, tc, servicescmd.NewListServicesCommand)
		})
	}
}
