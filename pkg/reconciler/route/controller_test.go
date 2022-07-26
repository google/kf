// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package route

import (
	"errors"
	"testing"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	networking "github.com/google/kf/v2/pkg/apis/networking/v1alpha3"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/reconciler/route/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
)

//go:generate mockgen --package=route --copyright_file ../../kf/internal/tools/option-builder/LICENSE_HEADER --destination=fake_reconciler.go --mock_names=Reconciler=FakeReconciler knative.dev/pkg/controller Reconciler

func TestBuildEnqueuer(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		obj          interface{}
		wantErr      error
		wantEnqueued []types.NamespacedName
	}{
		"app": {
			obj: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{Namespace: "some-namespace"},
				Status: v1alpha1.AppStatus{
					Routes: []v1alpha1.AppRouteStatus{
						{
							QualifiedRouteBinding: v1alpha1.QualifiedRouteBinding{
								Source: v1alpha1.RouteSpecFields{
									Domain: "some-domain1",
								},
							},
						},
						{
							QualifiedRouteBinding: v1alpha1.QualifiedRouteBinding{
								Source: v1alpha1.RouteSpecFields{
									Domain: "some-domain2",
								},
							},
						},
					},
				},
			},
			wantEnqueued: []types.NamespacedName{
				{Namespace: "some-namespace", Name: "some-domain1"},
				{Namespace: "some-namespace", Name: "some-domain2"},
			},
		},
		"route claim": {
			obj: &v1alpha1.Route{
				ObjectMeta: metav1.ObjectMeta{Namespace: "some-namespace"},
				Spec: v1alpha1.RouteSpec{
					RouteSpecFields: v1alpha1.RouteSpecFields{
						Domain: "some-domain",
					},
				},
			},
			wantEnqueued: []types.NamespacedName{
				{Namespace: "some-namespace", Name: "some-domain"},
			},
		},
		"unhandled type": {
			wantErr: errors.New("unexpected type: int"),
			obj:     99,
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			var gotEnqueued []types.NamespacedName

			f := BuildEnqueuer(func(obj types.NamespacedName) {
				gotEnqueued = append(gotEnqueued, obj)
			})
			err := f(tc.obj)

			testutil.AssertErrorsEqual(t, tc.wantErr, err)
			testutil.AssertEqual(t, "enqueued", tc.wantEnqueued, gotEnqueued)
		})
	}
}

func TestFilterVSManagedByKf(t *testing.T) {
	t.Parallel()

	buildVS := func(managedBy string) *networking.VirtualService {
		return &networking.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1alpha1.ManagedByLabel: managedBy,
				},
			},
		}
	}

	buildPod := func(managedBy string) *corev1.Pod {
		return &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1alpha1.ManagedByLabel: managedBy,
				},
			},
		}
	}

	f := FilterVSManagedByKf()
	testutil.AssertEqual(t, "non metav1.Object", false, f(99))
	testutil.AssertEqual(t, "wrong managed-by label", false, f(buildVS("not-kf")))
	testutil.AssertEqual(t, "no managed-by label", false, f(buildVS("")))
	testutil.AssertEqual(t, "correct label, wrong type", false, f(buildPod("kf")))
	testutil.AssertEqual(t, "correct everything", true, f(buildVS("kf")))
}

func TestEnqueueRoutesOfVirtualService(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		obj interface{}

		wantEnqueue bool
	}{
		"no domain specified": {
			obj: &networking.VirtualService{},
		},
		"has domain": {
			obj: (func() *networking.VirtualService {
				out := networking.VirtualService{}
				out.Annotations = map[string]string{
					resources.DomainAnnotation: "example.com",
				}
				return &out
			})(),
			wantEnqueue: true,
		},
		"handle non VirtualServices": {
			obj: 99,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {

			gotEnqueue := false
			enqueuer := func(_ interface{}) {
				gotEnqueue = true
			}
			f := EnqueueRoutesOfVirtualService(enqueuer)
			f(tc.obj)
			testutil.AssertEqual(t, "object enqueued", tc.wantEnqueue, gotEnqueue)
		})
	}
}

// TestEncodeDecodeKey ensures the behavior of cache.SplitMetaNamespaceKey
// and types.NamespacedName are compatible with domains.
func TestEncodeDecodeKey(t *testing.T) {
	t.Parallel()

	ns := "some-namespace"
	domain := "example.google.com"
	nn := types.NamespacedName{
		Namespace: ns,
		Name:      domain,
	}

	actualNs, actualDomain, err := cache.SplitMetaNamespaceKey(nn.String())
	testutil.AssertNil(t, "err", err)
	testutil.AssertEqual(t, "namespace", ns, actualNs)
	testutil.AssertEqual(t, "domain", domain, actualDomain)
}
