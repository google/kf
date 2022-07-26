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
	"kf-operator/pkg/apis/kfsystem/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/ptr"
)

// KfSystemOption enables further configuration of a KfSystem.
type KfSystemOption func(*v1alpha1.KfSystem)

// KfSystem creates a service with CloudRunOptions.
func KfSystem(name string, kfo ...KfSystemOption) *v1alpha1.KfSystem {
	kfs := &v1alpha1.KfSystem{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	for _, opt := range kfo {
		opt(kfs)
	}
	return kfs
}

// KfSystemWithDefaults creates a service with defaults and CloudRunOptions.
func KfSystemWithDefaults(name string, kfo ...KfSystemOption) *v1alpha1.KfSystem {
	options := append([]KfSystemOption{WithDefaults}, kfo...)
	return KfSystem(name, options...)
}

// WithKfSystemFinalizer adds the Cloudrun finalizer to the Cloudrun.
func WithKfSystemFinalizer(r *v1alpha1.KfSystem) {
	r.ObjectMeta.Finalizers = append(r.ObjectMeta.Finalizers, "kfsystems.kf.dev")
}

// WithKfEnabled enables serving.
func WithKfEnabled(version string) KfSystemOption {
	return func(kfs *v1alpha1.KfSystem) {
		kfs.Spec.Kf.Enabled = ptr.Bool(true)
		kfs.Spec.Kf.Version = version
	}
}

// WithTargetKfVersion sets the target version of kf
func WithTargetKfVersion(version string) KfSystemOption {
	return func(kfs *v1alpha1.KfSystem) {
		kfs.Status.TargetKfVersion = version
	}
}

// WithDefaults initializes conditions of KfSystem.
func WithDefaults(kfs *v1alpha1.KfSystem) {
	kfs.Status.InitializeConditions()
}

// WithKfInstallSucceeded marks eventing as ready.
func WithKfInstallSucceeded(version string) KfSystemOption {
	return func(kfs *v1alpha1.KfSystem) {
		kfs.Status.MarkKfInstallSucceeded(version)
	}
}

// WithKfInstallFailed marks kf failed.
func WithKfInstallFailed(message string) KfSystemOption {
	return func(kfs *v1alpha1.KfSystem) {
		kfs.Status.MarkKfInstallFailed(message)
	}
}

// WithKfInstallNotReady marks serving as not ready.
func WithKfInstallNotReady(kfs *v1alpha1.KfSystem) {
	kfs.Status.MarkKfInstallNotReady()
}

// WithDeletionTimestamp adds a DeletionTimestamp, triggering the finalizer.
func WithDeletionTimestamp(t *metav1.Time) KfSystemOption {
	return func(kfs *v1alpha1.KfSystem) {
		kfs.DeletionTimestamp = t
	}
}

// WithControllerCACerts adds the certs secret name.
func WithControllerCACerts(name string) KfSystemOption {
	return func(kfs *v1alpha1.KfSystem) {
		kfs.Spec.Kf.Config.Secrets.ControllerCACerts.Name = name
	}
}
