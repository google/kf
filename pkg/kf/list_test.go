package kf

import (
	"errors"
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"

	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	serving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestList(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name              string
		namespace         string
		wantErr           error
		servingFactoryErr error
		serviceListErr    error
		serviceNames      []string
	}{
		{
			name:      "configured namespace",
			namespace: "somenamespace",
		},
		{
			name:         "formats multiple services",
			serviceNames: []string{"service-a", "service-b"},
		},
		{
			name:           "list services error, returns error",
			serviceListErr: errors.New("some-error"),
			wantErr:        errors.New("some-error"),
		},
		{
			name:              "serving factor error, returns error",
			servingFactoryErr: errors.New("some-error"),
			wantErr:           errors.New("some-error"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expectedNamespace := tc.namespace
			if tc.namespace == "" {
				expectedNamespace = "default"
			}

			fake := &fake.FakeServingV1alpha1{
				Fake: &ktesting.Fake{},
			}

			serviceList := &v1alpha1.ServiceList{}
			for _, service := range tc.serviceNames {
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

			called := false
			fake.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				called = true
				if action.GetNamespace() != expectedNamespace {
					t.Fatalf("wanted namespace: %s, got: %s", expectedNamespace, action.GetNamespace())
				}

				return true, serviceList, tc.serviceListErr
			}))

			lister := NewLister(func() (serving.ServingV1alpha1Interface, error) {
				return fake, tc.servingFactoryErr
			})

			var opts []ListOption
			if tc.namespace != "" {
				opts = append(opts, WithListNamespace(tc.namespace))
			}

			apps, gotErr := lister.List(opts...)

			if tc.wantErr != nil {
				if fmt.Sprint(tc.wantErr) != fmt.Sprint(gotErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.wantErr, gotErr)
				}
				return
			}

			if !called {
				t.Fatal("Reactor was not invoked")
			}

			for i, service := range tc.serviceNames {
				if apps[i].Name != service {
					t.Fatalf("wanted app: %s: got:\n%v", service, apps[i].Name)
				}
			}
		})
	}
}
