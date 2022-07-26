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
)

// VolumeOption enables further configuration of a Volume.
type VolumeOption func(*corev1.Volume)

// Volume creates a Volume with Name name, and then applies VolumeOptions to
// it.
func Volume(name string, do ...VolumeOption) *corev1.Volume {
	vol := &corev1.Volume{
		Name: name,
	}
	for _, opt := range do {
		opt(vol)
	}
	return vol
}

// WithVolumeSecretSource creates a VolumeOption that sets the Secret field.
func WithVolumeSecretSource(s *corev1.SecretVolumeSource) VolumeOption {
	return func(v *corev1.Volume) {
		v.VolumeSource.Secret = s
	}
}
