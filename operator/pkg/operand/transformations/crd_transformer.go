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

package transformations

import (
	"context"

	mf "github.com/manifestival/manifestival"
	"go.uber.org/zap"
	apixinstall "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/install"
	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"
	"knative.dev/pkg/logging"
)

func init() {
	apixinstall.Install(scheme.Scheme)
}

// CRDTransformer returns a Transformer for CRD.
func CRDTransformer(ctx context.Context) mf.Transformer {
	logger := logging.FromContext(ctx)
	return func(u *unstructured.Unstructured) error {
		if u.GetKind() == "CustomResourceDefinition" && u.GetAPIVersion() == "apiextensions.k8s.io/v1" {
			return updateCustomResourceDefinition(u, logger)
		}
		return nil
	}
}

func updateCustomResourceDefinition(u *unstructured.Unstructured, log *zap.SugaredLogger) error {
	var crd = &apixv1.CustomResourceDefinition{}
	if err := scheme.Scheme.Convert(u, crd, nil); err != nil {
		log.Error(err, "Error converting Unstructured to CustomResourceDefinition", "unstructured", u, "crd", crd)
		return err
	}

	// Handle scenario where preserveUnknownFields is not reflecting the expected behavior of the webhook converter
	// https://github.com/kubernetes/apiextensions-apiserver/blob/kubernetes-1.18.10/pkg/apis/apiextensions/validation/validation.go#L300
	// If there is a WebhookConverter, preserveUnknownFields must be set to false (or unset).
	// However this explicitly sets it to false, in the case where the stored crd is set to true
	if crd.Spec.Conversion != nil && crd.Spec.Conversion.Strategy != apixv1.NoneConverter {
		// Unfortunately setting the typed `crd.Spec.PreserveUnknownFields` to false
		// does not explicitly set preserveUnknownFields in the unstructured type.
		// And thus does not overwrite preserveUnknownFields from true to false
		// https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#specifying-a-structural-schema
		unstructured.SetNestedField(u.Object, false, "spec", "preserveUnknownFields")
		log.Debugw("Set preserveUnknownFields", "name", u.GetName(), "unstructured", u.Object)
	}

	log.Debugw("Finished conversion", "name", u.GetName(), "unstructured", u.Object)
	return nil
}
