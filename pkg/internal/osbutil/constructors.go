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
	"errors"
	"fmt"
	"sort"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/kmeta"
	osbclient "sigs.k8s.io/go-open-service-broker-client/v2"
)

const (
	// credsSecretType is a special type for Kubernetes
	// secrets indicating that it contains a service broker credential.
	credsSecretType = "kf.dev/servicebrokercreds"

	// credsSecretUsernameKey is the key used in the secret
	// that holds the BasicAuth username used to connect to the broker.
	credsSecretUsernameKey = "username"
	// credsSecretPasswordKey is the key used in the secret
	// that holds the BasicAuth password used to connect to the broker.
	credsSecretPasswordKey = "password"
	// credsSecretURLKey is the key used in the secret
	// that holds the URL used to connect to the broker.
	credsSecretURLKey = "url"
)

var (
	errNilSecret       = errors.New("nil Secret not allowed")
	errSecretWrongType = fmt.Errorf("expected Secret to have type %q", credsSecretType)
)

// NewBasicAuthSecret creates a new secret for connecting to a service broker
// via HTTP Basic Authentication.
func NewBasicAuthSecret(name, username, password, url string, broker kmeta.OwnerRefable) *corev1.Secret {
	out := &corev1.Secret{}
	out.Name = name

	if ns := broker.GetObjectMeta().GetNamespace(); ns != "" {
		out.Namespace = ns
	} else {
		out.Namespace = v1alpha1.KfNamespace
	}
	out.OwnerReferences = append(out.OwnerReferences, *kmeta.NewControllerRef(broker))

	out.Type = credsSecretType
	out.Data = make(map[string][]byte)
	out.Data[credsSecretUsernameKey] = []byte(username)
	out.Data[credsSecretPasswordKey] = []byte(password)
	out.Data[credsSecretURLKey] = []byte(url)

	return out
}

// NewConfigFromSecret creates a new OSB connection configuration from
// a secret; in the process it validates the secret for structural
// correctness.
func NewConfigFromSecret(secret *corev1.Secret) (*osbclient.ClientConfiguration, error) {
	if secret == nil {
		return nil, errNilSecret
	}

	// validate secret is meant to be used by the client
	if secret.Type != credsSecretType {
		return nil, errSecretWrongType
	}

	config := osbclient.DefaultClientConfiguration()
	config.AuthConfig = &osbclient.AuthConfig{}
	config.AuthConfig.BasicAuthConfig = &osbclient.BasicAuthConfig{}

	// get username, password, url from secret
	for _, field := range []struct {
		dest *string
		key  string
	}{
		{&config.AuthConfig.BasicAuthConfig.Username, credsSecretUsernameKey},
		{&config.AuthConfig.BasicAuthConfig.Password, credsSecretPasswordKey},
		{&config.URL, credsSecretURLKey},
	} {
		val, ok := secret.Data[field.key]
		if !ok {
			return nil, fmt.Errorf("expected Secret to have field %q", field.key)
		}
		*field.dest = string(val)
	}

	return config, nil
}

// NewClient creates a new OSB client from a secret.
func NewClient(secret *corev1.Secret) (osbclient.Client, error) {
	config, err := NewConfigFromSecret(secret)
	if err != nil {
		return nil, err
	}

	client, err := osbclient.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// MapOSBToKfCatalog converts an OSB style catalog to Kf's version with
// deterministic ordering.
func MapOSBToKfCatalog(catalog *osbclient.CatalogResponse) (out []v1alpha1.ServiceOffering) {
	if catalog == nil {
		return
	}

	for _, osbService := range catalog.Services {
		var plans []v1alpha1.ServicePlan
		for _, osbPlan := range osbService.Plans {
			// free is a pointer, but defaults to true in OSB
			isFree := true
			if osbPlan.Free != nil {
				isFree = *osbPlan.Free
			}

			plans = append(plans, v1alpha1.ServicePlan{
				DisplayName: osbPlan.Name,
				UID:         osbPlan.ID,
				Free:        isFree,
				Description: osbPlan.Description,
			})
		}

		sort.Slice(plans, func(i, j int) bool {
			return plans[i].DisplayName < plans[j].DisplayName
		})

		out = append(out, v1alpha1.ServiceOffering{
			DisplayName: osbService.Name,
			UID:         osbService.ID,
			Description: osbService.Description,
			Tags:        osbService.Tags,
			Plans:       plans,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].DisplayName < out[j].DisplayName
	})

	return
}
