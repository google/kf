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

package config

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

const (
	SecretsConfigName       = "config-secrets"
	BuildImagePushSecretKey = "build.imagePushSecrets"
)

// SecretsConfig contains the configuration defined in the build secrets
// config map.
type SecretsConfig struct {
	// BuildImagePushSecrets are the names of the Secrets that should be used
	// in each space to push images via the build pipeline.
	BuildImagePushSecrets []corev1.ObjectReference
}

// NewSecretsConfigFromConfigMap creates a SecretConfig from the supplied
// ConfigMap
func NewSecretsConfigFromConfigMap(
	configMap *corev1.ConfigMap,
) (*SecretsConfig, error) {
	sc := &SecretsConfig{}

	for _, secret := range strings.Split(configMap.Data[BuildImagePushSecretKey], ",") {
		sc.BuildImagePushSecrets = append(sc.BuildImagePushSecrets, corev1.ObjectReference{
			Name: strings.TrimSpace(secret),
		})
	}

	return sc, nil
}
