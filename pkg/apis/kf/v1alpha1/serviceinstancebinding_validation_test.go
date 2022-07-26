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
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

func validServiceInstanceBinding() *ServiceInstanceBinding {
	return &ServiceInstanceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "my-ns",
		},
		Spec: *validAppServiceInstanceBindingSpec(),
	}
}

func validAppBindingType() *AppRef {
	return &AppRef{
		Name: "my-app",
	}
}

func validRouteBindingType() *RouteRef {
	return &RouteRef{
		Domain: "test.com",
	}
}

func TestServiceInstanceBinding_Validate(t *testing.T) {
	cases := testutil.ApisValidatableTestSuite{
		"ignores status updates": {
			Context: apis.WithinSubResourceUpdate(context.Background(), nil, "status"),
			Input:   &ServiceInstanceBinding{},
			Want:    nil,
		},
		"empty name": {
			Context: context.Background(),
			Input: (func() *ServiceInstanceBinding {
				binding := validServiceInstanceBinding()
				binding.Name = ""
				return binding
			}()),
			Want: &apis.FieldError{
				Message: "name or generateName is required",
				Paths:   []string{"metadata.name"},
			},
		},
		"update okay if no spec changes": {
			Context: apis.WithinUpdate(context.Background(), &ServiceInstanceBinding{}),
			Input: &ServiceInstanceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			Want: nil,
		},
		"update rejected if spec changes": {
			Context: apis.WithinUpdate(context.Background(), &ServiceInstanceBinding{}),
			Input: &ServiceInstanceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: ServiceInstanceBindingSpec{
					ParametersFrom: corev1.LocalObjectReference{
						Name: "foo",
					},
				},
			},
			Want: &apis.FieldError{
				Message: "Immutable fields changed (-old +new)",
				Paths:   []string{"spec"},
				Details: `{v1alpha1.ServiceInstanceBindingSpec}.ParametersFrom.Name:
	-: ""
	+: "foo"
`,
			},
		},
		"update ok if UnbindRequests in spec changes": {
			Context: apis.WithinUpdate(context.Background(), &ServiceInstanceBinding{}),
			Input: &ServiceInstanceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: ServiceInstanceBindingSpec{
					UnbindRequests: 1,
				},
			},
			Want: nil,
		},
	}

	cases.Run(t)
}

func TestServiceInstanceBindingSpec_Validate(t *testing.T) {
	cases := testutil.ApisValidatableTestSuite{
		"app binding nominal": {
			Context: context.Background(),
			Input:   validAppServiceInstanceBindingSpec(),
			Want:    nil,
		},
		"missing secret": {
			Context: context.Background(),
			Input: (func() *ServiceInstanceBindingSpec {
				spec := validAppServiceInstanceBindingSpec()
				spec.ParametersFrom.Name = ""
				return spec
			}()),
			Want: apis.ErrMissingField("parametersFrom.name"),
		},
		"missing instanceRef": {
			Context: context.Background(),
			Input: (func() *ServiceInstanceBindingSpec {
				spec := validAppServiceInstanceBindingSpec()
				spec.InstanceRef = corev1.LocalObjectReference{}
				return spec
			}()),
			Want: apis.ErrMissingField("instanceRef.name"),
		},
	}

	cases.Run(t)
}

func TestBindingType_Validate(t *testing.T) {
	cases := testutil.ApisValidatableTestSuite{
		"nominal": {
			Context: context.Background(),
			Input: &BindingType{
				App: validAppBindingType(),
			},
			Want: nil,
		},
		"error no bindingtype": {
			Context: context.Background(),
			Input:   &BindingType{},
			Want:    apis.ErrMissingOneOf("app", "route"),
		},
	}

	cases.Run(t)
}

func validAppServiceInstanceBindingSpec() *ServiceInstanceBindingSpec {
	return &ServiceInstanceBindingSpec{
		BindingType: BindingType{
			App: validAppBindingType(),
		},
		InstanceRef: corev1.LocalObjectReference{
			Name: "my-service",
		},
		ParametersFrom: corev1.LocalObjectReference{
			Name: "my-params-secret",
		},
		BindingNameOverride: "my-binding",
	}
}

func TestAppRef_Validate(t *testing.T) {
	cases := testutil.ApisValidatableTestSuite{
		"missing fields": {
			Context: context.Background(),
			Input:   &AppRef{},
			Want:    apis.ErrMissingField("appName"),
		},
		"populated": {
			Context: context.Background(),
			Input:   validAppBindingType(),
			Want:    nil,
		},
	}

	cases.Run(t)
}

func TestRouteRef_Validate(t *testing.T) {
	cases := testutil.ApisValidatableTestSuite{
		"missing fields": {
			Context: context.Background(),
			Input:   &RouteRef{},
			Want:    apis.ErrMissingField("routeDomain"),
		},
		"populated": {
			Context: context.Background(),
			Input:   validRouteBindingType(),
			Want:    nil,
		},
	}

	cases.Run(t)
}
