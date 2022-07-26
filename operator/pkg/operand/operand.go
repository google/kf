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
	operandv1alpha1 "kf-operator/pkg/apis/operand/v1alpha1"
	"kf-operator/pkg/operand/transformations"
	"kf-operator/pkg/transformer"

	opclient "kf-operator/pkg/client/clientset/versioned/typed/operand/v1alpha1"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	mf "github.com/manifestival/manifestival"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/logging"
)

// KfOperandName is the name of Operand for KfSystem Kf.
const KfOperandName = "kf"

// Factory encapsulates shared logic needed to create an OperandSpec from a desired state.
type Factory interface {
	// FromGeneralManifest creates an operand spec from a manifest without cloudrun specific configs.
	FromGeneralManifest(ctx context.Context, healthCheck bool, m mf.Manifest, transforms ...mf.Transformer) (*operandv1alpha1.OperandSpec, error)
}

type operandFactory struct {
	annotations transformer.Annotation
}

// CreateFactory makes an OperandFactory, returning an error if none can be created.
func CreateFactory(a transformer.Annotation) Factory {
	return &operandFactory{annotations: a}
}

// FromGeneralManifest creates an operand spec from a manifest without cloudrun specific configs.
func (o operandFactory) FromGeneralManifest(ctx context.Context, healthCheck bool, m mf.Manifest, transforms ...mf.Transformer) (*operandv1alpha1.OperandSpec, error) {
	transforms = append(
		transforms,
		o.annotations.Transform(ctx),
		transformations.CRDTransformer(ctx),
		transformations.AddMissingProtocol(ctx),
	)
	m, err := m.Transform(transforms...)
	if err != nil {
		return nil, err
	}
	return &operandv1alpha1.OperandSpec{
		CheckDeploymentHealth: healthCheck,
		SteadyState:           m.Resources(),
	}, nil
}

// EnsureOperand creates or updates an Operand with the given name and spec, and sets the given labels.
func EnsureOperand(ctx context.Context, name string, spec operandv1alpha1.OperandSpec, client opclient.OperandV1alpha1Interface, labels map[string]string) (*operandv1alpha1.Operand, error) {
	e, err := client.Operands().Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		op := &operandv1alpha1.Operand{
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: labels,
			},
			Spec: spec,
		}
		// TODO(b/170672490): Readd after KubeRun is cluster scoped.
		// op.SetOwnerReferences([]metav1.OwnerReference{*o.ownerRef})
		return client.Operands().Create(ctx, op, metav1.CreateOptions{})
	} else if err != nil {
		return nil, err
	}

	labelDiff := cmp.Diff(e.ObjectMeta.Labels, labels, cmpopts.EquateEmpty())
	diff := cmp.Diff(e.Spec, spec, cmpopts.EquateEmpty())
	if diff == "" && labelDiff == "" {
		return e, nil
	}
	logging.FromContext(ctx).Info("Saw spec difference", diff)
	logging.FromContext(ctx).Info("Saw label difference", labelDiff)

	e.Spec = spec
	e.SetLabels(labels)
	return client.Operands().Update(ctx, e, metav1.UpdateOptions{})
}
