// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	osbclient "sigs.k8s.io/go-open-service-broker-client/v2"
)

func fakeServiceInstance() *v1alpha1.ServiceInstance {
	instance := &v1alpha1.ServiceInstance{}
	instance.Name = "mydb"
	instance.Namespace = "test-ns"
	instance.UID = "00000000-0000-0000-0000-000008675309"
	instance.Spec.OSB = &v1alpha1.OSBInstance{
		ClassUID: "class-uid",
		PlanUID:  "plan-uid",
	}

	return instance
}

func TestMakeOSBDeprovisionRequest(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		serviceInstance *v1alpha1.ServiceInstance
	}{
		"good": {
			serviceInstance: fakeServiceInstance(),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			req := MakeOSBDeprovisionRequest(tc.serviceInstance)

			testutil.AssertGoldenJSONContext(t, "OSBDeprovisionRequest", req, map[string]interface{}{
				"serviceInstance": tc.serviceInstance,
			})
		})
	}
}

func TestMakeOSBLastOperationRequest(t *testing.T) {
	t.Parallel()

	exampleKey := "some-long-running-operation-key"

	cases := map[string]struct {
		serviceInstance *v1alpha1.ServiceInstance
		operationKey    *string
	}{
		"nil key": {
			serviceInstance: fakeServiceInstance(),
			operationKey:    nil,
		},
		"populated key": {
			serviceInstance: fakeServiceInstance(),
			operationKey:    &exampleKey,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			req := MakeOSBLastOperationRequest(tc.serviceInstance, tc.operationKey)

			testutil.AssertGoldenJSONContext(t, "OSBLastOperationRequest", req, map[string]interface{}{
				"serviceInstance": tc.serviceInstance,
				"operationKey":    tc.operationKey,
			})
		})
	}
}

func TestMakeOSBProvisionRequest(t *testing.T) {
	t.Parallel()

	goodNamespace := &corev1.Namespace{}
	goodNamespace.Name = "some-ns"
	goodNamespace.UID = "11111111-1111-1111-1111-111111111111"

	goodSecret := &corev1.Secret{}
	goodSecret.Data = map[string][]byte{
		v1alpha1.ServiceInstanceParamsSecretKey: []byte(`{"foo":"bar"}`),
	}

	cases := map[string]struct {
		serviceInstance *v1alpha1.ServiceInstance
		namespace       *corev1.Namespace
		paramsSecret    *corev1.Secret

		// NOTE: check for the invariant rather than specific error strings.
		wantErr bool
	}{
		"missing serviceInstance": {
			serviceInstance: nil,
			namespace:       goodNamespace,
			paramsSecret:    goodSecret,
			wantErr:         true,
		},
		"missing namespace": {
			serviceInstance: fakeServiceInstance(),
			namespace:       nil,
			paramsSecret:    goodSecret,
			wantErr:         true,
		},
		"missing secret": {
			serviceInstance: fakeServiceInstance(),
			namespace:       goodNamespace,
			paramsSecret:    nil,
			wantErr:         true,
		},
		"blank secret": {
			serviceInstance: fakeServiceInstance(),
			namespace:       goodNamespace,
			paramsSecret:    &corev1.Secret{},
			wantErr:         true,
		},
		"bad json secret": {
			serviceInstance: fakeServiceInstance(),
			namespace:       goodNamespace,
			paramsSecret: &corev1.Secret{
				Data: map[string][]byte{
					v1alpha1.ServiceInstanceParamsSecretKey: []byte(`{[]}`),
				},
			},
			wantErr: true,
		},
		"good": {
			serviceInstance: fakeServiceInstance(),
			namespace:       goodNamespace,
			paramsSecret:    goodSecret,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			req, err := MakeOSBProvisionRequest(
				tc.serviceInstance,
				tc.namespace,
				tc.paramsSecret,
			)

			if err != nil {
				// either request or err is nil
				testutil.AssertEqual(t, "request", (*osbclient.ProvisionRequest)(nil), req)
				testutil.AssertTrue(t, "wantErr", tc.wantErr)
			} else {
				testutil.AssertNotNil(t, "request", req) // either request or err is nil
				testutil.AssertGoldenJSONContext(t, "OSBProvisionRequest", req, map[string]interface{}{
					"serviceInstance": tc.serviceInstance,
					"namespace":       tc.namespace,
					"paramsSecret":    tc.paramsSecret,
				})
			}
		})
	}
}
