package services_test

import (
	"bytes"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/services"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/services/fake"
	"github.com/golang/mock/gomock"
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/spf13/cobra"
)

type commandFactory func(p *config.KfParams, client services.ClientInterface) *cobra.Command

func dummyServerInstance(instanceName string) *v1beta1.ServiceInstance {
	instance := v1beta1.ServiceInstance{}
	instance.Name = instanceName
	instance.Spec = v1beta1.ServiceInstanceSpec{}

	return &instance
}

type serviceTest struct {
	Args      []string
	Setup     func(t *testing.T, f *fake.FakeClientInterface)
	Namespace string

	ExpectedErr     error
	ExpectedStrings []string
}

func runTest(t *testing.T, tc serviceTest, newCommand commandFactory) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := fake.NewFakeClientInterface(ctrl)
	if tc.Setup != nil {
		tc.Setup(t, client)
	}

	buf := new(bytes.Buffer)
	p := &config.KfParams{
		Output:    buf,
		Namespace: tc.Namespace,
	}

	cmd := newCommand(p, client)
	cmd.SetOutput(buf)
	cmd.SetArgs(tc.Args)
	_, actualErr := cmd.ExecuteC()
	if tc.ExpectedErr != nil || actualErr != nil {
		testutil.AssertErrorsEqual(t, tc.ExpectedErr, actualErr)
		return
	}

	testutil.AssertContainsAll(t, buf.String(), tc.ExpectedStrings)
}
