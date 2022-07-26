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

package v1alpha1

import (
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

func validBrokerInstance() *ServiceBroker {
	return &ServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: ServiceBrokerSpec{
			CommonServiceBrokerSpec: CommonServiceBrokerSpec{},
			Credentials: corev1.LocalObjectReference{
				Name: "my-creds-secret",
			},
		},
	}
}

func TestServiceBroker_Validate(t *testing.T) {
	cases := testutil.ApisValidatableTestSuite{
		"empty name": {
			Context: defaultContext(),
			Input: (func() *ServiceBroker {
				tmp := validBrokerInstance()
				tmp.Name = ""
				return tmp
			}()),
			Want: &apis.FieldError{
				Message: "name or generateName is required",
				Paths:   []string{"metadata.name"},
			},
		},
		"missing cred name": {
			Context: defaultContext(),
			Input: (func() *ServiceBroker {
				tmp := validBrokerInstance()
				tmp.Spec.Credentials.Name = ""
				return tmp
			}()),
			Want: apis.ErrMissingField("spec.credentials.name"),
		},
		"valid": {
			Context: defaultContext(),
			Input:   validBrokerInstance(),
			Want:    nil,
		},
	}

	cases.Run(t)
}

func validClusterBrokerInstance() *ClusterServiceBroker {
	return &ClusterServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: ClusterServiceBrokerSpec{
			CommonServiceBrokerSpec: CommonServiceBrokerSpec{},
			Credentials: NamespacedObjectReference{
				Name:      "my-creds-secret",
				Namespace: "foo",
			},
		},
	}
}

func TestClusterServiceBroker_Validate(t *testing.T) {
	cases := testutil.ApisValidatableTestSuite{
		"empty name": {
			Context: defaultContext(),
			Input: (func() *ClusterServiceBroker {
				tmp := validClusterBrokerInstance()
				tmp.Name = ""
				return tmp
			}()),
			Want: &apis.FieldError{
				Message: "name or generateName is required",
				Paths:   []string{"metadata.name"},
			},
		},
		"valid": {
			Context: defaultContext(),
			Input:   validClusterBrokerInstance(),
			Want:    nil,
		},
		"missing cred name": {
			Context: defaultContext(),
			Input: (func() *ClusterServiceBroker {
				tmp := validClusterBrokerInstance()
				tmp.Spec.Credentials.Name = ""
				return tmp
			}()),
			Want: apis.ErrMissingField("spec.credentials.name"),
		},
		"missing cred namespace": {
			Context: defaultContext(),
			Input: (func() *ClusterServiceBroker {
				tmp := validClusterBrokerInstance()
				tmp.Spec.Credentials.Namespace = ""
				return tmp
			}()),
			Want: apis.ErrMissingField("spec.credentials.namespace"),
		},
	}

	cases.Run(t)
}
