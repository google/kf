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

	gomock "github.com/golang/mock/gomock"
	"github.com/google/kf/v2/pkg/apis/kf/config"
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/apis/networking/v1alpha3"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/reconciler"
	istio "istio.io/api/networking/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	pkgreconciler "knative.dev/pkg/reconciler"
)

//go:generate mockgen --package=route --copyright_file ../../kf/internal/tools/option-builder/LICENSE_HEADER --destination=fake_listers.go --mock_names=RouteLister=FakeRouteLister,RouteNamespaceLister=FakeRouteNamespaceLister,AppLister=FakeAppLister,AppNamespaceLister=FakeAppNamespaceLister,SpaceLister=FakeSpaceLister,ServiceInstanceBindingLister=FakeServiceInstanceBindingLister,ServiceInstanceBindingNamespaceLister=FakeServiceInstanceBindingNamespaceLister github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1 RouteLister,RouteNamespaceLister,AppLister,AppNamespaceLister,SpaceLister,ServiceInstanceBindingLister,ServiceInstanceBindingNamespaceLister
//go:generate mockgen --package=route --copyright_file ../../kf/internal/tools/option-builder/LICENSE_HEADER --destination=fake_corev1_listers.go --mock_names=NamespaceLister=FakeNamespaceLister k8s.io/client-go/listers/core/v1 NamespaceLister
//go:generate mockgen --package=route --copyright_file ../../kf/internal/tools/option-builder/LICENSE_HEADER --destination=fake_networking_client.go --mock_names=Interface=FakeNetworkingClient github.com/google/kf/v2/pkg/client/networking/clientset/versioned Interface
//go:generate mockgen --package=route --copyright_file ../../kf/internal/tools/option-builder/LICENSE_HEADER --destination=fake_networking.go --mock_names=NetworkingV1alpha3Interface=FakeNetworking,VirtualServiceInterface=FakeVirtualServiceInterface github.com/google/kf/v2/pkg/client/networking/clientset/versioned/typed/networking/v1alpha3 NetworkingV1alpha3Interface,VirtualServiceInterface
//go:generate mockgen --package=route --copyright_file ../../kf/internal/tools/option-builder/LICENSE_HEADER --destination=fake_kf.go --mock_names=Interface=FakeKfInterface github.com/google/kf/v2/pkg/client/kf/clientset/versioned Interface
//go:generate mockgen --package=route --copyright_file ../../kf/internal/tools/option-builder/LICENSE_HEADER --destination=fake_kf_v1alpha1.go --mock_names=KfV1alpha1Interface=FakeKfAlpha1Interface,RouteInterface=FakeRouteInterface github.com/google/kf/v2/pkg/client/kf/clientset/versioned/typed/kf/v1alpha1 KfV1alpha1Interface,RouteInterface
//go:generate mockgen --package=route --copyright_file ../../kf/internal/tools/option-builder/LICENSE_HEADER --destination=fake_networking_listers.go --mock_names=VirtualServiceLister=FakeVirtualServiceLister,VirtualServiceNamespaceLister=FakeVirtualServiceNamespaceLister github.com/google/kf/v2/pkg/client/networking/listers/networking/v1alpha3 VirtualServiceLister,VirtualServiceNamespaceLister

type testConfigStore struct {
	config *config.DefaultsConfig
}

func (t *testConfigStore) ToContext(ctx context.Context) context.Context {
	return config.ToContextForTest(ctx, config.CreateConfigForTest(t.config))
}

var _ pkgreconciler.ConfigStore = (*testConfigStore)(nil)

func TestReconciler_Reconcile_badKey(t *testing.T) {
	t.Parallel()

	r := &Reconciler{}
	err := r.Reconcile(context.Background(), "i/n/v/a/l/i/d")
	testutil.AssertErrorsEqual(t, errors.New(`unexpected key format: "i/n/v/a/l/i/d"`), err)
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
	}

	nn := types.NamespacedName{
		Namespace: "some-namespace",
		Name:      "some-domain",
	}

	testutil.AssertNil(t, "err", r.Reconcile(context.Background(), nn.String()))
}

