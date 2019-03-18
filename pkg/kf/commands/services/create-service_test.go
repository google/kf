package services_test

import (
	"errors"
	"testing"

	servicescmd "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/services"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/services"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/services/fake"
	"github.com/golang/mock/gomock"
)

func TestNewCreateServiceCommand(t *testing.T) {

	cases := map[string]serviceTest{
		"too few params": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 3 arg(s), received 0"),
		},
		"command params get passed correctly": {
			Args:      []string{"db-service", "free", "mydb", `--config={"ram_gb":4}`},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().CreateService("mydb", "db-service", "free", gomock.Any()).Do(func(instance, service, plan string, opts ...services.CreateServiceOption) {
					config := services.CreateServiceOptions(opts)
					testutil.AssertEqual(t, "params", map[string]interface{}{"ram_gb": 4.0}, config.Params())
					testutil.AssertEqual(t, "namespace", "custom-ns", config.Namespace())
				}).Return(dummyServerInstance("mydb"), nil)
			},
		},
		"defaults config": {
			Args: []string{"db-service", "free", "mydb"},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().CreateService("mydb", "db-service", "free", gomock.Any()).Do(func(instance, service, plan string, opts ...services.CreateServiceOption) {
					config := services.CreateServiceOptions(opts)
					testutil.AssertEqual(t, "params", map[string]interface{}{}, config.Params())
					testutil.AssertEqual(t, "namespace", "", config.Namespace())
				}).Return(dummyServerInstance("mydb"), nil)
			},
		},
		"bad path": {
			Args:        []string{"db-service", "free", "mydb", `--config=/some/bad/path`},
			ExpectedErr: errors.New("couldn't read file: open /some/bad/path: no such file or directory"),
		},
		"bad server call": {
			Args: []string{"db-service", "free", "mydb"},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().CreateService(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("server-call-error"))
			},
			ExpectedErr: errors.New("server-call-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			runTest(t, tc, servicescmd.NewCreateServiceCommand)
		})
	}
}
