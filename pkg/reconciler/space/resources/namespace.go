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
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/knative/pkg/kmeta"
	"github.com/knative/serving/pkg/resources"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	managedByLabel      = "app.kubernetes.io/managed-by"
	istioInjectionLabel = "istio-injection"
)

// NamespaceName gets the name of a namespace given the space.
func NamespaceName(space *v1alpha1.Space) string {
	return space.Name
}

// MakeNamespace creates a Namespace from a Space object.
func MakeNamespace(space *v1alpha1.Space) (*v1.Namespace, error) {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: NamespaceName(space),
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(space),
			},
			// Copy labels from the parent and enable istio-injection on the namespace
			// so services can communicate in an East/West direction rather than just
			// North/South.
			Labels: resources.UnionMaps(
				space.GetLabels(), map[string]string{
					istioInjectionLabel: "enabled",
					managedByLabel:      "kf",
				}),
		},
	}, nil
}
