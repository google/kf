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

package routes_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/routes"
	"github.com/google/kf/pkg/kf/commands/utils"
	routesfake "github.com/google/kf/pkg/kf/routes/fake"
	"github.com/google/kf/pkg/kf/testutil"
	"github.com/google/kf/pkg/reconciler/route/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	spacesfake "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestCreateRoute(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Namespace string
		Args      []string
		Setup     func(t *testing.T, routesfake *routesfake.FakeClient, spacesfake *spacesfake.FakeCoreV1)
		Assert    func(t *testing.T, buffer *bytes.Buffer, err error)
	}{
		"wrong number of args": {
			Args: []string{"some-space", "example.com", "extra", "--hostname=some-hostname"},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("accepts between 1 and 2 arg(s), received 3"), err)
			},
		},
		"namespace and space are different": {
			Namespace: "other-space",
			Args:      []string{"some-space", "example.com", "--hostname=some-hostname"},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New(`SPACE (argument="some-space") and namespace (flag="other-space") (if provided) must match`), err)
			},
		},
		"missing hostname flag": {
			Args:      []string{"some-space", "example.com"},
			Namespace: "some-space",
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("--hostname is required"), err)
			},
		},
		"fetching namespace fails": {
			Args:      []string{"some-space", "example.com", "--hostname=some-hostname"},
			Namespace: "some-space",
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient, spacesfake *spacesfake.FakeCoreV1) {
				spacesfake.Fake.AddReactor("*", "*", func(action ktesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("some-error")
				})
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to fetch space: some-error"), err)
			},
		},
		"creating route fails": {
			Args:      []string{"some-space", "example.com", "--hostname=some-hostname"},
			Namespace: "some-space",
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient, spacesfake *spacesfake.FakeCoreV1) {
				spacesfake.Fake.AddReactor("*", "*", func(action ktesting.Action) (bool, runtime.Object, error) {
					return false, nil, nil
				})
				routesfake.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("some-error"))
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to create Route: some-error"), err)
			},
		},
		"namespace": {
			Args:      []string{"some-space", "example.com", "--hostname=some-hostname"},
			Namespace: "some-space",
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient, spacesfake *spacesfake.FakeCoreV1) {
				spacesfake.Fake.AddReactor("*", "*", func(action ktesting.Action) (bool, runtime.Object, error) {
					testutil.AssertEqual(t, "namespace", "some-space", action.(ktesting.GetActionImpl).Name)
					return false, nil, nil
				})
				routesfake.EXPECT().Create("some-space", gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"without namespace": {
			Args: []string{"some-space", "example.com", "--hostname=some-hostname"},
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient, spacesfake *spacesfake.FakeCoreV1) {
				spacesfake.Fake.AddReactor("*", "*", func(action ktesting.Action) (bool, runtime.Object, error) {
					testutil.AssertEqual(t, "namespace", "some-space", action.(ktesting.GetActionImpl).Name)
					return false, nil, nil
				})
				routesfake.EXPECT().Create("some-space", gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New(utils.EmptyNamespaceError), err)
			},
		},
		"namespace is default and space is not": {
			Namespace: "default",
			Args:      []string{"some-space", "example.com", "--hostname=some-hostname"},
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient, spacesfake *spacesfake.FakeCoreV1) {
				spacesfake.Fake.AddReactor("*", "*", func(action ktesting.Action) (bool, runtime.Object, error) {
					testutil.AssertEqual(t, "namespace", "some-space", action.(ktesting.GetActionImpl).Name)
					return false, nil, nil
				})
				routesfake.EXPECT().Create("some-space", gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"uses namespace if SPACE is omitted": {
			Args:      []string{"example.com", "--hostname=some-hostname"},
			Namespace: "some-space",
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient, spacesfake *spacesfake.FakeCoreV1) {
				spacesfake.Fake.AddReactor("*", "*", func(action ktesting.Action) (bool, runtime.Object, error) {
					testutil.AssertEqual(t, "namespace", "some-space", action.(ktesting.GetActionImpl).Name)
					return false, nil, nil
				})
				routesfake.EXPECT().Create("some-space", gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"displays warning message if using space": {
			Args:      []string{"some-space", "example.com", "--hostname=some-hostname"},
			Namespace: "some-space",
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient, spacesfake *spacesfake.FakeCoreV1) {
				spacesfake.Fake.AddReactor("*", "*", func(action ktesting.Action) (bool, runtime.Object, error) {
					return false, nil, nil
				})
				routesfake.EXPECT().Create("some-space", gomock.Any())
			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertContainsAll(t, buffer.String(), []string{"deprecated"})
			},
		},
		"creates route with hostname and path": {
			Args:      []string{"some-space", "example.com", "--hostname=some-hostname", "--path=somepath"},
			Namespace: "some-space",
			Setup: func(t *testing.T, routesfake *routesfake.FakeClient, spacesfake *spacesfake.FakeCoreV1) {
				spacesfake.Fake.AddReactor("*", "*", func(action ktesting.Action) (bool, runtime.Object, error) {
					return true, &corev1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "some-space",
							UID:  types.UID("some-uid"),
						},
					}, nil
				})

				routesfake.EXPECT().Create(gomock.Any(),
					&v1alpha1.Route{
						TypeMeta: metav1.TypeMeta{
							Kind: "Route",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: resources.VirtualServiceName("some-hostname", "example.com", "/somepath"),
							OwnerReferences: []metav1.OwnerReference{
								{
									APIVersion: "v1",
									Kind:       "Namespace",
									Name:       "some-space",
									UID:        types.UID("some-uid"),
								},
							},
						},
						Spec: v1alpha1.RouteSpec{
							Hostname: "some-hostname",
							Domain:   "example.com",
							Path:     "/somepath",
						},
					},
				)

			},
			Assert: func(t *testing.T, buffer *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			routesfake := routesfake.NewFakeClient(ctrl)
			spacesfake := &spacesfake.FakeCoreV1{
				Fake: &ktesting.Fake{},
			}

			if tc.Setup != nil {
				tc.Setup(t, routesfake, spacesfake)
			}

			var buffer bytes.Buffer
			cmd := routes.NewCreateRouteCommand(
				&config.KfParams{
					Namespace: tc.Namespace,
				},
				routesfake,
				spacesfake,
			)
			cmd.SetArgs(tc.Args)
			cmd.SetOutput(&buffer)

			gotErr := cmd.Execute()

			if tc.Assert != nil {
				tc.Assert(t, &buffer, gotErr)
			}

			if gotErr != nil {
				return
			}
			ctrl.Finish()
		})
	}
}
