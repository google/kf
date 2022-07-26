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
	"fmt"
	"kf-operator/pkg/apis/operand/v1alpha1"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ns                       = "namespace"
	liveref v1alpha1.LiveRef = v1alpha1.LiveRef{
		Group:     "a",
		Version:   "b",
		Resource:  "c",
		Name:      "Blah",
		Namespace: ns,
	}
)

func TestFake(t *testing.T) {
	fake := Create()
	ctx := context.Background()

	ownerRef := &metav1.OwnerReference{
		Name: "blah",
	}
	err := fake.InjectOwnerRefs(ctx, ownerRef, []v1alpha1.LiveRef{liveref})

	if err != nil {
		t.Fatalf("wanted no error got %+v", err)
	}
	if len(fake.GetInvocations()) != 1 {
		t.Fatalf("Wanted 1 invocation got %+v", fake.GetInvocations())
	}
	want := Invocation{
		OwnerReference: ownerRef,
		Live:           []v1alpha1.LiveRef{liveref},
		Ctx:            ctx,
	}
	if !cmp.Equal(want, fake.GetInvocations()[0]) {
		t.Fatalf("Live ref differed (-want +got) \n %s", cmp.Diff(want, fake.GetInvocations()[0]))
	}

	fake.SetError(fmt.Errorf("Test %s", "test"))

	err = fake.InjectOwnerRefs(ctx, ownerRef, []v1alpha1.LiveRef{liveref})

	if want, got := fmt.Sprintf("%s", err), "Test test"; want != got {
		t.Fatalf("Wanted %s saw %s in error returned", want, got)
	}

	if len(fake.GetInvocations()) != 2 {
		t.Fatalf("Wanted 1 invocation got %+v", fake.GetInvocations())
	}
	if !cmp.Equal(want, fake.GetInvocations()[1]) {
		t.Fatalf("Live ref differed (-want +got) \n %s", cmp.Diff(want, fake.GetInvocations()[0]))
	}
}
