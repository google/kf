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
)

// ownerName labels the owner of ClusterActiveOperand.
var ownerName = "clusteractiveoperand.kuberun.cloud.google.com/owner-name"

// ClusterActiveOperandOption enables further configuration of a ClusterActiveOperand.
type ClusterActiveOperandOption func(*v1alpha1.ClusterActiveOperand)

// ClusterActiveOperand creates a service with ClusterActiveOperandOptions.
func ClusterActiveOperand(ref metav1.OwnerReference, name string, cro ...ClusterActiveOperandOption) *v1alpha1.ClusterActiveOperand {
	ao := &v1alpha1.ClusterActiveOperand{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	ao.SetOwnerReferences([]metav1.OwnerReference{ref})
	for _, opt := range cro {
		opt(ao)
	}
	return ao
}

// ClusterActiveOperandWithOwnerLabel creates a service with ClusterActiveOperandOptions with owner label.
func ClusterActiveOperandWithOwnerLabel(ref metav1.OwnerReference, name string, cro ...ClusterActiveOperandOption) *v1alpha1.ClusterActiveOperand {
	return ClusterActiveOperand(ref, name, append(cro, WithLabel(ownerName, ref.Name))...)
}

// ClusterActiveOperandWithDefaults creates a service with defaults and ClusterActiveOperandOptions.
func ClusterActiveOperandWithDefaults(name string, cro ...ClusterActiveOperandOption) *v1alpha1.ClusterActiveOperand {
	options := append([]ClusterActiveOperandOption{ClusterWithDefaults}, cro...)
	return ClusterActiveOperand(metav1.OwnerReference{}, name, options...)
}

// ClusterWithDefaults initializes conditions of ClusterActiveOperand.
func ClusterWithDefaults(ao *v1alpha1.ClusterActiveOperand) {
	ao.Status.InitializeConditions()
}

// WithDelegates creates a ClusterActiveOperandOption that sets Delegates.
func WithDelegates(delegates ...string) ClusterActiveOperandOption {
	return func(ao *v1alpha1.ClusterActiveOperand) {
		ao.Status.Delegates = []v1alpha1.DelegateRef{}
		for _, del := range delegates {
			ao.Status.Delegates = append(ao.Status.Delegates, v1alpha1.DelegateRef{Namespace: del})
		}
	}
}

// ClusterWithLiveRefs creates a ClusterActiveOperandOption that sets LiveRefs.
func ClusterWithLiveRefs(refs ...v1alpha1.LiveRef) ClusterActiveOperandOption {
	return func(ao *v1alpha1.ClusterActiveOperand) {
		ao.Spec.Live = refs
	}
}

// ClusterWithOwnerRefsInjectedFailed creates a ClusterActiveOperandOption
// that marks OwnerRefs injected failed.
func ClusterWithOwnerRefsInjectedFailed(err string) ClusterActiveOperandOption {
	return func(ao *v1alpha1.ClusterActiveOperand) {
		ao.Status.MarkOwnerRefsInjectedFailed(err)
	}
}

// ClusterWithOwnerRefsInjected creates a ClusterActiveOperandOption
// that marks OwnerRefs injected.
func ClusterWithOwnerRefsInjected() ClusterActiveOperandOption {
	return func(ao *v1alpha1.ClusterActiveOperand) {
		ao.Status.MarkOwnerRefsInjected()
	}
}

// WithNamespaceDelegatesReadyFailed creates a ClusterActiveOperandOption
// that marks NamespaceDelegatesReady failed.
func WithNamespaceDelegatesReadyFailed(err string) ClusterActiveOperandOption {
	return func(ao *v1alpha1.ClusterActiveOperand) {
		ao.Status.MarkNamespaceDelegatesReadyFailed(err)
	}
}

// WithNamespaceDelegatesReady creates a ClusterActiveOperandOption
// that marks NamespaceDelegatesReady.
func WithNamespaceDelegatesReady() ClusterActiveOperandOption {
	return func(ao *v1alpha1.ClusterActiveOperand) {
		ao.Status.MarkNamespaceDelegatesReady()
	}
}

// WithClusterLiveRefs creates a ClusterActiveOperandOption
// that sets ClusterLive.
func WithClusterLiveRefs(refs ...v1alpha1.LiveRef) ClusterActiveOperandOption {
	return func(ao *v1alpha1.ClusterActiveOperand) {
		ao.Status.ClusterLive = refs
	}
}

// WithLabel creates an OperandOption that adds the lable.
func WithLabel(key string, val string) ClusterActiveOperandOption {
	return func(o *v1alpha1.ClusterActiveOperand) {
		if o.Labels == nil {
			o.Labels = make(map[string]string)
		}
		o.Labels[key] = val
	}
}
