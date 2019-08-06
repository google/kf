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
	"strconv"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/systemenvinjector"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	servingv1beta1 "github.com/knative/serving/pkg/apis/serving/v1beta1"
	"github.com/knative/serving/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

// KnativeServiceName gets the name of a Knative Service given the route.
func KnativeServiceName(app *v1alpha1.App) string {
	return app.Name
}

// MakeKnativeService creates a KnativeService from an app definition.
func MakeKnativeService(
	app *v1alpha1.App,
	space *v1alpha1.Space,
	systemEnvInjector systemenvinjector.SystemEnvInjectorInterface,
) (*serving.Service, error) {

	image := app.Status.Image
	if image == "" {
		return nil, errors.New("waiting for source image in latestReadySource")
	}

	// don't modify the spec on the app
	podSpec := app.Spec.Template.Spec.DeepCopy()

	// XXX: Add a dummy environment variable that reflects the UpdateRequests.
	// This will cause knative to create a new revision of the service.
	podSpec.Containers[0].Env = append(
		podSpec.Containers[0].Env,
		corev1.EnvVar{
			Name:  fmt.Sprintf("KF_UPDATE_REQUESTS_%v", app.UID),
			Value: strconv.FormatInt(int64(app.Spec.Template.UpdateRequests), 10),
		},
	)

	// At this point in the lifecycle there should be exactly one container
	// if the webhhook is working but create one to avoid panics just in case.
	if len(podSpec.Containers) == 0 {
		podSpec.Containers = append(podSpec.Containers, corev1.Container{})
	}
	podSpec.Containers[0].Image = image
	// Execution environment variables come before others because they're built
	// to be overridden.
	podSpec.Containers[0].Env = append(space.Spec.Execution.Env, podSpec.Containers[0].Env...)

	// Inject VCAP env vars from secret
	podSpec.Containers[0].EnvFrom = []corev1.EnvFromSource{
		corev1.EnvFromSource{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: SecretName(app, space),
				},
			},
		},
	}

	return &serving.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KnativeServiceName(app),
			Namespace: app.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(app),
			},
			Labels: resources.UnionMaps(app.GetLabels(), app.ComponentLabels("app-scaler")),
		},
		Spec: serving.ServiceSpec{
			ConfigurationSpec: serving.ConfigurationSpec{
				Template: &serving.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels:      app.ComponentLabels("app-server"),
						Annotations: app.Spec.Instances.ScalingAnnotations(),
					},
					Spec: serving.RevisionSpec{
						RevisionSpec: servingv1beta1.RevisionSpec{
							PodSpec: *podSpec,
						},
					},
				},
			},
		},
	}, nil
}
