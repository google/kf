package commands

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestDeleteCommand(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name              string
		namespace         string
		wantErr           error
		servingFactoryErr error
		serviceDeleteErr  error
	}{
		{
			name:      "deletes given app in namespace",
			namespace: "some-namespace",
		},
		{
			name:              "serving factory error",
			wantErr:           errors.New("some error"),
			servingFactoryErr: errors.New("some error"),
		},
		{
			name:             "service delete error",
			wantErr:          errors.New("some error"),
			serviceDeleteErr: errors.New("some error"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			fake := &fake.FakeServingV1alpha1{
				Fake: &ktesting.Fake{},
			}
			const appName = "some-app"

			called := false
			fake.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				called = true
				if action.GetNamespace() != tc.namespace {
					t.Fatalf("wanted namespace: %s, got: %s", tc.namespace, action.GetNamespace())
				}

				if !action.Matches("delete", "services") {
					t.Fatal("wrong action")
				}

				if gn := action.(ktesting.DeleteAction).GetName(); gn != appName {
					t.Fatalf("wanted app name %s, got %s", appName, gn)
				}

				return tc.serviceDeleteErr != nil, nil, tc.serviceDeleteErr
			}))

			c := NewDeleteCommand(&KfParams{
				Namespace: tc.namespace,
				Output:    buffer,
				ServingFactory: func() (cserving.ServingV1alpha1Interface, error) {
					return fake, tc.servingFactoryErr
				},
			})

			gotErr := c.RunE(c, []string{appName})
			if tc.wantErr != nil || gotErr != nil {
				if fmt.Sprint(tc.wantErr) != fmt.Sprint(gotErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.wantErr, gotErr)
				}

				if !c.SilenceUsage {
					t.Fatalf("wanted %v, got %v", true, c.SilenceUsage)
				}

				return
			}

			if !called {
				t.Fatal("Reactor was not invoked")
			}
		})
	}
}
