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

package apps

import (
	"context"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"

	"github.com/google/kf/v2/pkg/internal/envutil"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KfApp provides a facade around Knative services for accessing and mutating its
// values.
type KfApp v1alpha1.App

func (k *KfApp) getOrCreateContainer() *corev1.Container {
	rl := &k.Spec.Template
	if len(rl.Spec.Containers) == 0 {
		rl.Spec.Containers = []v1.Container{{}}
	}

	return &k.Spec.Template.Spec.Containers[0]
}

// GetEnvVars reads the environment variables off an app.
func (k *KfApp) GetEnvVars() []corev1.EnvVar {
	if k == nil {
		return nil
	}

	if rl := &k.Spec.Template; rl != nil && len(rl.Spec.Containers) != 0 {
		return rl.Spec.Containers[0].Env
	}

	return nil
}

// SetEnvVars sets environment variables on an app.
func (k *KfApp) SetEnvVars(env []corev1.EnvVar) {
	k.getOrCreateContainer().Env = env
}

// MergeEnvVars adds the environment variables listed to the existing ones,
// overwriting duplicates by key.
func (k *KfApp) MergeEnvVars(env []corev1.EnvVar) {
	k.SetEnvVars(envutil.DeduplicateEnvVars(append(k.GetEnvVars(), env...)))
}

// DeleteEnvVars removes environment variables with the given key.
func (k *KfApp) DeleteEnvVars(names []string) {
	k.SetEnvVars(envutil.RemoveEnvVars(names, k.GetEnvVars()))
}

// Set a resource request for an app. Request amount can be cleared by passing in nil
func (k *KfApp) setResourceRequest(r v1.ResourceName, quantity *resource.Quantity) {
	container := k.getOrCreateContainer()
	resourceRequests := container.Resources.Requests

	if resourceRequests == nil {
		resourceRequests = v1.ResourceList{}
	}

	if quantity == nil {
		delete(resourceRequests, r)
	} else {
		resourceRequests[r] = *quantity
	}
	container.Resources.Requests = resourceRequests
}

// MergeRoute adds a route to the App, removing any duplicates that already
// exist.
func (k *KfApp) MergeRoute(route v1alpha1.RouteWeightBinding) {
	k.RemoveRoute(context.Background(), route)
	k.Spec.Routes = append(k.Spec.Routes, route)
}

// RemoveRoute removes any routes matching the binding.
func (k *KfApp) RemoveRoute(ctx context.Context, toRemove v1alpha1.RouteWeightBinding) {
	k.deleteMatchingRoutes(func(route v1alpha1.RouteWeightBinding) bool {
		return route.EqualsBinding(ctx, toRemove)
	})
}

// HasMatchingRoutes checks if any of the listed routes point to the claim.
func (k *KfApp) HasMatchingRoutes(claim v1alpha1.RouteSpecFields) bool {
	for _, route := range k.Spec.Routes {
		if route.RouteSpecFields.Equals(claim) {
			return true
		}
	}

	return false
}

// RemoveRoutesForClaim removes all routes matching the given claim.
func (k *KfApp) RemoveRoutesForClaim(claim v1alpha1.RouteSpecFields) {
	k.deleteMatchingRoutes(func(route v1alpha1.RouteWeightBinding) bool {
		return route.RouteSpecFields.Equals(claim)
	})
}

func (k *KfApp) deleteMatchingRoutes(matcher func(v1alpha1.RouteWeightBinding) bool) {
	var notMatching []v1alpha1.RouteWeightBinding
	for _, binding := range k.Spec.Routes {
		if !matcher(binding) {
			notMatching = append(notMatching, binding)
		}
	}

	k.Spec.Routes = notMatching
}

// ToApp casts this alias back into an App.
func (k *KfApp) ToApp() *v1alpha1.App {
	app := v1alpha1.App(*k)
	return &app
}

// NewKfApp creates a new KfApp.
func NewKfApp() KfApp {
	return KfApp{
		TypeMeta: metav1.TypeMeta{
			Kind:       "App",
			APIVersion: "kf.dev/v1alpha1",
		},
		Spec: v1alpha1.AppSpec{
			Template: v1alpha1.AppSpecTemplate{
				Spec: v1.PodSpec{
					Containers: []v1.Container{{}},
				},
			},
		},
	}
}

// NewFromApp creates a new KfApp from the given service pointer
// modifications to the KfApp will affect the underling app.
func NewFromApp(app *v1alpha1.App) *KfApp {
	return (*KfApp)(app)
}
