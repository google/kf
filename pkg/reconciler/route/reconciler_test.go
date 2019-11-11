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
	"context"
	"errors"
	"testing"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"

	gomock "github.com/golang/mock/gomock"
	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/testutil"
	"github.com/google/kf/pkg/reconciler"
	appresources "github.com/google/kf/pkg/reconciler/app/resources"
	"github.com/google/kf/pkg/reconciler/route/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logtesting "knative.dev/pkg/logging/testing"
)

//go:generate mockgen --package=route --copyright_file ../../kf/internal/tools/option-builder/LICENSE_HEADER --destination=fake_listers.go --mock_names=RouteLister=FakeRouteLister,RouteNamespaceLister=FakeRouteNamespaceLister,RouteClaimLister=FakeRouteClaimLister,RouteClaimNamespaceLister=FakeRouteClaimNamespaceLister github.com/google/kf/pkg/client/listers/kf/v1alpha1 RouteLister,RouteClaimLister,RouteNamespaceLister,RouteClaimNamespaceLister
//go:generate mockgen --package=route --copyright_file ../../kf/internal/tools/option-builder/LICENSE_HEADER --destination=fake_corev1_listers.go --mock_names=NamespaceLister=FakeNamespaceLister k8s.io/client-go/listers/core/v1 NamespaceLister
//go:generate mockgen --package=route --copyright_file ../../kf/internal/tools/option-builder/LICENSE_HEADER --destination=fake_shared_client.go --mock_names=Interface=FakeSharedClient knative.dev/pkg/client/clientset/versioned Interface
//go:generate mockgen --package=route --copyright_file ../../kf/internal/tools/option-builder/LICENSE_HEADER --destination=fake_networking.go --mock_names=NetworkingV1alpha3Interface=FakeNetworking,VirtualServiceInterface=FakeVirtualServiceInterface knative.dev/pkg/client/clientset/versioned/typed/istio/v1alpha3 NetworkingV1alpha3Interface,VirtualServiceInterface
//go:generate mockgen --package=route --copyright_file ../../kf/internal/tools/option-builder/LICENSE_HEADER --destination=fake_kf.go --mock_names=Interface=FakeKfInterface github.com/google/kf/pkg/client/clientset/versioned Interface
//go:generate mockgen --package=route --copyright_file ../../kf/internal/tools/option-builder/LICENSE_HEADER --destination=fake_kf_v1alpha1.go --mock_names=KfV1alpha1Interface=FakeKfAlpha1Interface,RouteInterface=FakeRouteInterface github.com/google/kf/pkg/client/clientset/versioned/typed/kf/v1alpha1 KfV1alpha1Interface,RouteInterface
//go:generate mockgen --package=route --copyright_file ../../kf/internal/tools/option-builder/LICENSE_HEADER --destination=fake_istio_listers.go --mock_names=VirtualServiceLister=FakeVirtualServiceLister,VirtualServiceNamespaceLister=FakeVirtualServiceNamespaceLister knative.dev/pkg/client/listers/istio/v1alpha3 VirtualServiceLister,VirtualServiceNamespaceLister
func TestReconciler_Reconcile_badKey(t *testing.T) {
	t.Parallel()

	r := &Reconciler{}
	err := r.Reconcile(context.Background(), "i/n/v/a/l/i/d")
	testutil.AssertErrorsEqual(t, errors.New(`invalid character 'i' looking for beginning of value`), err)
}

func TestReconciler_Reconcile_namespaceIsTerminating(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	fakeNamespaceLister := NewFakeNamespaceLister(ctrl)
	fakeNamespaceLister.EXPECT().
		Get("some-namespace").
		Return(&corev1.Namespace{
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceTerminating,
			},
		}, nil)

	r := &Reconciler{
		Base: &reconciler.Base{
			NamespaceLister: fakeNamespaceLister,
		},
		configStore: config.NewDefaultConfigStore(logtesting.TestLogger(t)),
	}

	testutil.AssertNil(t, "err", r.Reconcile(context.Background(), `{"namespace":"some-namespace"}`))
}

