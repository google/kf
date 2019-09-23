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

package marketplace

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalog "github.com/poy/service-catalog/pkg/svcat/service-catalog"
	servicecatalogfakes "github.com/poy/service-catalog/pkg/svcat/service-catalog/service-catalogfakes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClient_Marketplace(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		InstanceName  string
		Namespace     string
		GetClassesErr error
		GetPlansErr   error

		ExpectErr error
	}{
		"default values": {
			InstanceName: "instance-name",
			ExpectErr:    nil,
		},
		"custom values": {
			InstanceName: "instance-name",
			Namespace:    "custom-namespace",
			ExpectErr:    nil,
		},
		"error in get classes": {
			InstanceName:  "instance-name",
			GetClassesErr: errors.New("server-err"),
			ExpectErr:     errors.New("server-err"),
		},
		"error in get plans": {
			InstanceName: "instance-name",
			GetPlansErr:  errors.New("server-err"),
			ExpectErr:    errors.New("server-err"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			fakeClient := &servicecatalogfakes.FakeSvcatClient{}

			fakeClient.RetrieveClassesStub = func(scope servicecatalog.ScopeOptions) ([]servicecatalog.Class, error) {
				testutil.AssertEqual(t, "namespace", tc.Namespace, scope.Namespace)

				return nil, tc.GetClassesErr
			}

			fakeClient.RetrievePlansStub = func(classFilter string, scope servicecatalog.ScopeOptions) ([]servicecatalog.Plan, error) {
				testutil.AssertEqual(t, "namespace", tc.Namespace, scope.Namespace)
				testutil.AssertEqual(t, "classFilter", "", classFilter)

				return nil, tc.GetPlansErr
			}

			client := NewClient(func(ns string) servicecatalog.SvcatClient {
				testutil.AssertEqual(t, "namespace", tc.Namespace, ns)

				return fakeClient
			})

			_, actualErr := client.Marketplace(tc.Namespace)
			if tc.ExpectErr != nil || actualErr != nil {
				if fmt.Sprint(tc.ExpectErr) != fmt.Sprint(actualErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.ExpectErr, actualErr)
				}

				return
			}

			testutil.AssertEqual(t, "calls to RetrieveClasses", 1, fakeClient.RetrieveClassesCallCount())
			testutil.AssertEqual(t, "calls to RetrievePlans", 1, fakeClient.RetrievePlansCallCount())
		})
	}
}

func TestClient_BrokerName(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		ExpectedName                   string
		Namespace                      string
		ExpectErr                      error
		ExpectedRetrieveClassByNameErr error
	}{
		"returns broker name": {
			ExpectedName: "some-broker-name",
		},
		"fetching class fails": {
			ExpectedRetrieveClassByNameErr: errors.New("some-error"),
			ExpectErr:                      errors.New("some-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			fakeClient := &servicecatalogfakes.FakeSvcatClient{}
			fakeClass := &fakeClass{brokerName: tc.ExpectedName}
			expectedSvc := v1beta1.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tc.Namespace,
				},
			}

			fakeClient.RetrieveClassByNameStub = func(name string, opts servicecatalog.ScopeOptions) (servicecatalog.Class, error) {
				return fakeClass, tc.ExpectedRetrieveClassByNameErr
			}

			client := NewClient(func(ns string) servicecatalog.SvcatClient {
				testutil.AssertEqual(t, "namespace", tc.Namespace, ns)
				return fakeClient
			})

			name, actualErr := client.BrokerName(expectedSvc)
			if tc.ExpectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectErr, actualErr)
				return
			}

			testutil.AssertEqual(t, "broker name", tc.ExpectedName, name)
		})
	}
}

// fakeClass implements servicecatalog.Class. There isn't a fake provided.
type fakeClass struct {
	servicecatalog.Class
	brokerName string
}

func (f *fakeClass) GetServiceBrokerName() string {
	return f.brokerName
}
