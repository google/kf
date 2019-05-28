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

package kf_test

import (
	"errors"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/testutil"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLister_List(t *testing.T) {
	t.Parallel()

	setupListTest(
		t,
		func(s ...string) runtime.Object {
			return createServiceList(s)
		},
		func(l kf.AppLister, opts ...kf.ListOption) ([]string, error) {
			x, err := l.List(opts...)
			if err != nil {
				return nil, err
			}

			var names []string
			for _, s := range x {
				names = append(names, s.Name)
			}
			return names, nil
		})
}

func setupListTest(t *testing.T, resultsF func(...string) runtime.Object, listF func(kf.AppLister, ...kf.ListOption) ([]string, error)) {
	for tn, tc := range map[string]struct {
		reactor func(t *testing.T, action ktesting.Action) (handled bool, ret runtime.Object, err error)
		do      func(t *testing.T, l kf.AppLister)
	}{
		"configured namespace": {
			reactor: func(t *testing.T, action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				testutil.AssertEqual(t, "namespace", "some-namespace", action.GetNamespace())

				return false, nil, nil
			},
			do: func(t *testing.T, l kf.AppLister) {
				listMustPass(t)(listF(l, kf.WithListNamespace("some-namespace")))
			},
		},
		"default namespace": {
			reactor: func(t *testing.T, action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				testutil.AssertEqual(t, "namespace", "default", action.GetNamespace())

				return false, nil, nil
			},
			do: func(t *testing.T, l kf.AppLister) {
				listMustPass(t)(listF(l))
			},
		},
		"configured app name": {
			reactor: func(t *testing.T, action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				obj := action.(ktesting.ListActionImpl)

				if len(obj.ListRestrictions.Fields.Requirements()) != 1 {
					t.Fatalf("expected to have 1 requirement, got %d", len(obj.ListRestrictions.Fields.Requirements()))
				}

				testutil.AssertEqual(t, "FieldSelector Field", "metadata.name", obj.ListRestrictions.Fields.Requirements()[0].Field)
				testutil.AssertEqual(t, "FieldSelector Value", "some-app", obj.ListRestrictions.Fields.Requirements()[0].Value)
				return false, nil, nil
			},
			do: func(t *testing.T, l kf.AppLister) {
				listMustPass(t)(listF(l, kf.WithListAppName("some-app")))
			},
		},
		"formats multiple services": {
			reactor: func(t *testing.T, action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				testutil.AssertEqual(t, "namespace", "default", action.GetNamespace())
				return true, resultsF("service-a", "service-b"), nil
			},
			do: func(t *testing.T, l kf.AppLister) {
				expected := []string{"service-a", "service-b"}
				actual := listMustPass(t)(listF(l))
				for i, s := range expected {
					testutil.AssertEqual(t, "name", s, actual[i])
				}
			},
		},
		"list services error, returns error": {
			reactor: func(t *testing.T, action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				testutil.AssertEqual(t, "namespace", "default", action.GetNamespace())

				return true, nil, errors.New("some-error")
			},
			do: func(t *testing.T, l kf.AppLister) {
				_, err := listF(l)
				testutil.AssertErrorsEqual(t, errors.New("some-error"), err)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			fake := &fake.FakeServingV1alpha1{
				Fake: &ktesting.Fake{},
			}

			fake.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				if tc.reactor != nil {
					return tc.reactor(t, action)
				}
				return false, nil, nil
			}))

			lister := kf.NewLister(fake)

			tc.do(t, lister)
		})
	}
}

func listMustPass(t *testing.T) func([]string, error) []string {
	t.Helper()
	return func(s []string, err error) []string {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
}

func createConfigList(names []string) *v1alpha1.ConfigurationList {
	configurationList := &v1alpha1.ConfigurationList{}
	for _, configuration := range names {
		configurationList.Items = append(configurationList.Items, v1alpha1.Configuration{
			TypeMeta: metav1.TypeMeta{
				Kind:       "configuration",
				APIVersion: "serving.knative.dev/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: configuration,
			},
		})
	}

	return configurationList
}

func createServiceList(names []string) *v1alpha1.ServiceList {
	serviceList := &v1alpha1.ServiceList{}
	for _, service := range names {
		serviceList.Items = append(serviceList.Items, v1alpha1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "service",
				APIVersion: "serving.knative.dev/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: service,
			},
		})
	}

	return serviceList
}

func TestExtractOneService(t *testing.T) {
	cases := map[string]struct {
		services []v1alpha1.Service
		err      error

		expectErr error
	}{
		"error identity": {
			err:       errors.New("test-err"),
			expectErr: errors.New("test-err"),
		},
		"zero services": {
			expectErr: errors.New("expected 1 app, but found 0"),
		},
		"too many services": {
			services:  []v1alpha1.Service{v1alpha1.Service{}, v1alpha1.Service{}},
			expectErr: errors.New("expected 1 app, but found 2"),
		},
		"just right": {
			services: []v1alpha1.Service{v1alpha1.Service{}},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			svc, actualErr := kf.ExtractOneService(tc.services, tc.err)

			if tc.expectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.expectErr, actualErr)
				return
			}

			testutil.AssertNotNil(t, "service", svc)
		})
	}
}
