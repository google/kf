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

package testing

import (
	fakeoperatorclientset "kf-operator/pkg/client/clientset/versioned/fake"

	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"
)

// NewScheme creates a Scheme.
func NewScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()

	fakekubeclientset.AddToScheme(scheme)
	fakeoperatorclientset.AddToScheme(scheme)
	apixv1.AddToScheme(scheme)

	return scheme
}
