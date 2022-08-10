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

package resources

import (
	"context"
	"fmt"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/cfutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

// KfInjectedEnvSecretName gets the name of the secret for the given application.
func KfInjectedEnvSecretName(app *v1alpha1.App) string {
	return fmt.Sprintf("kf-injected-envs-%s", app.Name)
}

// MakeKfInjectedEnvSecret creates a Secret containing the env vars for the given application.
func MakeKfInjectedEnvSecret(ctx context.Context, app *v1alpha1.App, space *v1alpha1.Space, serviceBindings []v1alpha1.ServiceInstanceBinding, systemEnvInjector cfutil.SystemEnvInjector) (*v1.Secret, error) {
	computedEnv, err := systemEnvInjector.ComputeSystemEnv(ctx, app, serviceBindings)
	if err != nil {
		return nil, err
	}

	secret := &v1.Secret{
		Type: v1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:      KfInjectedEnvSecretName(app),
			Namespace: space.Name,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(app),
			},
			Labels: v1alpha1.UnionMaps(app.GetLabels(), app.ComponentLabels("secret")),
		},
		Data: make(map[string][]byte),
	}

	for _, envVar := range computedEnv {
		// k8s client handles base64 encoding and decoding the secret.Data value
		secret.Data[envVar.Name] = []byte(envVar.Value)
	}

	return secret, nil
}
