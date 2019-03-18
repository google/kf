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

func TestNewDeleteServiceCommand(t *testing.T) {
	cases := map[string]serviceTest{
		"too few params": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"command params get passed correctly": {
			Args:      []string{"mydb"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().DeleteService("mydb", gomock.Any()).Do(func(name string, opts ...services.DeleteServiceOption) {
					testutil.AssertEqual(t, "namespace", "custom-ns", services.DeleteServiceOptions(opts).Namespace())
				}).Return(nil)
			},
		},
		"bad server call": {
			Args: []string{"mydb"},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().DeleteService("mydb", gomock.Any()).Return(errors.New("server-call-error"))
			},
			ExpectedErr: errors.New("server-call-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			runTest(t, tc, servicescmd.NewDeleteServiceCommand)
		})
	}
}
