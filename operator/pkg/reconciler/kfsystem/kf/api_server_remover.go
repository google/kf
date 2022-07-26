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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
)

func init() {
	utilruntime.Must(apiregistrationv1.AddToScheme(scheme.Scheme))
}

// APIServerRemoverTransformer removes the API service and puts a ConfigMap in
// its place.
func APIServerRemoverTransformer(ctx context.Context) mf.Transformer {
	return func(u *unstructured.Unstructured) error {
		if u.GetKind() != "APIService" || u.GetName() != "v1alpha1.upload.kf.dev" {
			return nil
		}

		// Replace the API service with a ConfigMap. Our framework doesn't
		// allow us to simply remove it. It will also be nice to have the
		// place holder ConfigMap for troubleshooting purposes later. For
		// example, if we see the ConfigMap, then we know that the operator
		// hasn't seen the necessary pods.
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "apiservice-placeholder",
				Namespace: kfNamespace,
			},
			Data: map[string]string{
				"placeholder": "This config map implies that the operator didn't see the necessary pods for the API Service.",
			},
		}

		return scheme.Scheme.Convert(cm, u, nil)
	}
}
