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
	"fmt"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/systemenvinjector"
	"github.com/knative/serving/pkg/resources"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

// SecretName gets the name of the secret for the given application.
func SecretName(app *v1alpha1.App) string {
	return fmt.Sprintf("kf-injected-envs-%s", app.Name)
}

// MakeSecret creates a Secret containing the env vars for the given application.
func MakeSecret(app *v1alpha1.App, space *v1alpha1.Space, systemEnvInjector systemenvinjector.SystemEnvInjectorInterface) (*v1.Secret, error) {
	computedEnv, err := systemEnvInjector.ComputeSystemEnv(app)
	if err != nil {
		return nil, err
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SecretName(app),
			Namespace: space.Name,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(app),
			},
			Labels: resources.UnionMaps(app.GetLabels(), app.ComponentLabels("secret")),
		},
	}

	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}

	for _, envVar := range computedEnv {
		secret.Data[envVar.Name] = []byte(envVar.Value)
	}

	return secret, nil
}
