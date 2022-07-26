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

package fake

import (
	"context"
	"kf-operator/pkg/apis/operand/v1alpha1"
	"kf-operator/pkg/operand"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type fakeOwnerInjector struct {
	err         error
	invocations []Invocation
}

// Invocation contains OwnerRefence and LiveRef objects it injected into.
type Invocation struct {
	*metav1.OwnerReference
	Live []v1alpha1.LiveRef
	Ctx  context.Context
}

var _ operand.OwnerInjector = (OwnerInjector)(nil)

// OwnerInjector is able to inject the given Owner to an arbitrary
// set of objects referred to by LiveRefs.
type OwnerInjector interface {
	InjectOwnerRefs(context.Context, *metav1.OwnerReference, []v1alpha1.LiveRef) error
	SetError(error)
	GetInvocations() []Invocation
}

// Create creates a FakeOwnerInjector by extracting the dynamicclient
// from the context given.
func Create() OwnerInjector {
	return &fakeOwnerInjector{}
}

// InjectOwnerRefs injects the given OwnerRefence into all objects passed in as a LiveRef.
func (o *fakeOwnerInjector) InjectOwnerRefs(ctx context.Context, ownerRef *metav1.OwnerReference, live []v1alpha1.LiveRef) error {
	o.invocations = append(o.invocations, Invocation{OwnerReference: ownerRef, Live: live, Ctx: ctx})
	return o.err
}

// SetError makes this injector return an error.
func (o *fakeOwnerInjector) SetError(err error) {
	o.err = err
}

func (o *fakeOwnerInjector) GetInvocations() []Invocation {
	return o.invocations
}
