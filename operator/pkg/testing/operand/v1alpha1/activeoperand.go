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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ActiveOperandOption enables further configuration of a ActiveOperand.
type ActiveOperandOption func(*v1alpha1.ActiveOperand)

// ActiveOperand creates a service with ActiveOperandOptions.
func ActiveOperand(name, namespace string, cro ...ActiveOperandOption) *v1alpha1.ActiveOperand {
	ao := &v1alpha1.ActiveOperand{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	for _, opt := range cro {
		opt(ao)
	}
	return ao
}

// ActiveOperandWithDefaults creates an ActiveOperand and applies ActiveOperandOptions to it.
func ActiveOperandWithDefaults(name, namespace string, cro ...ActiveOperandOption) *v1alpha1.ActiveOperand {
	options := append([]ActiveOperandOption{WithDefaults}, cro...)
	return ActiveOperand(name, namespace, options...)
}

// WithDefaults initializes conditions of ActiveOperand.
func WithDefaults(ao *v1alpha1.ActiveOperand) {
	ao.Status.InitializeConditions()
}

// WithLiveRefs creates an ActiveOperandOption that sets LiveRefs.
func WithLiveRefs(refs ...v1alpha1.LiveRef) ActiveOperandOption {
	return func(ao *v1alpha1.ActiveOperand) {
		ao.Spec.Live = refs
	}
}

// WithOwnerRefsInjectedFailed creates an ActiveOperandOption
// that marks OwnerRefs injected failed.
func WithOwnerRefsInjectedFailed(err string) ActiveOperandOption {
	return func(ao *v1alpha1.ActiveOperand) {
		ao.Status.MarkOwnerRefsInjectedFailed(err)
	}
}

// WithOwnerRefsInjected creates an ActiveOperandOption
// that marks OwnerRefs injected.
func WithOwnerRefsInjected() ActiveOperandOption {
	return func(ao *v1alpha1.ActiveOperand) {
		ao.Status.MarkOwnerRefsInjected()
	}
}

// CreateLiveRef returns a new LiveRef for the details provided.
func CreateLiveRef(name string, namespace string, gk schema.GroupKind) v1alpha1.LiveRef {
	return v1alpha1.LiveRef{
		Group:     gk.Group,
		Kind:      gk.Kind,
		Name:      name,
		Namespace: namespace,
	}
}
