// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"encoding/json"
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func fakeCredentialsBinding() v1alpha1.ServiceInstanceBinding {
	return v1alpha1.ServiceInstanceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "my-ns",
		},
		Spec: v1alpha1.ServiceInstanceBindingSpec{
			BindingType: v1alpha1.BindingType{
				App: &v1alpha1.AppRef{
					Name: "my-app",
				},
			},
			InstanceRef: v1.LocalObjectReference{
				Name: "my-service",
			},
			ParametersFrom: v1.LocalObjectReference{
				Name: "my-params-secret",
			},
		},
	}
}

func TestMakeUserProvidedCredentialsSecret(t *testing.T) {
	fakeCredentials := json.RawMessage(`{"username":"fakeUser", "password":"fakePw"}`)
	fakeBindingParams := json.RawMessage(`{"username":"newUser", "anotherParam":"test"}`)
	emptyParams := json.RawMessage("{}")
	cases := map[string]struct {
		binding               v1alpha1.ServiceInstanceBinding
		serviceInstanceSecret v1.Secret
		bindingParamsSecret   v1.Secret
	}{
		"no binding params": {
			binding: fakeCredentialsBinding(),
			serviceInstanceSecret: v1.Secret{
				Data: map[string][]byte{
					"params": fakeCredentials,
				},
			},
			bindingParamsSecret: v1.Secret{
				Data: map[string][]byte{
					"params": emptyParams,
				},
			},
		},
		"merged binding params": {
			binding: fakeCredentialsBinding(),
			serviceInstanceSecret: v1.Secret{
				Data: map[string][]byte{
					"params": fakeCredentials,
				},
			},
			bindingParamsSecret: v1.Secret{
				Data: map[string][]byte{
					"params": fakeBindingParams,
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			secret, err := MergeCredentialsSecretForBinding(tc.binding, tc.serviceInstanceSecret, tc.bindingParamsSecret)
			assertSecretHasJSONValues(t, secret)
			testutil.AssertNil(t, "err", err)
			testutil.AssertGoldenJSON(t, "ups creds secret", secret)
		})
	}
}

func TestMakeCredentialsForOSBService(t *testing.T) {
	cases := map[string]struct {
		binding v1alpha1.ServiceInstanceBinding
		creds   map[string]interface{}
	}{
		"nil creds": {
			binding: fakeCredentialsBinding(),
			creds:   nil,
		},
		"blank creds": {
			binding: fakeCredentialsBinding(),
			creds:   make(map[string]interface{}),
		},
		"complex creds": {
			binding: fakeCredentialsBinding(),
			creds: map[string]interface{}{
				"i": -42,
				"f": float64(9.99),
				"s": "some-string",
				"a": []interface{}{1, 2, 3},
				"o": map[string]interface{}{
					"key": "value",
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			secret, err := MakeCredentialsForOSBService(&tc.binding, tc.creds)
			assertSecretHasJSONValues(t, secret)
			testutil.AssertNil(t, "err", err)
			testutil.AssertGoldenJSON(t, "osb creds", secret)
		})
	}
}

func assertSecretHasJSONValues(t *testing.T, secret *v1.Secret) {
	for k, v := range secret.Data {
		if !json.Valid(v) {
			t.Errorf("value for %q isn't valid JSON: %q", k, string(v))
		}
	}
}
