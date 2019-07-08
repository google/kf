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
	"github.com/knative/serving/pkg/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

// MakeSourceLabels creates labels that can be used to tie a source to a build.
func MakeSourceLabels(app *v1alpha1.App) map[string]string {
	return app.ComponentLabels("build")
}

// MakeSource creates a source for the given application.
func MakeSource(app *v1alpha1.App, space *v1alpha1.Space) (*v1alpha1.Source, error) {
	source := app.Spec.Source.DeepCopy()

	source.SetSpaceDefaults(space)

	return &v1alpha1.Source{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", app.Name),
			Namespace:    app.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(app),
			},
			Labels: resources.UnionMaps(app.GetLabels(), MakeSourceLabels(app)),
		},
		Spec: *source,
	}, nil
}
