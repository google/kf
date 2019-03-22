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

func TestNewUnbindServiceCommand(t *testing.T) {
	cases := map[string]serviceTest{
		"wrong number of args": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 2 arg(s), received 0"),
		},
		"command params get passed correctly": {
			Args:      []string{"APP_NAME", "SERVICE_INSTANCE"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Delete("SERVICE_INSTANCE", "APP_NAME", gomock.Any()).Do(func(instance, app string, opts ...servicebindings.DeleteOption) {
					config := servicebindings.DeleteOptions(opts)
					testutil.AssertEqual(t, "namespace", "custom-ns", config.Namespace())
				}).Return(nil)
			},
		},
		"defaults config": {
			Args: []string{"APP_NAME", "SERVICE_INSTANCE"},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Delete("SERVICE_INSTANCE", "APP_NAME", gomock.Any()).Do(func(instance, app string, opts ...servicebindings.DeleteOption) {
					config := servicebindings.DeleteOptions(opts)
					testutil.AssertEqual(t, "namespace", "", config.Namespace())
				}).Return(nil)
			},
		},
		"bad server call": {
			Args: []string{"APP_NAME", "SERVICE_INSTANCE"},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("api-error"))
			},
			ExpectedErr: errors.New("api-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			runTest(t, tc, servicebindingscmd.NewUnbindServiceCommand)
		})
	}
}
