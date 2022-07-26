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
	"errors"
	"fmt"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

const buildComponentName = "build"

// MakeBuildName creates the name of an Application's build.
func MakeBuildName(app *v1alpha1.App) string {
	return v1alpha1.GenerateName(app.Name, fmt.Sprintf("%d", app.Spec.Build.UpdateRequests))
}

// MakeBuild creates a Build for the given application.
func MakeBuild(app *v1alpha1.App, space *v1alpha1.Space) (*v1alpha1.Build, error) {
	buildSpec := app.Spec.Build.Spec
	if buildSpec == nil {
		return nil, errors.New("no build specified")
	}

	buildSpec = buildSpec.DeepCopy()

	// Set up environment for the Build, it's the Space build env, overridden by
	// the App env, overridden by the Build runtime environment variables.

	autoBuildEnv := []corev1.EnvVar{}
	autoBuildEnv = append(autoBuildEnv, space.Status.BuildConfig.Env...)
	if containers := app.Spec.Template.Spec.Containers; len(containers) > 0 {
		autoBuildEnv = append(autoBuildEnv, containers[0].Env...)
	}

	// Add in additinal CF style environment variables
	autoBuildEnv = append(autoBuildEnv, BuildRuntimeEnvVars(CFStaging, app)...)

	// Add in the environment variables from the declared Build spec.
	buildSpec.Env = append(autoBuildEnv, buildSpec.Env...)

	return &v1alpha1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:      MakeBuildName(app),
			Namespace: app.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(app),
			},
			Labels: v1alpha1.UnionMaps(app.GetLabels(), app.ComponentLabels(buildComponentName)),
		},
		Spec: *buildSpec,
	}, nil
}
