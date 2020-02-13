// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package services_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/services"
	"github.com/google/kf/pkg/kf/services/fake"
	"github.com/google/kf/pkg/kf/testutil"
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type commandFactory func(p *config.KfParams, client services.Client) *cobra.Command

func dummyServerInstance(instanceName string) *v1beta1.ServiceInstance {
	instance := v1beta1.ServiceInstance{}
	instance.Name = instanceName
	instance.Spec = v1beta1.ServiceInstanceSpec{}
	instance.Status = v1beta1.ServiceInstanceStatus{
		Conditions: []v1beta1.ServiceInstanceCondition{
			{LastTransitionTime: metav1.Time{Time: time.Now()}, Reason: "WrongStatus"},
			{LastTransitionTime: metav1.Time{Time: time.Now().Add(time.Second)}, Reason: "CorrectStatus"},
		},
	}

	return &instance
}

type serviceTest struct {
	Args      []string
	Setup     func(t *testing.T, f *fake.FakeClient)
	Namespace string

	ExpectedErr     error
	ExpectedStrings []string
}

func runTest(t *testing.T, tc serviceTest, newCommand commandFactory) {
	t.Helper()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := fake.NewFakeClient(ctrl)
	if tc.Setup != nil {
		tc.Setup(t, client)
	}

	buf := new(bytes.Buffer)
	p := &config.KfParams{
		Namespace: tc.Namespace,
	}

	cmd := newCommand(p, client)
	cmd.SetOutput(buf)
	cmd.SetArgs(tc.Args)
	_, actualErr := cmd.ExecuteC()
	if tc.ExpectedErr != nil || actualErr != nil {
		testutil.AssertErrorsEqual(t, tc.ExpectedErr, actualErr)
	}

	testutil.AssertContainsAll(t, buf.String(), tc.ExpectedStrings)
}
