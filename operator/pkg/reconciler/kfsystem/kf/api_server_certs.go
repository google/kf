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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	"knative.dev/pkg/logging"
)

func init() {
	utilruntime.Must(apiregistrationv1.AddToScheme(scheme.Scheme))
}

// APIServerCertsTransformer transforms built-in API services to have the given
// CACert.
func APIServerCertsTransformer(ctx context.Context, caCert []byte) mf.Transformer {
	log := logging.FromContext(ctx)

	return func(u *unstructured.Unstructured) error {
		if u.GetKind() != "APIService" || u.GetName() != "v1alpha1.upload.kf.dev" {
			return nil
		}

		apiService := &apiregistrationv1.APIService{}

		if err := scheme.Scheme.Convert(u, apiService, nil); err != nil {
			log.Error(err, "Error converting Unstructured to APIService", "unstructured", u, "apiservice", apiService)
			return err
		}

		apiService.Spec.InsecureSkipTLSVerify = false
		apiService.Spec.CABundle = caCert

		return scheme.Scheme.Convert(apiService, u, nil)
	}
}
