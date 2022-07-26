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
	"strings"
	"testing"

	"github.com/google/kf/v2/pkg/apis/networking/v1alpha3"
	"github.com/google/kf/v2/pkg/client/networking/clientset/versioned/typed/networking/v1alpha3/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
	"knative.dev/pkg/apis"
)

func TestRouteValidation(t *testing.T) {
	t.Parallel()
	goodObjMeta := metav1.ObjectMeta{
		Name:      "valid",
		Namespace: "valid",
	}
	goodRouteSpec := RouteSpec{
		RouteSpecFields: RouteSpecFields{
			Hostname: "host",
			Domain:   "example.com",
		},
	}
	badMeta := metav1.ObjectMeta{
		Name: strings.Repeat("A", 64), // Too long
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
		"invalid ObjectMeta": {
			route: &Route{
				ObjectMeta: badMeta,
				Spec:       goodRouteSpec,
			},
			want: apis.ValidateObjectMetadata(badMeta.GetObjectMeta()).ViaField("metadata"),
		},
		"missing domain": {
			route: &Route{
				ObjectMeta: goodObjMeta,
				Spec: RouteSpec{
					RouteSpecFields: RouteSpecFields{
						Hostname: "test",
						Domain:   "",
					},
				},
			},
			want: apis.ErrMissingField("spec.domain"),
		},
		"missing domain with allowed on context": {
			route: &Route{
				ObjectMeta: goodObjMeta,
				Spec: RouteSpec{
					RouteSpecFields: RouteSpecFields{
						Hostname: "test",
						Domain:   "",
					},
				},
			},
			setupContext: func(ctx context.Context) context.Context {
				return withAllowEmptyDomains(ctx)
			},
		},
		"www hostname": {
			route: &Route{
				ObjectMeta: goodObjMeta,
				Spec: RouteSpec{
					RouteSpecFields: RouteSpecFields{
						Hostname: "www",
						Domain:   "domain.com",
					},
				},
			},
			want: nil,
		},
		"invalid path": {
			route: &Route{
				ObjectMeta: goodObjMeta,
				Spec: RouteSpec{
					RouteSpecFields: RouteSpecFields{
						Hostname: "some-hostname",
						Domain:   "domain.com",
						Path:     "}invalid{",
					},
				},
			},
			want: apis.ErrInvalidValue("}invalid{", "spec.path"),
		},
		"star is valid host": {
			route: &Route{
				ObjectMeta: goodObjMeta,
				Spec: RouteSpec{
					RouteSpecFields: RouteSpecFields{
						Hostname: "*",
						Domain:   "example.com",
					},
				},
			},
			want: nil,
		},
		"hostname is missing": {
			route: &Route{
				ObjectMeta: goodObjMeta,
				Spec: RouteSpec{
					RouteSpecFields: RouteSpecFields{
						Hostname: "",
						Domain:   "domain.com",
						Path:     "/myapp",
					},
				},
			},
			want: nil,
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
					return ctx
				}
			}

			tc.setup(t, f)
			ctx = tc.setupContext(ctx)

			got := tc.route.Validate(ctx)

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}

func TestBuildPathRegexp(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		path      string
		wantRegex string
		wantErr   error
	}{
		"blank path": {
			path:      "",
			wantRegex: "^(/.*)?",
		},
		"single slash": {
			path:      "/",
			wantRegex: "^(/.*)?",
		},
		"simple path": {
			path:      "/some/sub/path",
			wantRegex: "^/some/sub/path(/.*)?",
		},
		"path with whitespace": {
			path:      "/so me/ sub /path   ",
			wantRegex: "^/so me/ sub /path   (/.*)?",
		},
		"path trailing slash": {
			path:      "/trailing/slash/",
			wantRegex: "^/trailing/slash/(/.*)?",
		},
		"with regex chars": {
			path:      "/(?:foo)/.../(^-^)/$username/",
			wantRegex: `^/\(\?:foo\)/\.\.\./\(\^-\^\)/\$username/(/.*)?`,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actualRegex, actualErr := BuildPathRegexp(tc.path)
			testutil.AssertErrorsEqual(t, tc.wantErr, actualErr)
			testutil.AssertEqual(t, "regex", tc.wantRegex, actualRegex)
		})
	}
}
