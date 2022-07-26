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
	"k8s.io/apimachinery/pkg/runtime"
)

type fakeOwnerHandler struct {
	ownerInjector OwnerInjector
	err           error
	noInjection   bool
}

var _ operand.OwnerHandler = (OwnerHandler)(nil)

// OwnerHandler is able to handle owner references with given ClusterActiveOperand/ActiveOperand.
type OwnerHandler interface {
	HandleOwnerRefs(context.Context, *metav1.OwnerReference, []v1alpha1.LiveRef, runtime.Object) error
	SetError(error)
	GetInvocations() []Invocation
	SetNoInjection(noInjection bool)
}

// CreateHandler creates a FakeOwnerHandler.
func CreateHandler() OwnerHandler {
	return &fakeOwnerHandler{
		ownerInjector: Create(),
	}
}

// HandleOwnerRefs handles owner references with given ClusterActiveOperand/ActiveOperand.
func (o *fakeOwnerHandler) HandleOwnerRefs(ctx context.Context, ownerRef *metav1.OwnerReference, live []v1alpha1.LiveRef, obj runtime.Object) error {
	if !o.noInjection {
		return o.ownerInjector.InjectOwnerRefs(ctx, ownerRef, live)
	}
	return o.err
}

// SetError makes this injector return an error.
func (o *fakeOwnerHandler) SetError(err error) {
	o.ownerInjector.SetError(err)
	o.err = err
}

func (o *fakeOwnerHandler) GetInvocations() []Invocation {
	return o.ownerInjector.GetInvocations()
}

func (o *fakeOwnerHandler) SetNoInjection(noInjection bool) {
	o.noInjection = noInjection
}
