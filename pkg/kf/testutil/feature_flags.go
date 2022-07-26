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

package testutil

import (
	"context"
	"encoding/json"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/client/injection/kube/client/fake"
)

// WithFeatureFlags returns a context that will have a fake namespace injected
// into the kubeClient.
func WithFeatureFlags(ctx context.Context, t *testing.T, ff map[string]bool) context.Context {
	t.Helper()
	data, err := json.Marshal(ff)
	AssertErrorsEqual(t, nil, err)

	fake.Get(ctx).CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kf",
			Annotations: map[string]string{
				// NOTE: We can't directly use the const in v1alpha1 because
				// it uses the testutil package and therefore we would have a
				// circular dependency.
				"kf.dev/feature-flags": string(data),
			},
		},
	}, metav1.CreateOptions{})
	return ctx
}
