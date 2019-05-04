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
	"fmt"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestDeleteCommand(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		namespace        string
		appName          string
		wantErr          error
		serviceDeleteErr error
	}{
		"deletes given app in namespace": {
			namespace: "some-namespace",
			appName:   "some-app",
		},
		"deletes given app in the default namespace": {
			appName: "some-app",
		},
		"empty app name, returns error": {
			wantErr: errors.New("invalid app name"),
		},
		"service delete error": {
			wantErr:          errors.New("some error"),
			serviceDeleteErr: errors.New("some error"),
			appName:          "some-app",
		},
	} {
		t.Run(tn, func(t *testing.T) {
			fake := &fake.FakeServingV1alpha1{
				Fake: &ktesting.Fake{},
			}

			expectedNamespace := tc.namespace
			if tc.namespace == "" {
				expectedNamespace = "default"
			}

			called := false
			fake.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				called = true
				if action.GetNamespace() != expectedNamespace {
					t.Fatalf("wanted namespace: %s, got: %s", expectedNamespace, action.GetNamespace())
				}

				if !action.Matches("delete", "services") {
					t.Fatal("wrong action")
				}

				if gn := action.(ktesting.DeleteAction).GetName(); gn != tc.appName {
					t.Fatalf("wanted app name %s, got %s", tc.appName, gn)
				}

				return tc.serviceDeleteErr != nil, nil, tc.serviceDeleteErr
			}))

			d := kf.NewDeleter(fake)

			var opts []kf.DeleteOption
			if tc.namespace != "" {
				opts = append(opts, kf.WithDeleteNamespace(tc.namespace))
			}

			gotErr := d.Delete(tc.appName, opts...)
			if tc.wantErr != nil || gotErr != nil {
				if fmt.Sprint(tc.wantErr) != fmt.Sprint(gotErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.wantErr, gotErr)
				}

				return
			}

			if !called {
				t.Fatal("Reactor was not invoked")
			}
		})
	}
}