func TestReconciler_Reconcile_ApplyChanges(t *testing.T) {
	t.Parallel()

	type fakes struct {
		fsl    *FakeSpaceLister
		fri    *FakeRouteInterface
		frl    *FakeRouteLister
		frnl   *FakeRouteNamespaceLister
		fvsi   *FakeVirtualServiceInterface
		fanl   *FakeAppNamespaceLister
		fvsnl  *FakeVirtualServiceNamespaceLister
		fsibl  *FakeServiceInstanceBindingLister
		fsibnl *FakeServiceInstanceBindingNamespaceLister
	}

	expectRouteListCall := func(frl *FakeRouteLister, frnl *FakeRouteNamespaceLister) {
		frl.EXPECT().
			Routes(gomock.Any()).
			Return(frnl)
	}

	goodDomain := "example.com"

	goodSpace := v1alpha1.Space{}
	goodSpace.Status.NetworkConfig.Domains = []v1alpha1.SpaceDomain{
		{Domain: goodDomain},
	}
	expectGoodSpace := func(fsl *FakeSpaceLister) {
		fsl.EXPECT().
			Get(gomock.Any()).
			Return(&goodSpace, nil).
			AnyTimes()
	}

	goodRoute := v1alpha1.Route{
		Spec: v1alpha1.RouteSpec{
			RouteSpecFields: v1alpha1.RouteSpecFields{
				Domain: goodDomain,
			},
		},
	}

	expectGoodRoute := func(frnl *FakeRouteNamespaceLister) {
		frnl.EXPECT().
			List(gomock.Any()).
			Return([]*v1alpha1.Route{
				&goodRoute,
			}, nil)
	}

	testCases := map[string]struct {
		ExpectedErr error
		Setup       func(t *testing.T, f fakes)
		Domain      string
		Namespace   string
	}{
		"fetching space fails": {
			ExpectedErr: errors.New("some-error"),
			Setup: func(t *testing.T, f fakes) {
				f.fsl.EXPECT().
					Get(gomock.Any()).
					Return(nil, errors.New("some-error"))
			},
		},
		"fetching routes fails": {
			Domain:      goodDomain,
			ExpectedErr: errors.New("some-error"),
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				f.frnl.EXPECT().
					List(gomock.Any()).
					Return(nil, errors.New("some-error"))
			},
		},
		"no routes deletes VirtualService": {
			Domain: goodDomain,
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				f.frnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsi.EXPECT().
					Delete(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"listing apps fails": {
			Domain:      goodDomain,
			ExpectedErr: errors.New("some-error"),
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				expectGoodRoute(f.frnl)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, errors.New("some-error"))
			},
		},
		"fetching service instance bindings fails": {
			Domain:      goodDomain,
			ExpectedErr: errors.New("some-error"),
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				expectGoodRoute(f.frnl)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, errors.New("some-error"))
			},
		},
		"making VirtualServices fails": {
			ExpectedErr: errors.New(`Error occurred while reconciling VirtualService: configuring: failed to convert path to regexp: mux: unbalanced braces in "/}invalid{"`),
			Domain:      goodDomain,
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				f.frnl.EXPECT().
					List(gomock.Any()).
					Return([]*v1alpha1.Route{
						{Spec: v1alpha1.RouteSpec{RouteSpecFields: v1alpha1.RouteSpecFields{Domain: goodDomain, Path: "}invalid{"}}},
					}, nil)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fri.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, c *v1alpha1.Route, o metav1.UpdateOptions) error {
					cond := c.Status.GetCondition(v1alpha1.RouteConditionReady)

					testutil.AssertTrue(t, "condition", cond.IsFalse())
					testutil.AssertEqual(t, "reason", "ReconciliationError", cond.Reason)
					testutil.AssertEqual(t, "message", `Error occurred while reconciling VirtualService: configuring: failed to convert path to regexp: mux: unbalanced braces in "/}invalid{"`, cond.Message)

					return nil
				})
			},
		},
		"getting VirtualServices fails": {
			Domain:      goodDomain,
			ExpectedErr: errors.New("Error occurred while reconciling VirtualService: some-error"),
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				expectGoodRoute(f.frnl)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(nil, errors.New("some-error"))

				f.fri.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"VirtualServices is not found, creating VirtualService fails": {
			Domain:      goodDomain,
			ExpectedErr: errors.New("Error occurred while reconciling VirtualService: some-error"),
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				expectGoodRoute(f.frnl)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(nil, apierrors.NewNotFound(v1alpha3.Resource("VirtualService"), "VirtualService"))

				f.fvsi.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("some-error"))

				f.fri.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"VirtualServices is not found, creating VirtualService": {
			Domain: goodDomain,
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				expectGoodRoute(f.frnl)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(nil, apierrors.NewNotFound(v1alpha3.Resource("VirtualService"), "VirtualService"))

				f.fvsi.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil)

				f.fri.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"VirtualServices is being deleted": {
			Domain: goodDomain,
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				expectGoodRoute(f.frnl)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(&v1alpha3.VirtualService{
						ObjectMeta: metav1.ObjectMeta{
							DeletionTimestamp: &metav1.Time{Time: time.Now()},
						},
					}, nil)

				f.fri.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"update VirtualServices fails": {
			Domain:      goodDomain,
			ExpectedErr: errors.New("Error occurred while reconciling VirtualService: some-error"),
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				expectGoodRoute(f.frnl)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(&v1alpha3.VirtualService{}, nil)

				f.fvsi.EXPECT().
					Update(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("some-error"))

				f.fri.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, c *v1alpha1.Route, o metav1.UpdateOptions) error {
					testutil.AssertTrue(t, "condition", c.Status.GetCondition(v1alpha1.RouteConditionReady).IsFalse())
					return nil
				})
			},
		},
		"update VirtualServices fails 409": {
			Domain:      goodDomain,
			ExpectedErr: errors.New(`Error occurred while reconciling VirtualService: Operation cannot be fulfilled on virtualservices.istio.io "MyService": the object has been modified; please apply your changes to the latest version and try again`),
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				expectGoodRoute(f.frnl)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(&v1alpha3.VirtualService{}, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				conflict := apierrors.NewConflict(
					schema.GroupResource{
						Group:    "istio.io",
						Resource: "virtualservices",
					},
					"MyService",
					errors.New("the object has been modified; please apply your changes to the latest version and try again"),
				)
				f.fvsi.EXPECT().
					Update(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, conflict)

				f.fri.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, c *v1alpha1.Route, o metav1.UpdateOptions) error {
					// The 409 should be passed through to cause an unknown status on the route
					// rather than a failure because it is a retryable error.
					testutil.AssertTrue(t, "condition", c.Status.GetCondition(v1alpha1.RouteConditionReady).IsUnknown())
					return nil
				})
			},
		},
		"update VirtualServices": {
			Domain:    goodDomain,
			Namespace: "some-namespace",
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				expectGoodRoute(f.frnl)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(&v1alpha3.VirtualService{
						ObjectMeta: metav1.ObjectMeta{
							OwnerReferences: []metav1.OwnerReference{{Name: "some-name"}},
						},
					}, nil)

				f.fvsi.EXPECT().
					Update(gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, vs *v1alpha3.VirtualService, o metav1.UpdateOptions) {
						testutil.AssertEqual(t, "OwnerReference len", 1, len(vs.OwnerReferences))
						testutil.AssertEqual(t, "HTTPRoutes len", 1, len(vs.Spec.Http))
					})

				f.fri.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"update hosts VirtualServices": {
			Domain:    goodDomain,
			Namespace: "some-namespace",
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				expectGoodRoute(f.frnl)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(&v1alpha3.VirtualService{
						ObjectMeta: metav1.ObjectMeta{
							OwnerReferences: []metav1.OwnerReference{{Name: "some-name"}},
						},
						Spec: istio.VirtualService{
							Hosts: []string{goodDomain},
						},
					}, nil)

				f.fvsi.EXPECT().
					Update(gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, vs *v1alpha3.VirtualService, o metav1.UpdateOptions) {
						testutil.AssertEqual(t, "OwnerReference len", 1, len(vs.OwnerReferences))
						testutil.AssertEqual(t, "Hosts len", 2, len(vs.Spec.Hosts))
					})

				f.fri.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"invalid space domain deletes VS": {
			Domain: goodDomain,
			Setup: func(t *testing.T, f fakes) {
				f.fsl.EXPECT().
					Get(gomock.Any()).
					Return(&v1alpha1.Space{}, nil) // domain not on space

				expectRouteListCall(f.frl, f.frnl)
				expectGoodRoute(f.frnl)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsi.EXPECT().
					Delete(gomock.Any(), gomock.Any(), gomock.Any())

				f.fri.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"routes reflect invalid domain": {
			Domain: goodDomain,
			Setup: func(t *testing.T, f fakes) {
				f.fsl.EXPECT().
					Get(gomock.Any()).
					Return(&v1alpha1.Space{}, nil) // domain not on space

				expectRouteListCall(f.frl, f.frnl)
				expectGoodRoute(f.frnl)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsi.EXPECT().
					Delete(gomock.Any(), gomock.Any(), gomock.Any())

				f.fri.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, c *v1alpha1.Route, o metav1.UpdateOptions) error {
					// The Route should fail due to invalid domain
					testutil.AssertTrue(t, "condition", c.Status.GetCondition(v1alpha1.RouteConditionReady).IsFalse())
					return nil
				})
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeRouteLister := NewFakeRouteLister(ctrl)
			fakeRouteNamespaceLister := NewFakeRouteNamespaceLister(ctrl)
			fakeNetworkingClient := NewFakeNetworkingClient(ctrl)
			fakeNetworking := NewFakeNetworking(ctrl)
			fakeVirtualServiceInterface := NewFakeVirtualServiceInterface(ctrl)
			fakeKfInterface := NewFakeKfInterface(ctrl)
			fakeKfAlpha1Interface := NewFakeKfAlpha1Interface(ctrl)
			fakeRouteInterface := NewFakeRouteInterface(ctrl)
			fakeAppLister := NewFakeAppLister(ctrl)
			fakeSpaceLister := NewFakeSpaceLister(ctrl)
			fakeAppNamespaceLister := NewFakeAppNamespaceLister(ctrl)
			fakeVirtualServiceLister := NewFakeVirtualServiceLister(ctrl)
			fakeVirtualServiceNamespaceLister := NewFakeVirtualServiceNamespaceLister(ctrl)
			fakeServiceInstanceBindingLister := NewFakeServiceInstanceBindingLister(ctrl)
			fakeServiceInstanceBindingNamespaceLister := NewFakeServiceInstanceBindingNamespaceLister(ctrl)

			fakeVirtualServiceLister.EXPECT().
				VirtualServices(gomock.Any()).
				Return(fakeVirtualServiceNamespaceLister).
				AnyTimes()

			fakeKfInterface.EXPECT().
				KfV1alpha1().
				Return(fakeKfAlpha1Interface).
				AnyTimes()

			fakeNetworkingClient.EXPECT().
				NetworkingV1alpha3().
				Return(fakeNetworking).
				AnyTimes()

			fakeNetworking.EXPECT().
				VirtualServices(gomock.Any()).
				Return(fakeVirtualServiceInterface).
				AnyTimes()

			fakeKfAlpha1Interface.EXPECT().
				Routes(gomock.Any()).
				Return(fakeRouteInterface).
				AnyTimes()

			fakeAppLister.EXPECT().
				Apps(gomock.Any()).
				Return(fakeAppNamespaceLister).
				AnyTimes()

			fakeServiceInstanceBindingLister.EXPECT().
				ServiceInstanceBindings(gomock.Any()).
				Return(fakeServiceInstanceBindingNamespaceLister).
				AnyTimes()

			if tc.Setup != nil {
				tc.Setup(t, fakes{
					fri:    fakeRouteInterface,
					frl:    fakeRouteLister,
					frnl:   fakeRouteNamespaceLister,
					fvsi:   fakeVirtualServiceInterface,
					fvsnl:  fakeVirtualServiceNamespaceLister,
					fanl:   fakeAppNamespaceLister,
					fsl:    fakeSpaceLister,
					fsibl:  fakeServiceInstanceBindingLister,
					fsibnl: fakeServiceInstanceBindingNamespaceLister,
				})
			}

			r := &Reconciler{
				Base: &reconciler.Base{
					KfClientSet: fakeKfInterface,
				},
				networkingClientSet:          fakeNetworkingClient,
				routeLister:                  fakeRouteLister,
				virtualServiceLister:         fakeVirtualServiceLister,
				appLister:                    fakeAppLister,
				spaceLister:                  fakeSpaceLister,
				serviceInstanceBindingLister: fakeServiceInstanceBindingLister,
				kfConfigStore: &testConfigStore{&config.DefaultsConfig{
					RouteTrackVirtualService: true,
				}},
			}

			err := r.ApplyChanges(
				context.Background(),
				tc.Namespace,
				tc.Domain,
			)

			testutil.AssertErrorsEqual(t, tc.ExpectedErr, err)

		})
	}
}

