// Copyright 2020 Google LLC
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

package v1alpha1

import (
	"context"
	"math"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

func validBindingInstance() *ServiceInstance {
	return &ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: ServiceInstanceSpec{
			ServiceType: ServiceType{
				Brokered: validBrokeredInstance(),
			},
			Tags: []string{"mysql", "db", "maria"},
			ParametersFrom: corev1.LocalObjectReference{
				Name: "my-params-secret",
			},
		},
	}
}

func validUserProvidedInstance() *ServiceInstance {
	return &ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: *validUPSServiceInstanceSpec(),
	}
}

func TestServiceInstance_Validate(t *testing.T) {
	cases := testutil.ApisValidatableTestSuite{
		"ignores status updates": {
			Context: apis.WithinSubResourceUpdate(context.Background(), nil, "status"),
			Input:   &ServiceInstance{},
			Want:    nil,
		},
		"empty name": {
			Context: context.Background(),
			Input: (func() *ServiceInstance {
				inst := validBindingInstance()
				inst.Name = ""
				return inst
			}()),
			Want: &apis.FieldError{
				Message: "name or generateName is required",
				Paths:   []string{"metadata.name"},
			},
		},
		"update okay if no spec changes": {
			Context: apis.WithinUpdate(context.Background(), &ServiceInstance{
				Spec: ServiceInstanceSpec{
					ServiceType: ServiceType{
						Brokered: validBrokeredInstance(),
					},
				},
			}),
			Input: &ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: ServiceInstanceSpec{
					ServiceType: ServiceType{
						Brokered: validBrokeredInstance(),
					},
				},
			},
			Want: nil,
		},
		"brokered service update rejected if spec changes": {
			Context: apis.WithinUpdate(context.Background(), validBindingInstance()),
			Input: (func() *ServiceInstance {
				inst := validBindingInstance()
				inst.Spec.ParametersFrom.Name = "foo"
				return inst
			}()),
			Want: &apis.FieldError{
				Message: "Immutable fields changed (-old +new)",
				Paths:   []string{"spec"},
				Details: `{v1alpha1.ServiceInstanceSpec}.ParametersFrom.Name:
	-: "my-params-secret"
	+: "foo"
`,
			},
		},
		"update ok if DeleteRequests in spec changes": {
			Context: apis.WithinUpdate(context.Background(), &ServiceInstance{
				Spec: ServiceInstanceSpec{
					ServiceType: ServiceType{
						Brokered: validBrokeredInstance(),
					},
				},
			}),
			Input: &ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: ServiceInstanceSpec{
					ServiceType: ServiceType{
						Brokered: validBrokeredInstance(),
					},
					DeleteRequests: 1,
				},
			},
			Want: nil,
		},
	}

	cases.Run(t)
}

func validUPSServiceInstanceSpec() *ServiceInstanceSpec {
	return &ServiceInstanceSpec{
		ServiceType: ServiceType{
			UPS: validUPSInstance(),
		},
		Tags: []string{"ups"},
		ParametersFrom: corev1.LocalObjectReference{
			Name: "ups-secret",
		},
	}
}

func TestServiceInstanceSpec_Validate(t *testing.T) {
	cases := testutil.ApisValidatableTestSuite{
		"ups nominal": {
			Context: context.Background(),
			Input:   validUPSServiceInstanceSpec(),
			Want:    nil,
		},
		"type error path": {
			Context: context.Background(),
			Input: (func() *ServiceInstanceSpec {
				spec := validUPSServiceInstanceSpec()
				spec.Brokered = validBrokeredInstance()
				return spec
			}()),
			Want: apis.ErrMultipleOneOf("brokered", "userProvided"),
		},
		"missing secret": {
			Context: context.Background(),
			Input: (func() *ServiceInstanceSpec {
				spec := validUPSServiceInstanceSpec()
				spec.ParametersFrom.Name = ""
				return spec
			}()),
			Want: apis.ErrMissingField("parametersFrom.name"),
		},
	}

	cases.Run(t)
}

func TestServiceType_Validate(t *testing.T) {
	cases := testutil.ApisValidatableTestSuite{
		"multiple types selected": {
			Context: context.Background(),
			Input: &ServiceType{
				Brokered: validBrokeredInstance(),
				UPS:      validUPSInstance(),
			},
			Want: apis.ErrMultipleOneOf("brokered", "userProvided"),
		},
		"nominal": {
			Context: context.Background(),
			Input: &ServiceType{
				UPS: validUPSInstance(),
			},
			Want: nil,
		},
		"error path appends": {
			Context: context.Background(),
			Input: &ServiceType{
				Brokered: (func() *BrokeredInstance {
					out := validBrokeredInstance()
					out.PlanName = ""
					return out
				}()),
			},
			Want: apis.ErrMissingField("brokered.plan"),
		},
		"error no servicetype": {
			Context: context.Background(),
			Input:   &ServiceType{},
			Want:    apis.ErrMissingOneOf("brokered", "osb", "userProvided", "volume"),
		},
	}

	cases.Run(t)
}

func validUPSInstance() *UPSInstance {
	return &UPSInstance{}
}

func TestUPSInstance_Validate(t *testing.T) {
	cases := testutil.ApisValidatableTestSuite{
		"populated": {
			Context: context.Background(),
			Input:   validUPSInstance(),
			Want:    nil,
		},
	}

	cases.Run(t)
}

func validBrokeredInstance() *BrokeredInstance {
	return &BrokeredInstance{
		Broker:     "gcp-service-broker",
		ClassName:  "cloudsql",
		PlanName:   "small",
		Namespaced: false,
	}
}

func TestBrokeredInstance_Validate(t *testing.T) {
	cases := testutil.ApisValidatableTestSuite{
		"missing fields": {
			Context: context.Background(),
			Input:   &BrokeredInstance{},
			Want:    apis.ErrMissingField("class", "plan"),
		},
		"populated": {
			Context: context.Background(),
			Input:   validBrokeredInstance(),
			Want:    nil,
		},
	}

	cases.Run(t)
}

func validOSBInstance() *OSBInstance {
	return &OSBInstance{
		BrokerName: "gcp-service-broker",
		ClassName:  "CloudSQL",
		ClassUID:   "abc-def",
		PlanName:   "free-tier",
		PlanUID:    "ghi-jkl",
		Namespaced: false,
	}
}

func TestOSBInstance_Validate(t *testing.T) {
	cases := testutil.ApisValidatableTestSuite{
		"missing fields": {
			Context: context.Background(),
			Input:   &OSBInstance{},
			Want: apis.ErrMissingField(
				"className",
				"classUID",
				"planName",
				"planUID",
				"brokerName",
			),
		},
		"populated": {
			Context: context.Background(),
			Input:   validOSBInstance(),
			Want:    nil,
		},
		"bad progress deadline": {
			Context: context.Background(),
			Input: (func() *OSBInstance {
				osb := validOSBInstance()
				osb.ProgressDeadlineSeconds = -1
				return osb
			}()),
			Want: apis.ErrOutOfBoundsValue(-1, 1, math.MaxInt64, "progressDeadlineSeconds"),
		},
	}

	cases.Run(t)
}
