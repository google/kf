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

package v1alpha1

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ktesting "k8s.io/client-go/testing"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/istio/v1alpha3"
	"knative.dev/pkg/client/clientset/versioned/typed/istio/v1alpha3/fake"
)

func TestRouteValidation(t *testing.T) {
	t.Parallel()
	goodObjMeta := metav1.ObjectMeta{
		Name:      "valid",
		Namespace: "valid",
	}
	goodRouteSpec := RouteSpec{
		AppName: "some-app",
		RouteSpecFields: RouteSpecFields{
			Domain: "example.com",
		},
	}

	cases := map[string]struct {
		route        *Route
		want         *apis.FieldError
		setup        func(t *testing.T, fake *fake.FakeNetworkingV1alpha3)
		setupContext func(ctx context.Context) context.Context
	}{
		"good": {
			route: &Route{
				ObjectMeta: goodObjMeta,
				Spec:       goodRouteSpec,
			},
		},
		"don't check spec if update is status update": {
			setupContext: func(ctx context.Context) context.Context {
				return apis.WithinSubResourceUpdate(ctx, nil, "status")
			},
			route: &Route{
				ObjectMeta: metav1.ObjectMeta{Name: "", Namespace: ""},
				Spec:       goodRouteSpec,
			},
			want: nil,
		},
		"missing name": {
			route: &Route{
				ObjectMeta: metav1.ObjectMeta{Name: "", Namespace: "valid"},
				Spec:       goodRouteSpec,
			},
			want: apis.ErrMissingField("name"),
		},
		"missing appName": {
			route: &Route{
				ObjectMeta: goodObjMeta,
				Spec: RouteSpec{
					AppName: "",
					RouteSpecFields: RouteSpecFields{
						Domain: "some-domain",
					},
				},
			},
			want: apis.ErrMissingField("spec.appName"),
		},
		"missing domain": {
			route: &Route{
				ObjectMeta: goodObjMeta,
				Spec: RouteSpec{
					AppName: "app-1",
					RouteSpecFields: RouteSpecFields{
						Domain: "",
					},
				},
			},
			want: apis.ErrMissingField("spec.routeSpecFields.domain"),
		},
		"invalid hostname": {
			route: &Route{
				ObjectMeta: goodObjMeta,
				Spec: RouteSpec{
					AppName: "some-app",
					RouteSpecFields: RouteSpecFields{
						Hostname: "www",
						Domain:   "domain.com",
					},
				},
			},
			want: &apis.FieldError{
				Message: "invalid value: hostname",
				Paths:   []string{"spec.routeSpecFields.www"},
			},
		},
		"invalid path": {
			route: &Route{
				ObjectMeta: goodObjMeta,
				Spec: RouteSpec{
					AppName: "some-app",
					RouteSpecFields: RouteSpecFields{
						Hostname: "some-hostname",
						Domain:   "domain.com",
						Path:     "}invalid{",
					},
				},
			},
			want: &apis.FieldError{
				Message: "invalid value: path",
				Paths:   []string{"spec.routeSpecFields.}invalid{"},
			},
		},
		"fetching VirtualServices returns an error": {
			setup: func(t *testing.T, fake *fake.FakeNetworkingV1alpha3) {
				fake.AddReactor("get", "virtualservices", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					testutil.AssertEqual(t, "namespace", KfNamespace, action.GetNamespace())
					return true, nil, errors.New("some-error")
				})
			},
			route: &Route{
				ObjectMeta: goodObjMeta,
				Spec:       goodRouteSpec,
			},
			want: &apis.FieldError{
				Message: "failed to validate hostname + domain collisions",
				Details: "failed to fetch VirtualServices: some-error",
			},
		},
		"fetching VirtualServices returns a not found error": {
			setup: func(t *testing.T, fake *fake.FakeNetworkingV1alpha3) {
				fake.AddReactor("get", "virtualservices", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					testutil.AssertEqual(t, "namespace", KfNamespace, action.GetNamespace())
					return true, nil, apierrs.NewNotFound(schema.GroupResource{}, "some-name")
				})
			},
			route: &Route{
				ObjectMeta: goodObjMeta,
				Spec:       goodRouteSpec,
			},
		},
		"existing VirtualService has different space annotation": {
			setup: func(t *testing.T, fake *fake.FakeNetworkingV1alpha3) {
				fake.AddReactor("get", "virtualservices", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					testutil.AssertEqual(t, "namespace", KfNamespace, action.GetNamespace())
					return true, &v1alpha3.VirtualService{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"space": "some-other-space",
							},
						},
					}, nil
				})
			},
			route: &Route{
				ObjectMeta: goodObjMeta,
				Spec:       goodRouteSpec,
			},
			want: &apis.FieldError{
				Message: "Immutable field changed",
				Paths:   []string{"namespace"},
				Details: fmt.Sprintf("The route is invalid: Routes for this host and domain have been reserved for another space."),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			f := &fake.FakeNetworkingV1alpha3{
				Fake: &ktesting.Fake{},
			}

			if tc.setup == nil {
				tc.setup = func(t *testing.T, fake *fake.FakeNetworkingV1alpha3) {
					fake.AddReactor("get", "virtualservices", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						testutil.AssertEqual(t, "namespace", KfNamespace, action.GetNamespace())
						return true, &v1alpha3.VirtualService{
							ObjectMeta: metav1.ObjectMeta{
								Annotations: map[string]string{
									"space": "valid",
								},
							},
						}, nil
					})
				}
			}

			ctx := context.Background()
			if tc.setupContext == nil {
				tc.setupContext = func(ctx context.Context) context.Context {
					return SetupIstioClient(ctx, f)
				}
			}

			tc.setup(t, f)
			ctx = tc.setupContext(ctx)

			got := tc.route.Validate(ctx)

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}

