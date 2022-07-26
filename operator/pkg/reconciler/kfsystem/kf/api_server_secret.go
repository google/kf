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

package kf

import (
	"context"

	mf "github.com/manifestival/manifestival"
	"knative.dev/pkg/logging"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"
	certresources "knative.dev/pkg/webhook/certificates/resources"
)

const (
	apiServiceSecretName = "upload-api-server-secret"
)

// APIServerSecretTransformer transforms the API service secret to have the
// given TLS configuration.
func APIServerSecretTransformer(ctx context.Context, cert, key, cacert []byte) mf.Transformer {
	log := logging.FromContext(ctx)

	return func(u *unstructured.Unstructured) error {
		if u.GetKind() != "Secret" || u.GetName() != apiServiceSecretName {
			return nil
		}

		secret := &corev1.Secret{}
		if err := scheme.Scheme.Convert(u, secret, nil); err != nil {
			log.Error(err, "Error converting Unstructured to Secret", "unstructured", u, "secret", secret)
			return err
		}

		secret.Data = map[string][]byte{
			certresources.ServerCert: cert,
			certresources.ServerKey:  key,
			certresources.CACert:     cacert,
		}

		return scheme.Scheme.Convert(secret, u, nil)
	}
}
