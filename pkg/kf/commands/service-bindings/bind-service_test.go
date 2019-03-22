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

func TestNewBindServiceCommand(t *testing.T) {
	cases := map[string]serviceTest{
		"wrong number of args": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 2 arg(s), received 0"),
		},
		"command params get passed correctly": {
			Args:      []string{"APP_NAME", "SERVICE_INSTANCE", `--config={"ram_gb":4}`, "--binding-name=BINDING_NAME"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Create("SERVICE_INSTANCE", "APP_NAME", gomock.Any()).Do(func(instance, app string, opts ...servicebindings.CreateOption) {
					config := servicebindings.CreateOptions(opts)
					testutil.AssertEqual(t, "params", map[string]interface{}{"ram_gb": 4.0}, config.Params())
					testutil.AssertEqual(t, "namespace", "custom-ns", config.Namespace())
				}).Return(dummyBindingInstance("APP_NAME", "SERVICE_INSTANCE"), nil)
			},
		},
		"defaults config": {
			Args: []string{"APP_NAME", "SERVICE_INSTANCE"},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Create("SERVICE_INSTANCE", "APP_NAME", gomock.Any()).Do(func(instance, app string, opts ...servicebindings.CreateOption) {
					config := servicebindings.CreateOptions(opts)
					testutil.AssertEqual(t, "params", map[string]interface{}{}, config.Params())
					testutil.AssertEqual(t, "namespace", "", config.Namespace())
				}).Return(dummyBindingInstance("APP_NAME", "SERVICE_INSTANCE"), nil)
			},
		},
		"bad config path": {
			Args:        []string{"APP_NAME", "SERVICE_INSTANCE", `--config=/some/bad/path`},
			ExpectedErr: errors.New("couldn't read file: open /some/bad/path: no such file or directory"),
		},
		"bad server call": {
			Args: []string{"APP_NAME", "SERVICE_INSTANCE"},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("api-error"))
			},
			ExpectedErr: errors.New("api-error"),
		},
		"writes binding info": {
			Args: []string{"APP_NAME", "SERVICE_INSTANCE"},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(dummyBindingInstance("APP_NAME", "SERVICE_INSTANCE"), nil)
			},
			ExpectedStrings: []string{"APP_NAME", "SERVICE_INSTANCE"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			runTest(t, tc, servicebindingscmd.NewBindServiceCommand)
		})
	}
}
