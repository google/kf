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

package config

import (
	"log"
	"strings"

	"cloud.google.com/go/compute/metadata"
	corev1 "k8s.io/api/core/v1"
)

const (
	// SecretsConfigName is the name of ConfigMap that stores secrets
	SecretsConfigName = "config-secrets"
	// BuildImagePushSecretKey is the key of BuildImagePushSecret in the Configmap
	BuildImagePushSecretKey = "build.imagePushSecrets"
	// GoogleServiceAccountKey is the key of GSA in the Configmap
	GoogleServiceAccountKey = "wi.googleServiceAccount"
	// GoogleProjectIDKey is the key of Google project Id in the Configmap
	GoogleProjectIDKey = "wi.googleProjectID"
)

// SecretsConfig contains the configuration defined in the build secrets
// config map.
type SecretsConfig struct {
	// BuildImagePushSecrets are the names of the Secrets that should be used
	// in each space to push images via the build pipeline.
	BuildImagePushSecrets []corev1.ObjectReference

	// GoogleProjectID is the GCP project ID used with Workload Identity. If
	// left empty, the metadata server will be used to fetch it.
	GoogleProjectID string

	// GoogleServiceAccount is the GSA that is linked to the KSA in each space
	// for Workload Identity. If it's empty, then WI is disabled and the Kf
	// controllers will not make calls to the IAM API.
	GoogleServiceAccount string
}

// NewSecretsConfigFromConfigMap creates a SecretConfig from the supplied
// ConfigMap
func NewSecretsConfigFromConfigMap(
	configMap *corev1.ConfigMap,
) (*SecretsConfig, error) {
	sc := &SecretsConfig{}

	sc.GoogleServiceAccount = configMap.Data[GoogleServiceAccountKey]

	sc.GoogleProjectID = configMap.Data[GoogleProjectIDKey]
	if sc.GoogleProjectID == "" {
		// Use the metadata server to fetch the project ID
		projectID, err := metadata.ProjectID()
		if err != nil {
			// Log the error instead of failing given the default value uses
			// this. It will also only fail if the user isn't on GCP.
			log.Printf("failed to fetch GCP project ID for WI: %v", err)
		}

		sc.GoogleProjectID = projectID
	}

	// Parse build.imagePushSecerts
	for _, secret := range strings.Split(configMap.Data[BuildImagePushSecretKey], ",") {
		sc.BuildImagePushSecrets = append(sc.BuildImagePushSecrets, corev1.ObjectReference{
			Name: strings.TrimSpace(secret),
		})
	}

	return sc, nil
}
