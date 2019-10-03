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
	"path"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

const buildComponentName = "build"

// MakeSourceName creates the name of an Application's source.
func MakeSourceName(app *v1alpha1.App) string {
	return fmt.Sprintf("%s-%x", app.Name, app.Spec.Source.UpdateRequests)
}

// BuildpackBuildImageDestination gets the image name for an application build.
func BuildpackBuildImageDestination(app *v1alpha1.App, space *v1alpha1.Space) string {
	registry := space.Spec.BuildpackBuild.ContainerRegistry

	// Use underscores because those aren't permitted in k8s names so you can't
	// cause accidental conflicts.
	image := fmt.Sprintf("app_%s_%s:%x", app.Namespace, app.Name, app.Spec.Source.UpdateRequests)

	return path.Join(registry, image)
}

// MakeSource creates a source for the given application.
func MakeSource(app *v1alpha1.App, space *v1alpha1.Space) (*v1alpha1.Source, error) {
	source := app.Spec.Source.DeepCopy()

	source.ServiceAccount = space.Spec.Security.BuildServiceAccount

	switch {
	case source.IsBuildpackBuild():
		// user defined values in buildpackbuild.env take priority from buildpackbuild.env
		source.BuildpackBuild.Env = append(space.Spec.BuildpackBuild.Env, source.BuildpackBuild.Env...)
		source.BuildpackBuild.Image = BuildpackBuildImageDestination(app, space)
		source.BuildpackBuild.BuildpackBuilder = space.Spec.BuildpackBuild.BuilderImage

	case source.IsDockerfileBuild():
		source.Dockerfile.Image = BuildpackBuildImageDestination(app, space)
	}

	return &v1alpha1.Source{
		ObjectMeta: metav1.ObjectMeta{
			Name:      MakeSourceName(app),
			Namespace: app.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(app),
			},
			Labels: v1alpha1.UnionMaps(app.GetLabels(), app.ComponentLabels(buildComponentName)),
		},
		Spec: *source,
	}, nil
}
