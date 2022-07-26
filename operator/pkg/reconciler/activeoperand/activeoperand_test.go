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

package activeoperand

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"kf-operator/pkg/apis/operand/v1alpha1"
	"kf-operator/pkg/operand/fake"
	crrtesting "kf-operator/pkg/reconciler/testing"

	. "kf-operator/pkg/testing/operand/v1alpha1"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/kmeta"
)

var (
	ns = "namespace"
	gk = schema.GroupKind{
		Group: "a",
		Kind:  "c",
	}
	liveref = CreateLiveRef("bcd", ns, gk)
)

func TestReconcileKind(t *testing.T) {
	ctx := context.TODO()
	tests := []struct {
		name            string
		before          *v1alpha1.ActiveOperand
		after           *v1alpha1.ActiveOperand
		err             error
		skipInjection   bool
		wantInvocations []fake.Invocation
	}{
		{
			name:   "inject owner reference successfully",
			before: ActiveOperandWithDefaults("blah", ns, WithLiveRefs(liveref)),
			after:  ActiveOperandWithDefaults("blah", ns, WithLiveRefs(liveref), WithOwnerRefsInjected()),
			wantInvocations: []fake.Invocation{{
				Live: []v1alpha1.LiveRef{liveref},
				Ctx:  ctx,
			}},
		}, {
			name:   "inject owner reference with error",
			before: ActiveOperandWithDefaults("blah", ns, WithLiveRefs(liveref)),
			after:  ActiveOperandWithDefaults("blah", ns, WithLiveRefs(liveref), WithOwnerRefsInjectedFailed("whoops")),
			err:    errors.New("whoops"),
			wantInvocations: []fake.Invocation{{
				Live: []v1alpha1.LiveRef{liveref},
				Ctx:  ctx,
			}},
		}, {
			name:          "no owner reference injection",
			before:        ActiveOperandWithDefaults("blah", ns, WithLiveRefs(liveref)),
			after:         ActiveOperandWithDefaults("blah", ns, WithLiveRefs(liveref), WithOwnerRefsInjected()),
			skipInjection: true,
			// when setting skipInjection to true, there will be no wantInvocations.
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf(test.name), func(t *testing.T) {
			fakeHandler := fake.CreateHandler()
			fakeHandler.SetError(test.err)
			fakeHandler.SetNoInjection(test.skipInjection)

			temp := &v1alpha1.ActiveOperand{}
			test.before.DeepCopyInto(temp)
			reconciler{
				OwnerHandler: fakeHandler,
				enqueueAfter: func(interface{}, time.Duration) {},
			}.ReconcileKind(ctx, temp)

			got := fakeHandler.GetInvocations()
			if len(got) != len(test.wantInvocations) {
				t.Fatalf("Wanted %d Invocations but saw %+v", len(test.wantInvocations), got)
			}

			for i, want := range test.wantInvocations {
				want.OwnerReference = kmeta.NewControllerRef(temp)
				if !cmp.Equal(want, got[i]) {
					t.Fatalf("Wanted interactions (-want +got)\n %s", cmp.Diff(want, got[i]))
				}
			}

			if !cmp.Equal(test.after, temp, crrtesting.IgnoreLastTransitionTime) {
				t.Fatalf("Wanted result (-want +got)\n %s", cmp.Diff(test.after, temp, crrtesting.IgnoreLastTransitionTime))

			}
		})
	}
}
