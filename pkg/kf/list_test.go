package kf_test

import (
	"errors"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	serving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
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
		func(l *kf.Lister, opts ...kf.ListOption) ([]string, error) {
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

func setupListTest(t *testing.T, resultsF func(...string) runtime.Object, listF func(*kf.Lister, ...kf.ListOption) ([]string, error)) {
	for tn, tc := range map[string]struct {
		servingFactoryErr error

		reactor func(t *testing.T, action ktesting.Action) (handled bool, ret runtime.Object, err error)
		do      func(t *testing.T, l *kf.Lister)
	}{
		"configured namespace": {
			reactor: func(t *testing.T, action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				testutil.AssertEqual(t, "namespace", "some-namespace", action.GetNamespace())

				return false, nil, nil
			},
			do: func(t *testing.T, l *kf.Lister) {
				listMustPass(t)(listF(l, kf.WithListNamespace("some-namespace")))
			},
		},
		"default namespace": {
			reactor: func(t *testing.T, action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				testutil.AssertEqual(t, "namespace", "default", action.GetNamespace())

				return false, nil, nil
			},
			do: func(t *testing.T, l *kf.Lister) {
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
			do: func(t *testing.T, l *kf.Lister) {
				listMustPass(t)(listF(l, kf.WithListAppName("some-app")))
			},
		},
		"formats multiple services": {
			reactor: func(t *testing.T, action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				testutil.AssertEqual(t, "namespace", "default", action.GetNamespace())
				return true, resultsF("service-a", "service-b"), nil
			},
			do: func(t *testing.T, l *kf.Lister) {
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
			do: func(t *testing.T, l *kf.Lister) {
				_, err := listF(l)
				testutil.AssertErrorsEqual(t, errors.New("some-error"), err)
			},
		},
		"serving factor error, returns error": {
			servingFactoryErr: errors.New("some-error"),
			reactor: func(t *testing.T, action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return false, nil, nil
			},
			do: func(t *testing.T, l *kf.Lister) {
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

			lister := kf.NewLister(func() (serving.ServingV1alpha1Interface, error) {
				return fake, tc.servingFactoryErr
			})

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
