package servicebindings_test

import (
	"errors"
	"testing"

	servicebindingscmd "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/service-bindings"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	servicebindings "github.com/GoogleCloudPlatform/kf/pkg/kf/service-bindings"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/service-bindings/fake"
	"github.com/golang/mock/gomock"
)

func TestNewVcapServicesCommand(t *testing.T) {
	cases := map[string]serviceTest{
		"wrong number of args": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"command params get passed correctly": {
			Args:      []string{"APP_NAME"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().GetVcapServices("APP_NAME", gomock.Any()).Do(func(app string, opts ...servicebindings.GetVcapServicesOption) {
					config := servicebindings.GetVcapServicesOptions(opts)
					testutil.AssertEqual(t, "namespace", "custom-ns", config.Namespace())
				}).Return(nil, nil)
			},
		},
		"bad server call": {
			Args: []string{"APP_NAME"},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().GetVcapServices(gomock.Any(), gomock.Any()).Return(nil, errors.New("api-error"))
			},
			ExpectedErr: errors.New("api-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			runTest(t, tc, servicebindingscmd.NewVcapServicesCommand)
		})
	}
}
