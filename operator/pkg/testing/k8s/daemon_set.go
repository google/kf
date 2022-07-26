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

package k8s

import (
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DaemonSetOption enables further configuration of a DaemonSet.
type DaemonSetOption func(*apps.DaemonSet)

// DaemonSet creates a DaemonSet with Name name and Namespace test,
// and then applies DaemonSetOptions to it.
func DaemonSet(name string, do ...DaemonSetOption) *apps.DaemonSet {
	dep := &apps.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
	}
	for _, opt := range do {
		opt(dep)
	}
	return dep
}

// WithDaemonSetContainer creates a DaemonSetOption that updates Containers
// in DaemonSet.
func WithDaemonSetContainer(c *corev1.Container) DaemonSetOption {
	return func(dep *apps.DaemonSet) {
		dep.Spec.Template.Spec.Containers = append(dep.Spec.Template.Spec.Containers, *c)
	}
}
