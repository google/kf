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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"
	osstype "knative.dev/operator/pkg/apis/operator/v1alpha1"
	osscommon "knative.dev/operator/pkg/reconciler/common"
	"knative.dev/pkg/logging"

	"kf-operator/pkg/apis/kfsystem/kf"
	"kf-operator/pkg/apis/kfsystem/v1alpha1"
	kfconfig "kf-operator/pkg/reconciler/kfsystem/kf/config"
)

const (
	googleServiceAccountSuffix = ".iam.gserviceaccount.com"
)

// ConfigSecretsTransform transforms configurations into ConfigMap data
func ConfigSecretsTransform(ctx context.Context, secrets v1alpha1.SecretSpec) mf.Transformer {
	logger := logging.FromContext(ctx)
	configData := osstype.ConfigMapData{}
	updateSecrets(&configData, secrets)

	return osscommon.ConfigMapTransform(configData, logger)
}

// ConfigDefaultsTransform transforms default configurations into ConfigMap data
func ConfigDefaultsTransform(ctx context.Context, defaults kf.DefaultsConfig) mf.Transformer {
	return func(u *unstructured.Unstructured) error {
		if u.GetKind() == "ConfigMap" && u.GetName() == kf.DefaultsConfigName {
			log := logging.FromContext(ctx)
			cm := &corev1.ConfigMap{}
			if err := scheme.Scheme.Convert(u, cm, nil); err != nil {
				log.Error(err, "Error converting Unstructured to ConfigMap", "unstructured", u, "configmap", cm)
				return err
			}

			if err := defaults.PatchConfigMap(cm); err != nil {
				log.Error(err, "Error patch ConfigMap with DefaultsConfig", "unstructured", u, "configmap", cm)
				return err
			}

			return scheme.Scheme.Convert(cm, u, nil)
		}

		return nil
	}
}

func updateSecrets(configData *osstype.ConfigMapData, secrets v1alpha1.SecretSpec) {
	(*configData)[kfconfig.SecretsConfigName] = make(map[string]string)
	secretConfigMap := (*configData)[kfconfig.SecretsConfigName]

	if secrets.WorkloadIdentity != nil {
		secretConfigMap[kfconfig.GoogleServiceAccountKey] = getGoogleServiceAccount(secrets.WorkloadIdentity.GoogleServiceAccount, secrets.WorkloadIdentity.GoogleProjectID)
		secretConfigMap[kfconfig.GoogleProjectIDKey] = secrets.WorkloadIdentity.GoogleProjectID
	} else if secrets.Build != nil {
		secretConfigMap[kfconfig.BuildImagePushSecretKey] = secrets.Build.ImagePushSecretName
	}
}

// GetGoogleServiceAccount return the full string of Google Service Account
func getGoogleServiceAccount(serviceAccountName string, projectID string) string {
	return serviceAccountName + "@" + projectID + googleServiceAccountSuffix
}