func TestReconciler_Reconcile_ApplyChanges_NotTrackingVirtualService(t *testing.T) {
	t.Parallel()

	type fakes struct {
		fsl    *FakeSpaceLister
		fri    *FakeRouteInterface
		frl    *FakeRouteLister
		frnl   *FakeRouteNamespaceLister
		fvsi   *FakeVirtualServiceInterface
		fanl   *FakeAppNamespaceLister
		fvsnl  *FakeVirtualServiceNamespaceLister
		fsibl  *FakeServiceInstanceBindingLister
		fsibnl *FakeServiceInstanceBindingNamespaceLister
	}

	expectRouteListCall := func(frl *FakeRouteLister, frnl *FakeRouteNamespaceLister) {
		frl.EXPECT().
			Routes(gomock.Any()).
			Return(frnl)
	}

	goodDomain := "example.com"

	goodSpace := v1alpha1.Space{}
	goodSpace.Status.NetworkConfig.Domains = []v1alpha1.SpaceDomain{
		{Domain: goodDomain},
	}
	expectGoodSpace := func(fsl *FakeSpaceLister) {
		fsl.EXPECT().
			Get(gomock.Any()).
			Return(&goodSpace, nil).
			AnyTimes()
	}

	goodRoute := v1alpha1.Route{
		Spec: v1alpha1.RouteSpec{
			RouteSpecFields: v1alpha1.RouteSpecFields{
				Domain: goodDomain,
			},
		},
	}

	expectGoodRoute := func(frnl *FakeRouteNamespaceLister) {
		frnl.EXPECT().
			List(gomock.Any()).
			Return([]*v1alpha1.Route{
				&goodRoute,
			}, nil)
	}

	testCases := map[string]struct {
		ExpectedErr error
		Setup       func(t *testing.T, f fakes)
		Domain      string
		Namespace   string
	}{
		"making VirtualServices fails": {
			Domain:      goodDomain,
			ExpectedErr: errors.New(`Error occurred while reconciling VirtualService: configuring: failed to convert path to regexp: mux: unbalanced braces in "/}invalid{"`),
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				f.frnl.EXPECT().
					List(gomock.Any()).
					Return([]*v1alpha1.Route{
						{Spec: v1alpha1.RouteSpec{RouteSpecFields: v1alpha1.RouteSpecFields{Domain: goodDomain, Path: "}invalid{"}}},
					}, nil)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fri.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, c *v1alpha1.Route, o metav1.UpdateOptions) error {
					cond := c.Status.GetCondition(v1alpha1.RouteConditionReady)
					testutil.AssertTrue(t, "condition", cond.IsTrue())
					return nil
				})
			},
		},
		"getting VirtualServices fails": {
			Domain:      goodDomain,
			ExpectedErr: errors.New("Error occurred while reconciling VirtualService: some-error"),
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				expectGoodRoute(f.frnl)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(nil, errors.New("some-error"))

				f.fri.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"VirtualServices is not found, creating VirtualService fails": {
			Domain:      goodDomain,
			ExpectedErr: errors.New("Error occurred while reconciling VirtualService: some-error"),
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				expectGoodRoute(f.frnl)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(nil, apierrors.NewNotFound(v1alpha3.Resource("VirtualService"), "VirtualService"))

				f.fvsi.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("some-error"))

				f.fri.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"VirtualServices is not found, creating VirtualService": {
			Domain: goodDomain,
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				expectGoodRoute(f.frnl)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(nil, apierrors.NewNotFound(v1alpha3.Resource("VirtualService"), "VirtualService"))

				f.fvsi.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil)

				f.fri.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"VirtualServices is being deleted": {
			Domain: goodDomain,
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				expectGoodRoute(f.frnl)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(&v1alpha3.VirtualService{
						ObjectMeta: metav1.ObjectMeta{
							DeletionTimestamp: &metav1.Time{Time: time.Now()},
						},
					}, nil)

				f.fri.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"update VirtualServices fails": {
			Domain:      goodDomain,
			ExpectedErr: errors.New("Error occurred while reconciling VirtualService: some-error"),
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				expectGoodRoute(f.frnl)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(&v1alpha3.VirtualService{}, nil)

				f.fvsi.EXPECT().
					Update(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("some-error"))

				f.fri.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, c *v1alpha1.Route, o metav1.UpdateOptions) error {
					testutil.AssertTrue(t, "condition", c.Status.GetCondition(v1alpha1.RouteConditionReady).IsTrue())
					return nil
				})
			},
		},
		"update VirtualServices fails 409": {
			Domain:      goodDomain,
			ExpectedErr: errors.New(`Error occurred while reconciling VirtualService: Operation cannot be fulfilled on virtualservices.istio.io "MyService": the object has been modified; please apply your changes to the latest version and try again`),
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				expectGoodRoute(f.frnl)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(&v1alpha3.VirtualService{}, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				conflict := apierrors.NewConflict(
					schema.GroupResource{
						Group:    "istio.io",
						Resource: "virtualservices",
					},
					"MyService",
					errors.New("the object has been modified; please apply your changes to the latest version and try again"),
				)
				f.fvsi.EXPECT().
					Update(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, conflict)

				f.fri.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, c *v1alpha1.Route, o metav1.UpdateOptions) error {
					testutil.AssertTrue(t, "condition", c.Status.GetCondition(v1alpha1.RouteConditionReady).IsTrue())
					return nil
				})
			},
		},
		"update VirtualServices": {
			Domain:    goodDomain,
			Namespace: "some-namespace",
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				expectGoodRoute(f.frnl)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(&v1alpha3.VirtualService{
						ObjectMeta: metav1.ObjectMeta{
							OwnerReferences: []metav1.OwnerReference{{Name: "some-name"}},
						},
					}, nil)

				f.fvsi.EXPECT().
					Update(gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, vs *v1alpha3.VirtualService, o metav1.UpdateOptions) {
						testutil.AssertEqual(t, "OwnerReference len", 1, len(vs.OwnerReferences))
						testutil.AssertEqual(t, "HTTPRoutes len", 1, len(vs.Spec.Http))
					})

				f.fri.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"update hosts VirtualServices": {
			Domain:    goodDomain,
			Namespace: "some-namespace",
			Setup: func(t *testing.T, f fakes) {
				expectGoodSpace(f.fsl)
				expectRouteListCall(f.frl, f.frnl)
				expectGoodRoute(f.frnl)

				f.fanl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fsibnl.EXPECT().
					List(gomock.Any()).
					Return(nil, nil)

				f.fvsnl.EXPECT().
					Get(gomock.Any()).
					Return(&v1alpha3.VirtualService{
						ObjectMeta: metav1.ObjectMeta{
							OwnerReferences: []metav1.OwnerReference{{Name: "some-name"}},
						},
						Spec: istio.VirtualService{
							Hosts: []string{goodDomain},
						},
					}, nil)

				f.fvsi.EXPECT().
					Update(gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, vs *v1alpha3.VirtualService, o metav1.UpdateOptions) {
						testutil.AssertEqual(t, "OwnerReference len", 1, len(vs.OwnerReferences))
						testutil.AssertEqual(t, "Hosts len", 2, len(vs.Spec.Hosts))
					})

				f.fri.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeRouteLister := NewFakeRouteLister(ctrl)
			fakeRouteNamespaceLister := NewFakeRouteNamespaceLister(ctrl)
			fakeNetworkingClient := NewFakeNetworkingClient(ctrl)
			fakeNetworking := NewFakeNetworking(ctrl)
			fakeVirtualServiceInterface := NewFakeVirtualServiceInterface(ctrl)
			fakeKfInterface := NewFakeKfInterface(ctrl)
			fakeKfAlpha1Interface := NewFakeKfAlpha1Interface(ctrl)
			fakeRouteInterface := NewFakeRouteInterface(ctrl)
			fakeAppLister := NewFakeAppLister(ctrl)
			fakeSpaceLister := NewFakeSpaceLister(ctrl)
			fakeAppNamespaceLister := NewFakeAppNamespaceLister(ctrl)
			fakeVirtualServiceLister := NewFakeVirtualServiceLister(ctrl)
			fakeVirtualServiceNamespaceLister := NewFakeVirtualServiceNamespaceLister(ctrl)
			fakeServiceInstanceBindingLister := NewFakeServiceInstanceBindingLister(ctrl)
			fakeServiceInstanceBindingNamespaceLister := NewFakeServiceInstanceBindingNamespaceLister(ctrl)

			fakeVirtualServiceLister.EXPECT().
				VirtualServices(gomock.Any()).
				Return(fakeVirtualServiceNamespaceLister).
				AnyTimes()

			fakeKfInterface.EXPECT().
				KfV1alpha1().
				Return(fakeKfAlpha1Interface).
				AnyTimes()

			fakeNetworkingClient.EXPECT().
				NetworkingV1alpha3().
				Return(fakeNetworking).
				AnyTimes()

			fakeNetworking.EXPECT().
				VirtualServices(gomock.Any()).
				Return(fakeVirtualServiceInterface).
				AnyTimes()

			fakeKfAlpha1Interface.EXPECT().
				Routes(gomock.Any()).
				Return(fakeRouteInterface).
				AnyTimes()

			fakeAppLister.EXPECT().
				Apps(gomock.Any()).
				Return(fakeAppNamespaceLister).
				AnyTimes()

			fakeServiceInstanceBindingLister.EXPECT().
				ServiceInstanceBindings(gomock.Any()).
				Return(fakeServiceInstanceBindingNamespaceLister).
				AnyTimes()

			if tc.Setup != nil {
				tc.Setup(t, fakes{
					fri:    fakeRouteInterface,
					frl:    fakeRouteLister,
					frnl:   fakeRouteNamespaceLister,
					fvsi:   fakeVirtualServiceInterface,
					fvsnl:  fakeVirtualServiceNamespaceLister,
					fanl:   fakeAppNamespaceLister,
					fsl:    fakeSpaceLister,
					fsibl:  fakeServiceInstanceBindingLister,
					fsibnl: fakeServiceInstanceBindingNamespaceLister,
				})
			}

			r := &Reconciler{
				Base: &reconciler.Base{
					KfClientSet: fakeKfInterface,
				},
				networkingClientSet:          fakeNetworkingClient,
				routeLister:                  fakeRouteLister,
				virtualServiceLister:         fakeVirtualServiceLister,
				appLister:                    fakeAppLister,
				spaceLister:                  fakeSpaceLister,
				serviceInstanceBindingLister: fakeServiceInstanceBindingLister,
				kfConfigStore: &testConfigStore{&config.DefaultsConfig{
					RouteTrackVirtualService: false,
				}},
			}

			err := r.ApplyChanges(
				context.Background(),
				tc.Namespace,
				tc.Domain,
			)

			testutil.AssertErrorsEqual(t, tc.ExpectedErr, err)
		})
	}
}
