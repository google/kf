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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigMapOption enables further configuration of a ConfigMap.
type ConfigMapOption func(*corev1.ConfigMap)

// WithData creates a ConfigMapOption that updates key, value
// in ConfigMap.
func WithData(key, value string) ConfigMapOption {
	return func(cm *corev1.ConfigMap) {
		if cm.Data == nil {
			cm.Data = map[string]string{}
		}
		cm.Data[key] = value
	}
}

// WithConfigMapNamespace configures the ConfigMap to have the given
// namespace.
func WithConfigMapNamespace(ns string) ConfigMapOption {
	return func(cm *corev1.ConfigMap) {
		cm.Namespace = ns
	}
}

// ConfigMap creates a ConfigMap with Name name and Namespace test,
// and then applies ConfigMapOptions to it.
func ConfigMap(name string, do ...ConfigMapOption) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
	}
	for _, opt := range do {
		opt(cm)
	}
	return cm
}