func TestRouteClaimValidation(t *testing.T) {
	t.Parallel()
	goodObjMeta := metav1.ObjectMeta{
		Name:      "valid",
		Namespace: "valid",
	}
	goodRouteSpec := RouteClaimSpec{
		RouteSpecFields: RouteSpecFields{
			Domain: "example.com",
		},
	}

	cases := map[string]struct {
		route        *RouteClaim
		want         *apis.FieldError
		setup        func(t *testing.T, fake *fake.FakeNetworkingV1alpha3)
		setupContext func(ctx context.Context) context.Context
	}{
		"good": {
			route: &RouteClaim{
				ObjectMeta: goodObjMeta,
				Spec:       goodRouteSpec,
			},
		},
		"don't check spec if update is status update": {
			setupContext: func(ctx context.Context) context.Context {
				return apis.WithinSubResourceUpdate(ctx, nil, "status")
			},
			route: &RouteClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "", Namespace: ""},
				Spec:       goodRouteSpec,
			},
			want: nil,
		},
		"missing name": {
			route: &RouteClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "", Namespace: "valid"},
				Spec:       goodRouteSpec,
			},
			want: apis.ErrMissingField("name"),
		},
		"missing domain": {
			route: &RouteClaim{
				ObjectMeta: goodObjMeta,
				Spec: RouteClaimSpec{
					RouteSpecFields: RouteSpecFields{
						Domain: "",
					},
				},
			},
			want: apis.ErrMissingField("spec.routeSpecFields.domain"),
		},
		"invalid hostname": {
			route: &RouteClaim{
				ObjectMeta: goodObjMeta,
				Spec: RouteClaimSpec{
					RouteSpecFields: RouteSpecFields{
						Hostname: "www",
						Domain:   "domain.com",
					},
				},
			},
			want: &apis.FieldError{
				Message: "invalid value: hostname",
				Paths:   []string{"spec.routeSpecFields.www"},
			},
		},
		"invalid path": {
			route: &RouteClaim{
				ObjectMeta: goodObjMeta,
				Spec: RouteClaimSpec{
					RouteSpecFields: RouteSpecFields{
						Hostname: "some-hostname",
						Domain:   "domain.com",
						Path:     "}invalid{",
					},
				},
			},
			want: &apis.FieldError{
				Message: "invalid value: path",
				Paths:   []string{"spec.routeSpecFields.}invalid{"},
			},
		},
		"fetching VirtualServices returns an error": {
			setup: func(t *testing.T, fake *fake.FakeNetworkingV1alpha3) {
				fake.AddReactor("get", "virtualservices", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					testutil.AssertEqual(t, "namespace", KfNamespace, action.GetNamespace())
					return true, nil, errors.New("some-error")
				})
			},
			route: &RouteClaim{
				ObjectMeta: goodObjMeta,
				Spec:       goodRouteSpec,
			},
			want: &apis.FieldError{
				Message: "failed to validate hostname + domain collisions",
				Details: "failed to fetch VirtualServices: some-error",
			},
		},
		"fetching VirtualServices returns a not found error": {
			setup: func(t *testing.T, fake *fake.FakeNetworkingV1alpha3) {
				fake.AddReactor("get", "virtualservices", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					testutil.AssertEqual(t, "namespace", KfNamespace, action.GetNamespace())
					return true, nil, apierrs.NewNotFound(schema.GroupResource{}, "some-name")
				})
			},
			route: &RouteClaim{
				ObjectMeta: goodObjMeta,
				Spec:       goodRouteSpec,
			},
		},
		"existing VirtualService has different space annotation": {
			setup: func(t *testing.T, fake *fake.FakeNetworkingV1alpha3) {
				fake.AddReactor("get", "virtualservices", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					testutil.AssertEqual(t, "namespace", KfNamespace, action.GetNamespace())
					return true, &v1alpha3.VirtualService{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"space": "some-other-space",
							},
						},
					}, nil
				})
			},
			route: &RouteClaim{
				ObjectMeta: goodObjMeta,
				Spec:       goodRouteSpec,
			},
			want: &apis.FieldError{
				Message: "Immutable field changed",
				Paths:   []string{"namespace"},
				Details: fmt.Sprintf("The route is invalid: Routes for this host and domain have been reserved for another space."),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			f := &fake.FakeNetworkingV1alpha3{
				Fake: &ktesting.Fake{},
			}

			if tc.setup == nil {
				tc.setup = func(t *testing.T, fake *fake.FakeNetworkingV1alpha3) {
					fake.AddReactor("get", "virtualservices", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						testutil.AssertEqual(t, "namespace", KfNamespace, action.GetNamespace())
						return true, &v1alpha3.VirtualService{
							ObjectMeta: metav1.ObjectMeta{
								Annotations: map[string]string{
									"space": "valid",
								},
							},
						}, nil
					})
				}
			}

			ctx := context.Background()
			if tc.setupContext == nil {
				tc.setupContext = func(ctx context.Context) context.Context {
					return SetupIstioClient(ctx, f)
				}
			}

			tc.setup(t, f)
			ctx = tc.setupContext(ctx)

			got := tc.route.Validate(ctx)

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}
