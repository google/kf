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
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	networking "knative.dev/pkg/apis/istio/v1alpha3"
)

//go:generate mockgen --package=route --copyright_file ../../kf/internal/tools/option-builder/LICENSE_HEADER --destination=fake_reconciler.go --mock_names=Reconciler=FakeReconciler knative.dev/pkg/controller Reconciler

func TestBuildEnqueuer(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		ExpectedErr error
		Obj         interface{}
		Enqueue     func(obj interface{})
	}{
		"route": {
			Obj: &v1alpha1.Route{
				ObjectMeta: metav1.ObjectMeta{Namespace: "some-namespace"},
				Spec: v1alpha1.RouteSpec{
					RouteSpecFields: v1alpha1.RouteSpecFields{
						Hostname: "some-hostname",
					},
				},
			},
			Enqueue: func(obj interface{}) {
				testutil.AssertJSONEqual(t, `{"namespace":"some-namespace", "hostname":"some-hostname"}`, string(obj.(cache.ExplicitKey)))
			},
		},
		"route claim": {
			Obj: &v1alpha1.RouteClaim{
				ObjectMeta: metav1.ObjectMeta{Namespace: "some-namespace"},
				Spec: v1alpha1.RouteClaimSpec{
					RouteSpecFields: v1alpha1.RouteSpecFields{
						Hostname: "some-hostname",
					},
				},
			},
			Enqueue: func(obj interface{}) {
				testutil.AssertJSONEqual(t, `{"namespace":"some-namespace", "hostname":"some-hostname"}`, string(obj.(cache.ExplicitKey)))
			},
		},
		"unhandled type": {
			ExpectedErr: errors.New("unexpected type: int"),
			Obj:         99,
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			f := BuildEnqueuer(tc.Enqueue)
			err := f(tc.Obj)

			testutil.AssertErrorsEqual(t, tc.ExpectedErr, err)
			if err != nil {
				return
			}
		})
	}
}

func TestFilterVSWithNamespace(t *testing.T) {
	t.Parallel()

	buildVS := func(namespace string) *networking.VirtualService {
		return &networking.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
			},
		}
	}

	buildPod := func(namespace string) *corev1.Pod {
		return &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
			},
		}
	}

	f := FilterVSWithNamespace("some-namespace")
	testutil.AssertEqual(t, "non metav1.Object", false, f(99))
	testutil.AssertEqual(t, "wrong namespace", false, f(buildVS("wrong-namespace")))
	testutil.AssertEqual(t, "correct namespace, wrong type", false, f(buildPod("some-namespace")))
	testutil.AssertEqual(t, "correct everything", true, f(buildVS("some-namespace")))
}

func TestEnqueueRoutesOfVirtualService(t *testing.T) {
	t.Parallel()

	buildVS := func() *networking.VirtualService {
		return &networking.VirtualService{}
	}

	testCases := map[string]struct {
		ExpectedErr   error
		Obj           interface{}
		BuildEnqueuer func(t *testing.T) func(interface{})
		Setup         func(t *testing.T, f *FakeRouteLister, fn *FakeRouteNamespaceLister)
	}{
		"enqueues each route": {
			Obj: buildVS(),
			Setup: func(t *testing.T, f *FakeRouteLister, fn *FakeRouteNamespaceLister) {
				f.EXPECT().
					Routes(gomock.Any()).
					Return(fn)

				fn.EXPECT().
					List(gomock.Any()).
					Return([]*v1alpha1.Route{
						{Spec: v1alpha1.RouteSpec{RouteSpecFields: v1alpha1.RouteSpecFields{Hostname: "host-1"}}},
						{Spec: v1alpha1.RouteSpec{RouteSpecFields: v1alpha1.RouteSpecFields{Hostname: "host-2"}}},
					}, nil)
			},
			BuildEnqueuer: func(t *testing.T) func(interface{}) {
				var i int
				return func(obj interface{}) {
					i++
					r := obj.(*v1alpha1.Route)
					testutil.AssertEqual(
						t,
						fmt.Sprintf("route-%d", i),
						fmt.Sprintf("host-%d", i),
						r.Spec.Hostname,
					)
				}
			},
		},
		"handle non VirtualServices": {
			Obj: 99,
		},
		"route lister fails": {
			Obj:         buildVS(),
			ExpectedErr: errors.New("failed to list corresponding routes: some-error"),
			Setup: func(t *testing.T, f *FakeRouteLister, fn *FakeRouteNamespaceLister) {
				f.EXPECT().
					Routes(gomock.Any()).
					Return(fn)

				fn.EXPECT().
					List(gomock.Any()).
					Return(nil, errors.New("some-error"))
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeRouteLister := NewFakeRouteLister(ctrl)
			fakeRouteNamespaceLister := NewFakeRouteNamespaceLister(ctrl)

			if tc.Setup != nil {
				tc.Setup(t, fakeRouteLister, fakeRouteNamespaceLister)
			}

			if tc.BuildEnqueuer == nil {
				tc.BuildEnqueuer = func(*testing.T) func(interface{}) {
					return func(interface{}) {}
				}
			}

			f := EnqueueRoutesOfVirtualService(tc.BuildEnqueuer(t), fakeRouteLister)
			err := f(tc.Obj)
			testutil.AssertErrorsEqual(t, tc.ExpectedErr, err)

			if err != nil {
				return
			}
			ctrl.Finish()
		})
	}
}
