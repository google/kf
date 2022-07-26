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

func fakeServiceInstanceBinding() *v1alpha1.ServiceInstanceBinding {
	binding := &v1alpha1.ServiceInstanceBinding{}
	binding.Name = "mydb-binding"
	binding.Namespace = "test-ns"
	binding.UID = "22222222-2222-2222-2222-222222222222"

	return binding
}

func TestMakeOSBUnbindRequest(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		serviceInstance *v1alpha1.ServiceInstance
		binding         *v1alpha1.ServiceInstanceBinding
	}{
		"good": {
			serviceInstance: fakeServiceInstance(),
			binding:         fakeServiceInstanceBinding(),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			req := MakeOSBUnbindRequest(tc.serviceInstance, tc.binding)

			testutil.AssertGoldenJSONContext(t, "OSBUnbindRequest", req, map[string]interface{}{
				"serviceInstance": tc.serviceInstance,
				"binding":         tc.binding,
			})
		})
	}
}

func TestMakeOSBBindingLastOperationRequest(t *testing.T) {
	t.Parallel()

	exampleKey := "some-long-running-operation-key"

	cases := map[string]struct {
		serviceInstance *v1alpha1.ServiceInstance
		binding         *v1alpha1.ServiceInstanceBinding
		operationKey    *string
	}{
		"nil key": {
			serviceInstance: fakeServiceInstance(),
			binding:         fakeServiceInstanceBinding(),
			operationKey:    nil,
		},
		"populated key": {
			serviceInstance: fakeServiceInstance(),
			binding:         fakeServiceInstanceBinding(),
			operationKey:    &exampleKey,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			req := MakeOSBBindingLastOperationRequest(tc.serviceInstance, tc.binding, tc.operationKey)

			testutil.AssertGoldenJSONContext(t, "OSBBindingLastOperationRequest", req, map[string]interface{}{
				"serviceInstance": tc.serviceInstance,
				"binding":         tc.binding,
				"operationKey":    tc.operationKey,
			})
		})
	}
}

func TestMakeOSBBindRequest(t *testing.T) {
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
		binding         *v1alpha1.ServiceInstanceBinding
		namespace       *corev1.Namespace
		paramsSecret    *corev1.Secret

		// NOTE: check for the invariant rather than specific error strings.
		wantErr bool
	}{
		"missing serviceInstance": {
			serviceInstance: nil,
			binding:         fakeServiceInstanceBinding(),
			namespace:       goodNamespace,
			paramsSecret:    goodSecret,
			wantErr:         true,
		},
		"missing binding": {
			serviceInstance: fakeServiceInstance(),
			binding:         nil,
			namespace:       goodNamespace,
			paramsSecret:    goodSecret,
			wantErr:         true,
		},
		"missing namespace": {
			serviceInstance: fakeServiceInstance(),
			binding:         fakeServiceInstanceBinding(),
			namespace:       nil,
			paramsSecret:    goodSecret,
			wantErr:         true,
		},
		"missing secret": {
			serviceInstance: fakeServiceInstance(),
			binding:         fakeServiceInstanceBinding(),
			namespace:       goodNamespace,
			paramsSecret:    nil,
			wantErr:         true,
		},
		"blank secret": {
			serviceInstance: fakeServiceInstance(),
			binding:         fakeServiceInstanceBinding(),
			namespace:       goodNamespace,
			paramsSecret:    &corev1.Secret{},
			wantErr:         true,
		},
		"bad json secret": {
			serviceInstance: fakeServiceInstance(),
			binding:         fakeServiceInstanceBinding(),
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
			binding:         fakeServiceInstanceBinding(),
			namespace:       goodNamespace,
			paramsSecret:    goodSecret,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			req, err := MakeOSBBindRequest(
				tc.serviceInstance,
				tc.binding,
				tc.namespace,
				tc.paramsSecret,
			)

			if err != nil {
				// either request or err is nil
				testutil.AssertEqual(t, "request", (*osbclient.BindRequest)(nil), req)
				testutil.AssertTrue(t, "wantErr", tc.wantErr)
			} else {
				testutil.AssertNotNil(t, "request", req) // either request or err is nil
				testutil.AssertGoldenJSONContext(t, "OSBBindRequest", req, map[string]interface{}{
					"serviceInstance": tc.serviceInstance,
					"binding":         tc.binding,
					"namespace":       tc.namespace,
					"paramsSecret":    tc.paramsSecret,
				})
			}
		})
	}
}

func TestMakeOSBGetBindingRequest(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		serviceInstance *v1alpha1.ServiceInstance
		binding         *v1alpha1.ServiceInstanceBinding
	}{
		"good": {
			serviceInstance: fakeServiceInstance(),
			binding:         fakeServiceInstanceBinding(),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			req := MakeOSBGetBindingRequest(tc.serviceInstance, tc.binding)

			testutil.AssertGoldenJSONContext(t, "OSBGetBindingRequest", req, map[string]interface{}{
				"serviceInstance": tc.serviceInstance,
				"binding":         tc.binding,
			})
		})
	}
}
