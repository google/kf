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

package operand

import (
	"context"
	"fmt"

	mf "github.com/manifestival/manifestival"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	poperand "kf-operator/pkg/operand"
)

const (
	// HealthCheckDeploymentsNotReady is the state that deployments are not ready.
	HealthCheckDeploymentsNotReady = "DEPLOYMENTS_NOT_READY"
)

type healthcheckReconciler struct {
	delegate poperand.ResourceReconciler
	kc       kubernetes.Interface
}

// NewHealthcheckReconciler creates a healthchecking ResourceReconciler which delegates to delegate.
func NewHealthcheckReconciler(delegate poperand.ResourceReconciler, kc kubernetes.Interface) poperand.ResourceReconciler {
	return healthcheckReconciler{delegate: delegate, kc: kc}
}

func (r healthcheckReconciler) Apply(ctx context.Context, resources []unstructured.Unstructured) error {
	return r.delegate.Apply(ctx, resources)

}

// GetState checks that the installation is completed and that all deployments are ready.
func (r healthcheckReconciler) GetState(ctx context.Context, resources []unstructured.Unstructured) (string, error) {
	if genericResult, err := r.delegate.GetState(ctx, resources); genericResult != poperand.Installed || err != nil {
		return genericResult, err
	}
	return r.checkAllDeploymentsReady(ctx, resources)
}

func (r healthcheckReconciler) checkAllDeploymentsReady(ctx context.Context, resources []unstructured.Unstructured) (string, error) {
	for _, u := range resources {
		if u.GetKind() == "Deployment" {
			if d, err := r.getDeployment(ctx, u); err != nil {
				return poperand.Error, err
			} else if d != nil {
				if ready := r.checkDeploymentAvailable(d, corev1.ConditionTrue); !ready {
					if d.GetName() == "controller" && r.checkDeploymentAvailable(d, corev1.ConditionFalse) {
						// We treat explicitly unavailable controller as unhealthy, not unready.
						return poperand.Error, fmt.Errorf("controller in ns %s is not healthy", d.GetNamespace())
					}
					return HealthCheckDeploymentsNotReady, nil
				}
			}
		}
	}
	return poperand.Installed, nil
}

func (r healthcheckReconciler) getDeployment(ctx context.Context, u unstructured.Unstructured) (*appsv1.Deployment, error) {
	deployment, err := r.kc.AppsV1().Deployments(u.GetNamespace()).Get(ctx, u.GetName(), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return deployment, nil
}

func (r healthcheckReconciler) checkDeploymentAvailable(deployment *appsv1.Deployment, cs corev1.ConditionStatus) bool {
	available := false
	for _, c := range deployment.Status.Conditions {
		if c.Type == appsv1.DeploymentAvailable && c.Status == cs {
			available = true
			break
		}
	}
	return available
}

// byGroupKind returns resources matching a particular group and kind
func byGroupKind(kind, group string) mf.Predicate {
	return func(u *unstructured.Unstructured) bool {
		groupKind := u.GroupVersionKind().GroupKind()
		return groupKind.Kind == kind && groupKind.Group == group
	}
}
