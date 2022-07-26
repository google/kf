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

package resources

import (
	"errors"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

// AutoscalerName gets the name of a Deployment given the app.
func AutoscalerName(app *v1alpha1.App) string {
	return app.Name
}

// MakeHorizontalPodAutoScaler creates a HorizontalPodAutoScaler from an app definition.
func MakeHorizontalPodAutoScaler(
	app *v1alpha1.App,
) (*autoscalingv1.HorizontalPodAutoscaler, error) {

	if app.Spec.Instances.Stopped || !app.Spec.Instances.Autoscaling.RequiresHPA() {
		return nil, nil
	}

	if len(app.Spec.Instances.Autoscaling.Rules) > 1 {
		return nil, errors.New("too many autoscaling rules")
	}

	autoscaler := &autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AutoscalerName(app),
			Namespace: app.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(app),
			},
			Labels: v1alpha1.UnionMaps(app.GetLabels(), app.ComponentLabels("autoscaler")),
		},
		Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
				APIVersion: app.GetGroupVersionKind().GroupVersion().String(),
				Kind:       app.GetGroupVersionKind().Kind,
				Name:       app.Name,
			},
			MinReplicas: app.Spec.Instances.Autoscaling.MinReplicas,
			MaxReplicas: *app.Spec.Instances.Autoscaling.MaxReplicas,
		},
	}

	rule := app.Spec.Instances.Autoscaling.Rules[0]
	if rule.RuleType != v1alpha1.CPURuleType {
		return nil, errors.New("invalid autoscaling rule")
	}

	autoscaler.Spec.TargetCPUUtilizationPercentage = rule.Target

	return autoscaler, nil
}