func TestReconciler_Reconcile_ApplyChanges(t *testing.T) {
	t.Parallel()

	type fakes struct {
		frcnl *FakeRouteClaimNamespaceLister
		fn    *FakeNetworking
		fvsi  *FakeVirtualServiceInterface
		fkfi  *FakeKfInterface
		fkfai *FakeKfAlpha1Interface
		fri   *FakeRouteInterface
		frl   *FakeRouteLister
		frnl  *FakeRouteNamespaceLister
		fvsl  *FakeVirtualServiceLister
		fvsnl *FakeVirtualServiceNamespaceLister
	}

	testCases := map[string]struct {
		ExpectedErr     error
		Setup           func(t *testing.T, f fakes)
		RouteSpecFields v1alpha1.RouteSpecFields
		Namespace       string
	}{
		"fetching route claims fails": {
			ExpectedErr: errors.New("some-error"),
			Setup: func(t *testing.T, f fakes) {
				f.frcnl.EXPECT().
					List(gomock.Any()).
					Return(nil, errors.New("some-error"))
			},
		},
		"no claims, deleting VirtualServices fails": {
			ExpectedErr: errors.New("some-error"),
			Setup: func(t *testing.T, f fakes) {
				f.frcnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fn.EXPECT().
					VirtualServices(gomock.Any()).
					Return(f.fvsi)

				f.fvsi.EXPECT().
					Delete(gomock.Any(), gomock.Any()).
					Return(errors.New("some-error"))
			},
		},
		"no claims, deleting Routes fails": {
			ExpectedErr: errors.New("some-error"),
			Setup: func(t *testing.T, f fakes) {
				f.frcnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fn.EXPECT().
					VirtualServices(gomock.Any()).
					Return(f.fvsi)

				f.fvsi.EXPECT().
					Delete(gomock.Any(), gomock.Any()).
					Return(nil)

				f.fkfi.EXPECT().
					Kf().
					Return(f.fkfai)

				f.fkfai.EXPECT().
					Routes(gomock.Any()).
					Return(f.fri)

				f.fri.EXPECT().
					DeleteCollection(gomock.Any(), gomock.Any()).
					Return(errors.New("some-error"))
			},
		},
		"no claims, deleting VirtualServices returns not found": {
			ExpectedErr: nil, // No need to return an error
			Setup: func(t *testing.T, f fakes) {
				f.frcnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fn.EXPECT().
					VirtualServices(gomock.Any()).
					Return(f.fvsi)

				f.fvsi.EXPECT().
					Delete(gomock.Any(), gomock.Any()).
					Return(apierrors.NewNotFound(v1alpha3.Resource("VirtualService"), "VirtualService"))

				f.fkfi.EXPECT().
					Kf().
					Return(f.fkfai)

				f.fkfai.EXPECT().
					Routes(gomock.Any()).
					Return(f.fri)

				f.fri.EXPECT().
					DeleteCollection(gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		"no claims, deleting Routes returns not found": {
			ExpectedErr: nil, // No need to return an error
			Setup: func(t *testing.T, f fakes) {
				f.frcnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fn.EXPECT().
					VirtualServices(gomock.Any()).
					Return(f.fvsi)

				f.fvsi.EXPECT().
					Delete(gomock.Any(), gomock.Any()).
					Return(nil)

				f.fkfi.EXPECT().
					Kf().
					Return(f.fkfai)

				f.fkfai.EXPECT().
					Routes(gomock.Any()).
					Return(f.fri)

				f.fri.EXPECT().
					DeleteCollection(gomock.Any(), gomock.Any()).
					Return(apierrors.NewNotFound(v1alpha3.Resource("Route"), "Route"))
			},
		},
		"no claims, deletes Routes and VirtualServices": {
			Namespace: "some-namespace",
			RouteSpecFields: v1alpha1.RouteSpecFields{
				Hostname: "some-hostname",
				Domain:   "example.com",
			},
			Setup: func(t *testing.T, f fakes) {
				f.frcnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fn.EXPECT().
					VirtualServices("some-namespace").
					Return(f.fvsi)

				f.fvsi.EXPECT().
					Delete(
						v1alpha1.GenerateName(
							"some-hostname",
							"example.com",
						),
						&metav1.DeleteOptions{},
					).
					Return(nil)

				f.fkfi.EXPECT().
					Kf().
					Return(f.fkfai)

				f.fkfai.EXPECT().
					Routes("some-namespace").
					Return(f.fri)

				f.fri.EXPECT().
					DeleteCollection(
						&metav1.DeleteOptions{},
						metav1.ListOptions{
							LabelSelector: appresources.MakeRouteSelectorNoPath(v1alpha1.RouteSpecFields{
								Hostname: "some-hostname",
								Domain:   "example.com",
							}).String(),
						},
					).
					Return(nil)
			},
		},
		"listing routes fails": {
			ExpectedErr: errors.New("some-error"),
			Setup: func(t *testing.T, f fakes) {
				f.frcnl.EXPECT().
					List(gomock.Any()).
					Return([]*v1alpha1.RouteClaim{
						{},
					}, nil)

				f.frl.EXPECT().
					Routes(gomock.Any()).
					Return(f.frnl)

				f.frnl.EXPECT().
					List(gomock.Any()).
					Return(nil, errors.New("some-error"))
			},
		},
		"making VirtualServices fails": {
			ExpectedErr: errors.New(`failed to convert path to regexp: mux: unbalanced braces in "/}invalid{"`),
			Setup: func(t *testing.T, f fakes) {
				f.frcnl.EXPECT().
					List(gomock.Any()).
					Return([]*v1alpha1.RouteClaim{
						{Spec: v1alpha1.RouteClaimSpec{RouteSpecFields: v1alpha1.RouteSpecFields{Path: "}invalid{"}}},
					}, nil)

				f.frl.EXPECT().
					Routes(gomock.Any()).
					Return(f.frnl)

				f.frnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)
			},
		},
		"getting VirtualServices fails": {
			ExpectedErr: errors.New("some-error"),
			Setup: func(t *testing.T, f fakes) {
				f.frcnl.EXPECT().
					List(gomock.Any()).
					Return([]*v1alpha1.RouteClaim{
						{Spec: v1alpha1.RouteClaimSpec{RouteSpecFields: v1alpha1.RouteSpecFields{}}},
					}, nil)

				f.frl.EXPECT().
					Routes(gomock.Any()).
					Return(f.frnl)

				f.frnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsl.EXPECT().
					VirtualServices(gomock.Any()).
					Return(f.fvsnl)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(nil, errors.New("some-error"))
			},
		},
		"VirtualServices is not found, creating VirtualService fails": {
			ExpectedErr: errors.New("some-error"),
			Setup: func(t *testing.T, f fakes) {
				f.frcnl.EXPECT().
					List(gomock.Any()).
					Return([]*v1alpha1.RouteClaim{
						{Spec: v1alpha1.RouteClaimSpec{RouteSpecFields: v1alpha1.RouteSpecFields{}}},
					}, nil)

				f.frl.EXPECT().
					Routes(gomock.Any()).
					Return(f.frnl)

				f.frnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsl.EXPECT().
					VirtualServices(gomock.Any()).
					Return(f.fvsnl)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(nil, apierrors.NewNotFound(v1alpha3.Resource("VirtualService"), "VirtualService"))

				f.fn.EXPECT().
					VirtualServices(gomock.Any()).
					Return(f.fvsi)

				f.fvsi.EXPECT().
					Create(gomock.Any()).
					Return(nil, errors.New("some-error"))
			},
		},
		"VirtualServices is not found, creating VirtualService": {
			Setup: func(t *testing.T, f fakes) {
				f.frcnl.EXPECT().
					List(gomock.Any()).
					Return([]*v1alpha1.RouteClaim{
						{Spec: v1alpha1.RouteClaimSpec{RouteSpecFields: v1alpha1.RouteSpecFields{}}},
					}, nil)

				f.frl.EXPECT().
					Routes(gomock.Any()).
					Return(f.frnl)

				f.frnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsl.EXPECT().
					VirtualServices(gomock.Any()).
					Return(f.fvsnl)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(nil, apierrors.NewNotFound(v1alpha3.Resource("VirtualService"), "VirtualService"))

				f.fn.EXPECT().
					VirtualServices(gomock.Any()).
					Return(f.fvsi)

				f.fvsi.EXPECT().
					Create(gomock.Any()).
					Return(nil, nil)
			},
		},
		"VirtualServices is being deleted": {
			Setup: func(t *testing.T, f fakes) {
				f.frcnl.EXPECT().
					List(gomock.Any()).
					Return([]*v1alpha1.RouteClaim{
						{Spec: v1alpha1.RouteClaimSpec{RouteSpecFields: v1alpha1.RouteSpecFields{}}},
					}, nil)

				f.frl.EXPECT().
					Routes(gomock.Any()).
					Return(f.frnl)

				f.frnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsl.EXPECT().
					VirtualServices(gomock.Any()).
					Return(f.fvsnl)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(&v1alpha3.VirtualService{
						ObjectMeta: metav1.ObjectMeta{
							DeletionTimestamp: &metav1.Time{Time: time.Now()},
						},
					}, nil)
			},
		},
		"update VirtualServices fails": {
			ExpectedErr: errors.New("some-error"),
			Setup: func(t *testing.T, f fakes) {
				f.frcnl.EXPECT().
					List(gomock.Any()).
					Return([]*v1alpha1.RouteClaim{
						{Spec: v1alpha1.RouteClaimSpec{RouteSpecFields: v1alpha1.RouteSpecFields{}}},
					}, nil)

				f.frl.EXPECT().
					Routes(gomock.Any()).
					Return(f.frnl)

				f.frnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsl.EXPECT().
					VirtualServices(gomock.Any()).
					Return(f.fvsnl)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(&v1alpha3.VirtualService{}, nil)

				f.fn.EXPECT().
					VirtualServices(gomock.Any()).
					Return(f.fvsi)

				f.fvsi.EXPECT().
					Update(gomock.Any()).
					Return(nil, errors.New("some-error"))
			},
		},
		"update VirtualServices": {
			Namespace: "some-namespace",
			Setup: func(t *testing.T, f fakes) {
				f.frcnl.EXPECT().
					List(gomock.Any()).
					Return([]*v1alpha1.RouteClaim{
						{Spec: v1alpha1.RouteClaimSpec{RouteSpecFields: v1alpha1.RouteSpecFields{}}},
					}, nil)

				f.frl.EXPECT().
					Routes(gomock.Any()).
					Return(f.frnl)

				f.frnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsl.EXPECT().
					VirtualServices(gomock.Any()).
					Return(f.fvsnl)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(&v1alpha3.VirtualService{
						ObjectMeta: metav1.ObjectMeta{
							OwnerReferences: []metav1.OwnerReference{{Name: "some-name"}},
						},
					}, nil)

				f.fn.EXPECT().
					VirtualServices(gomock.Any()).
					Return(f.fvsi)

				f.fvsi.EXPECT().
					Update(gomock.Any()).
					Do(func(vs *v1alpha3.VirtualService) {
						testutil.AssertEqual(t, "OwnerReference len", 1, len(vs.OwnerReferences))
						testutil.AssertEqual(t, "HTTPRoutes len", 1, len(vs.Spec.HTTP))
					})
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeRouteClaimLister := NewFakeRouteClaimLister(ctrl)
			fakeRouteClaimNamespaceLister := NewFakeRouteClaimNamespaceLister(ctrl)

			fakeSharedClient := NewFakeSharedClient(ctrl)
			fakeNetworking := NewFakeNetworking(ctrl)
			fakeVirtualServiceInterface := NewFakeVirtualServiceInterface(ctrl)
			fakeKfInterface := NewFakeKfInterface(ctrl)
			fakeKfAlpha1Interface := NewFakeKfAlpha1Interface(ctrl)
			fakeRouteInterface := NewFakeRouteInterface(ctrl)
			fakeRouteLister := NewFakeRouteLister(ctrl)
			fakeRouteNamespaceLister := NewFakeRouteNamespaceLister(ctrl)
			fakeVirtualServiceLister := NewFakeVirtualServiceLister(ctrl)
			fakeVirtualServiceNamespaceLister := NewFakeVirtualServiceNamespaceLister(ctrl)

			fakeSharedClient.EXPECT().
				Networking().
				Return(fakeNetworking)

			fakeRouteClaimLister.EXPECT().
				RouteClaims(gomock.Any()).
				Return(fakeRouteClaimNamespaceLister)

			if tc.Setup != nil {
				tc.Setup(t, fakes{
					frcnl: fakeRouteClaimNamespaceLister,
					fn:    fakeNetworking,
					fvsi:  fakeVirtualServiceInterface,
					fkfi:  fakeKfInterface,
					fkfai: fakeKfAlpha1Interface,
					fri:   fakeRouteInterface,
					frl:   fakeRouteLister,
					frnl:  fakeRouteNamespaceLister,
					fvsl:  fakeVirtualServiceLister,
					fvsnl: fakeVirtualServiceNamespaceLister,
				})
			}

			r := &Reconciler{
				Base: &reconciler.Base{
					SharedClientSet: fakeSharedClient,
					KfClientSet:     fakeKfInterface,
				},
				routeClaimLister:     fakeRouteClaimLister,
				routeLister:          fakeRouteLister,
				virtualServiceLister: fakeVirtualServiceLister,
				configStore:          config.NewDefaultConfigStore(logtesting.TestLogger(t)),
			}

			ctx := r.configStore.ToContext(context.Background())

			err := r.ApplyChanges(
				ctx,
				tc.Namespace,
				tc.RouteSpecFields,
			)

			testutil.AssertErrorsEqual(t, tc.ExpectedErr, err)
		})
	}
}
