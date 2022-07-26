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

package v1alpha1

import (
	"kf-operator/pkg/apis/operand/v1alpha1"
	"kf-operator/pkg/operand"
	kfoperand "kf-operator/pkg/operand/kf"
	"kf-operator/pkg/transformer"
	"testing"

	manifestivalHelpers "kf-operator/pkg/testing/manifestival"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// TestFactory is provided to allow for factory creation without required pod annotations.
var TestFactory operand.Factory = operand.CreateFactory(transformer.Annotation{PodAnnotations: map[string]string{}})

// OperandOption enables further configuration of a Operand.
type OperandOption func(*v1alpha1.Operand)

// Operand creates a service with OperandOptions.
func Operand(name string, cro ...OperandOption) *v1alpha1.Operand {
	o := &v1alpha1.Operand{}
	o.SetName(name)
	for _, opt := range cro {
		opt(o)
	}
	return o
}

// OperandWithDefaults creates a service with defaults and OperandOptions.
func OperandWithDefaults(name string, cro ...OperandOption) *v1alpha1.Operand {
	options := append([]OperandOption{operandWithDefaults}, cro...)
	return Operand(name, options...)
}

func operandWithDefaults(ao *v1alpha1.Operand) {
	ao.Status.InitializeConditions()
}

// WithOperandInstallFailed creates an OperandOption that
// marks OperandInstall failed.
func WithOperandInstallFailed(err error) OperandOption {
	return func(o *v1alpha1.Operand) {
		o.Status.MarkOperandInstallFailed(err)
	}
}

// WithClusterOwner sets the given cluster owner via labels.
func WithClusterOwner(name string) OperandOption {
	return func(o *v1alpha1.Operand) {
		var labels map[string]string
		if labels = o.GetLabels(); labels == nil {
			labels = make(map[string]string)
		}
		labels[kfoperand.OwnerName] = name
		o.SetLabels(labels)
	}
}

// WithOwner sets the given owner via labels.
func WithOwner(name, ns string) OperandOption {
	return func(o *v1alpha1.Operand) {
		var labels map[string]string
		if labels = o.GetLabels(); labels == nil {
			labels = make(map[string]string)
		}
		labels[kfoperand.OwnerNamespace] = ns
		labels[kfoperand.OwnerName] = name
		o.SetLabels(labels)
	}
}

// Generations sets the current and observed generation
// on an Operand.
func Generations(generation, observed int64) OperandOption {
	return func(o *v1alpha1.Operand) {
		o.SetGeneration(generation)
		o.Status.ObservedGeneration = observed
	}
}

// WithInstalledSteadyStateGeneration sets
// InstalledSteadyStateGeneration in the Operand's Status.
func WithInstalledSteadyStateGeneration(generation int64) OperandOption {
	return func(o *v1alpha1.Operand) {
		o.Status.InstalledSteadyStateGeneration = generation
	}
}

// WithOperandInstallNotReady creates an OperandOption that
// marks OperandInstall not ready.
func WithOperandInstallNotReady(msg string) OperandOption {
	return func(o *v1alpha1.Operand) {
		o.Status.MarkOperandInstallNotReady(msg)
	}
}

// WithOperandInstallSuccessful creates an OperandOption that
// marks OperandInstall successful.
func WithOperandInstallSuccessful() OperandOption {
	return func(o *v1alpha1.Operand) {
		o.Status.MarkOperandInstallSuccessful()
	}
}

// WithOperandPostInstallNotReady creates an OperandOption that marks
// the OperandPostInstall condition not ready.
func WithOperandPostInstallNotReady(msg string) OperandOption {
	return func(o *v1alpha1.Operand) {
		o.Status.MarkOperandPostInstallNotReady(msg)
	}
}

// WithOperandPostInstallFailed created an OperandOption that marks
// the OperandPostInstall condition failed.
func WithOperandPostInstallFailed(err error) OperandOption {
	return func(o *v1alpha1.Operand) {
		o.Status.MarkOperandPostInstallFailed(err)
	}
}

// WithResetLatestCreatedActiveOperand creates an OperandOption that
// resets LatestCreatedActiveOperand.
func WithResetLatestCreatedActiveOperand(operand string) OperandOption {
	return func(o *v1alpha1.Operand) {
		o.Status.ResetLatestCreatedActiveOperand(operand)
	}
}

// WithCheckDeploymentHealth sets CheckDeploymentHealth true.
func WithCheckDeploymentHealth(o *v1alpha1.Operand) {
	o.Spec.CheckDeploymentHealth = true
}

// ToUnstructured converts objs to unstructured.Unstructured.
func ToUnstructured(objs ...runtime.Object) []unstructured.Unstructured {
	ss := make([]unstructured.Unstructured, len(objs))
	for i, obj := range objs {
		u := manifestivalHelpers.ToUnstructured(obj)
		if u.GetKind() == "Deployment" {
			unstructured.RemoveNestedField(u.Object, "metadata", "creationTimestamp")
		}
		ss[i] = *u
	}
	return ss
}

// WithSteadyState creates an OperandOption that sets SteadyState.
func WithSteadyState(t *testing.T, objects ...runtime.Object) OperandOption {
	return func(o *v1alpha1.Operand) {
		o.Spec.SteadyState = ToUnstructured(objects...)
	}
}

// WithSteadyStateNoStatus creates an OperandOption that sets SteadyState without Status field.
func WithSteadyStateNoStatus(t *testing.T, objects ...runtime.Object) OperandOption {
	return func(o *v1alpha1.Operand) {
		withSteadyState := WithSteadyState(t, objects...)
		withSteadyState(o)
		for i := range o.Spec.SteadyState {
			unstructured.RemoveNestedField(o.Spec.SteadyState[i].Object, "status")
		}
	}
}

// WithPostInstall creates an OperandOption that populates PostInstall with the given objects.
func WithPostInstall(t *testing.T, objects ...runtime.Object) OperandOption {
	return func(o *v1alpha1.Operand) {
		o.Spec.PostInstall = ToUnstructured(objects...)
	}
}

// WithLatestActiveOperandCreated creates an OperandOption
// that marks LatestActiveOperand created.
func WithLatestActiveOperandCreated(operand string) OperandOption {
	return func(o *v1alpha1.Operand) {
		o.Status.MarkLatestActiveOperandCreated(operand)
	}
}

// WithLatestActiveOperandReady creates an OperandOption
// that marks LatestActiveOperand ready.
func WithLatestActiveOperandReady(operand string) OperandOption {
	return func(o *v1alpha1.Operand) {
		o.Status.MarkLatestActiveOperandReady(operand)
	}
}
