package services_test

import (
	"errors"
	"testing"

	servicescmd "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/services"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/services"
	"github.com/golang/mock/gomock"
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/services/fake"
	servicecatalog "github.com/poy/service-catalog/pkg/svcat/service-catalog"
)

func TestNewMarketplaceCommand(t *testing.T) {
	cases := map[string]serviceTest{
		"too many params": {
			Args:        []string{"mydb"},
			ExpectedErr: errors.New("accepts 0 arg(s), received 1"),
		},
		"command params get passed correctly": {
			Args:      []string{},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Marketplace(gomock.Any()).Do(func(opts ...services.MarketplaceOption) {
					testutil.AssertEqual(t, "namespace", "custom-ns", services.MarketplaceOptions(opts).Namespace())
				}).Return(&services.KfMarketplace{}, nil)
			},
		},
		"command output outputs instance info": {
			Args: []string{},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				fakeService := &v1beta1.ClusterServiceClass{}
				fakeService.Name = "00000000-0000-0000-0000-000000000000"
				fakeService.Spec.ExternalName = "fake-service"
				fakeService.Spec.Description = "fake-description"
				fakeService.Spec.ClusterServiceBrokerName = "fake-broker"

				f.EXPECT().Marketplace(gomock.Any()).Return(&services.KfMarketplace{
					Services: []servicecatalog.Class{fakeService},
					Plans:    []servicecatalog.Plan{},
				}, nil)
			},
			ExpectedStrings: []string{"fake-service", "fake-description", "fake-broker"},
		},
		"command output outputs plan info": {
			Args: []string{"--service=fake-service"},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				fakeService := &v1beta1.ClusterServiceClass{}
				fakeService.Name = "00000000-0000-0000-0000-000000000000"
				fakeService.Spec.ExternalName = "fake-service"

				fakePlan := &v1beta1.ClusterServicePlan{}
				fakePlan.Name = "fake-plan"
				fakePlan.Spec.ExternalName = "fake-plan"
				fakePlan.Spec.Description = "description"
				fakePlan.Spec.ClusterServiceClassRef.Name = fakeService.Name
				fakePlan.Spec.CommonServicePlanSpec.ExternalName = fakePlan.Name

				f.EXPECT().Marketplace(gomock.Any()).Return(&services.KfMarketplace{
					Services: []servicecatalog.Class{fakeService},
					Plans:    []servicecatalog.Plan{fakePlan},
				}, nil)
			},
			ExpectedStrings: []string{"fake-plan", "description"},
		},
		"blank marketplace": {
			Args: []string{},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Marketplace(gomock.Any()).Return(&services.KfMarketplace{}, nil)
			},
			ExpectedStrings: []string{},
		},
		"bad server call": {
			Args: []string{},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Marketplace(gomock.Any()).Return(nil, errors.New("server-call-error"))
			},
			ExpectedErr: errors.New("server-call-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			runTest(t, tc, servicescmd.NewMarketplaceCommand)
		})
	}
}
