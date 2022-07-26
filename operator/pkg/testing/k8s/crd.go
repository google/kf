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

package k8s

import (
	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CRDOption enables further configuration of a CustomResourceDefinition.
type CRDOption func(*apixv1.CustomResourceDefinition)

// WithPreserveUnknownFields creates a CRDOption that sets PreserveUnknownFields
// in CustomResourceDefinition.
func WithPreserveUnknownFields(preserveUnknownFields bool) CRDOption {
	return func(crd *apixv1.CustomResourceDefinition) {
		crd.Spec.PreserveUnknownFields = preserveUnknownFields
	}
}

// WithConversionStrategy creates a CRDOption that sets Strategy
// in CustomResourceDefinition.
func WithConversionStrategy(strategy apixv1.ConversionStrategyType) CRDOption {
	return func(crd *apixv1.CustomResourceDefinition) {
		if crd.Spec.Conversion == nil {
			crd.Spec.Conversion = &apixv1.CustomResourceConversion{}
		}
		crd.Spec.Conversion.Strategy = strategy
	}
}

// CRD creates a CustomResourceDefinition with Name name
// and then applies CRDOptions to it.
func CRD(name string, do ...CRDOption) *apixv1.CustomResourceDefinition {
	crd := &apixv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	for _, opt := range do {
		opt(crd)
	}
	return crd
}
