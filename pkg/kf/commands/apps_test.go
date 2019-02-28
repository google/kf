package commands

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"

	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	serving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAppsCommand(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name             string
		namespace        string
		wantErr          error
		servingFactorErr error
		serviceListErr   error
		serviceNames     []string
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
			name:             "serving factor error, returns error",
			servingFactorErr: errors.New("some-error"),
			wantErr:          errors.New("some-error"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			fake := &fake.FakeServingV1alpha1{
				Fake: &ktesting.Fake{},
			}

			items := []v1alpha1.Service{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "should-not-see",
						APIVersion: "serving.knative.dev/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "should-not-see",
					},
				},
			}
			for _, service := range tc.serviceNames {
				items = append(items, v1alpha1.Service{
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
				if action.GetNamespace() != tc.namespace {
					t.Fatalf("wanted namespace: %s, got: %s", tc.namespace, action.GetNamespace())
				}

				return true, &v1alpha1.ServiceList{Items: items}, tc.serviceListErr
			}))

			c := NewAppsCommand(&KfParams{
				Namespace: tc.namespace,
				Output:    buffer,
				ServingFactory: func() (serving.ServingV1alpha1Interface, error) {
					return fake, tc.servingFactorErr
				},
			})

			gotErr := c.RunE(&cobra.Command{}, nil)
			if tc.wantErr != nil {
				if fmt.Sprint(tc.wantErr) != fmt.Sprint(gotErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.wantErr, gotErr)
				}
				return
			}

			defer func() {
				if !called {
					t.Fatal("Reactor was not invoked")
				}
			}()

			for _, service := range tc.serviceNames {
				if strings.Index(buffer.String(), service) < 0 {
					t.Fatalf("wanted output: %s: got:\n%v", service, buffer.String())
				}
			}
		})
	}
}
