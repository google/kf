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

package doctor

import (
	"context"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
)

// OperatorDiagnostic tests whether the Operator is properly installaed.
type OperatorDiagnostic struct {
	kubeClient kubernetes.Interface
}

var _ Diagnosable = (*OperatorDiagnostic)(nil)

// Diagnose validates the operator components are properly installead.
func (o *OperatorDiagnostic) Diagnose(ctx context.Context, d *Diagnostic) {
	d.Run(ctx, "Operator", func(ctx context.Context, d *Diagnostic) {
		diagnoseOperator(ctx, d, o.kubeClient, o.kubeClient.Discovery())
	})
}

func diagnoseOperator(ctx context.Context, d *Diagnostic, kubernetes kubernetes.Interface, vc discovery.ServerResourcesInterface) {
	expectedDeployments := []expectedDeployments{
		{
			Namespace:   "appdevexperience",
			Deployments: []string{"appdevexperience-operator"},
		},
	}
	diagnoseDeployments(expectedDeployments, ctx, d, kubernetes)

	expectedComponents := map[string]expectedComponent{
		"Operand": {
			groupVersion:      "operand.run.cloud.google.com/v1alpha1",
			expectedResources: []string{"clusteractiveoperands", "activeoperands", "operands"},
		},
		"kfsystem": {
			groupVersion:      "kf.dev/v1alpha1",
			expectedResources: []string{"kfsystems"},
		},
	}
	diagnoseComponents(expectedComponents, ctx, d, vc)
}

// NewOperatorDiagnostic creates a new OperatorDiagnostic to validate the Operator.
func NewOperatorDiagnostic(
	kubeClient kubernetes.Interface,
) *OperatorDiagnostic {
	return &OperatorDiagnostic{
		kubeClient: kubeClient,
	}
}
