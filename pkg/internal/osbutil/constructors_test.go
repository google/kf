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

package osbutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/kmeta"
	osbclient "sigs.k8s.io/go-open-service-broker-client/v2"
)

func TestNewBasicAuthSecret_valid(t *testing.T) {
	t.Parallel()

	fakeBroker := createFakeServiceBroker()

	fakeClusterBroker := &v1alpha1.ClusterServiceBroker{}
	fakeClusterBroker.Name = "name"
	fakeClusterBroker.APIVersion = "kf.dev/v1alpha1"
	fakeClusterBroker.Kind = "ClusterServiceBroker"

	cases := map[string]struct {
		name   string
		user   string
		pass   string
		url    string
		broker kmeta.OwnerRefable
	}{
		"namespaced": {
			name:   "test",
			user:   "user",
			pass:   "pass",
			url:    "https://www.google.com",
			broker: fakeBroker,
		},
		"cluster": {
			name:   "test",
			user:   "user",
			pass:   "pass",
			url:    "https://www.google.com",
			broker: fakeClusterBroker,
		},
		"blank user pass": {
			name:   "test",
			user:   "",
			pass:   "",
			url:    "https://www.google.com",
			broker: fakeClusterBroker,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			secret := NewBasicAuthSecret(tc.name, tc.user, tc.pass, tc.url, tc.broker)

			// assert the Secret is valid
			_, err := NewConfigFromSecret(secret)
			testutil.AssertNil(t, "parsing error", err)
		})
	}
}

func createFakeServiceBroker() *v1alpha1.ServiceBroker {
	fakeBroker := &v1alpha1.ServiceBroker{}
	fakeBroker.Name = "name"
	fakeBroker.Namespace = "default"
	fakeBroker.APIVersion = "kf.dev/v1alpha1"
	fakeBroker.Kind = "ServiceBroker"
	return fakeBroker
}

func createFakeBrokerSecret() *corev1.Secret {
	broker := createFakeServiceBroker()

	return NewBasicAuthSecret("test", "user", "pass", "http://www.google.com", broker)
}

func TestNewConfigFromSecret(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		secret  *corev1.Secret
		wantErr error
	}{
		"nil secret": {
			secret:  nil,
			wantErr: errNilSecret,
		},
		"good secret": {
			secret:  createFakeBrokerSecret(),
			wantErr: nil,
		},
		"wrong type": {
			secret: (func() *corev1.Secret {
				secret := createFakeBrokerSecret()
				secret.Type = "Opaque"
				return secret
			})(),
			wantErr: errSecretWrongType,
		},
		"missing field": {
			secret: (func() *corev1.Secret {
				secret := createFakeBrokerSecret()
				delete(secret.Data, credsSecretUsernameKey)
				return secret
			})(),
			wantErr: errors.New(`expected Secret to have field "username"`),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, gotErr := NewConfigFromSecret(tc.secret)
			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
		})
	}
}

func ExampleNewClient_error() {
	_, err := NewClient(nil)
	fmt.Println(err.Error())

	// Output: nil Secret not allowed
}

func ExampleNewClient() {
	// Assume the broker and secret already exist
	secret := createFakeBrokerSecret()

	// NewClient checks the structural validity of the secret, but doesn't attempt
	// to ping the broker on startup.
	client, err := NewClient(secret)
	if err != nil {
		// XXX: Handle the error correctly for your context.
		panic(err)
	}

	// XXX: Make a call to the broker here e.g. client.GetCatalog()

	fmt.Printf("Client initialized? %t\n", client != nil)

	// Output: Client initialized? true
}

func TestMapOSBToKfCatalog(t *testing.T) {
	t.Parallel()

	var minibrokerCatalog osbclient.CatalogResponse
	contents, err := ioutil.ReadFile(filepath.Join("testdata", "minibroker-catalog.json"))
	testutil.AssertNil(t, "reading catalog", err)

	err = json.Unmarshal(contents, &minibrokerCatalog)
	testutil.AssertNil(t, "unmarshal catalog", err)

	cases := map[string]struct {
		response *osbclient.CatalogResponse
	}{
		"nil": {
			response: nil,
		},
		"minibroker": {
			response: &minibrokerCatalog,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			mapped := MapOSBToKfCatalog(tc.response)
			testutil.AssertGoldenJSON(t, "kfcatalog", mapped)
		})
	}
}
