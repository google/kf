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

func TestList(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		expectedNamespace string
		appName           string
		wantErr           error
		servingFactoryErr error
		serviceListErr    error
		serviceNames      []string
		opts              []kf.ListOption
	}{
		"configured namespace": {
			expectedNamespace: "some-namespace",
			opts: []kf.ListOption{
				kf.WithListNamespace("some-namespace"),
			},
		},
		"configured app name": {
			appName:           "some-app",
			expectedNamespace: "default",
			opts: []kf.ListOption{
				kf.WithListAppName("some-app"),
			},
		},
		"formats multiple services": {
			serviceNames:      []string{"service-a", "service-b"},
			expectedNamespace: "default",
		},
		"list services error, returns error": {
			serviceListErr:    errors.New("some-error"),
			wantErr:           errors.New("some-error"),
			expectedNamespace: "default",
		},
		"serving factor error, returns error": {
			servingFactoryErr: errors.New("some-error"),
			wantErr:           errors.New("some-error"),
			expectedNamespace: "default",
		},
	} {
		t.Run(tn, func(t *testing.T) {
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

			fake.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				obj := action.(ktesting.ListActionImpl)
				testutil.AssertEqual(t, "namespace", tc.expectedNamespace, action.GetNamespace())

				if tc.appName != "" {
					if len(obj.ListRestrictions.Fields.Requirements()) != 1 {
						t.Fatalf("expected to have 1 requirement, got %d", len(obj.ListRestrictions.Fields.Requirements()))
					}

					testutil.AssertEqual(t, "FieldSelector Field", "metadata.name", obj.ListRestrictions.Fields.Requirements()[0].Field)
					testutil.AssertEqual(t, "FieldSelector Value", tc.appName, obj.ListRestrictions.Fields.Requirements()[0].Value)
				}

				return true, serviceList, tc.serviceListErr
			}))

			lister := kf.NewLister(func() (serving.ServingV1alpha1Interface, error) {
				return fake, tc.servingFactoryErr
			})

			apps, gotErr := lister.List(tc.opts...)

			if tc.wantErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
				return
			}

			for i, service := range tc.serviceNames {
				testutil.AssertEqual(t, "app", service, apps[i].Name)
			}
		})
	}
}
