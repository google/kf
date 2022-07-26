// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apiservercerts

import (
	"context"
	"errors"
	"testing"

	apiserviceclientfake "github.com/google/kf/v2/pkg/client/kube-aggregator/clientset/versioned/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/system"
	_ "github.com/google/kf/v2/pkg/system/testing"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
	apiv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	certresources "knative.dev/pkg/webhook/certificates/resources"
)

//go:generate go run ../../kf/internal/tools/fakelister/generator.go --pkg apiservercerts --object-type Secret --object-pkg k8s.io/api/core/v1 --lister-pkg k8s.io/client-go/listers/core/v1
//go:generate go run ../../kf/internal/tools/fakelister/generator.go --pkg apiservercerts --object-type APIService --object-pkg k8s.io/kube-aggregator/pkg/apis/apiregistration/v1 --lister-pkg k8s.io/kube-aggregator/pkg/client/listers/apiregistration/v1 --namespaced=false

func TestReconciler_Reconcile(t *testing.T) {
	t.Parallel()

	caCert := []byte("some-cert")
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SecretName,
			Namespace: system.Namespace(),
		},
		Data: map[string][]byte{
			certresources.CACert: caCert,
		},
	}

	apiService := &apiv1.APIService{
		ObjectMeta: metav1.ObjectMeta{
			Name: apiServiceName,
		},
		Spec: apiv1.APIServiceSpec{
			CABundle: caCert,
		},
	}

	type fakes struct {
		fakeSecretLister     *fakeSecretLister
		fakeAPIServiceLister *fakeAPIServiceLister

		fakeAPIServiceClientSet *apiserviceclientfake.Clientset
		fakeClientSet           *fake.Clientset
	}

	testCases := []struct {
		name  string
		setup func(t *testing.T, f *fakes)
		err   error
	}{
		{
			name: "getting secret fails",
			setup: func(t *testing.T, f *fakes) {
				f.fakeSecretLister.err = errors.New("some-error")
			},
			err: errors.New("some-error"),
		},
		{
			name: "getting API Service fails",
			setup: func(t *testing.T, f *fakes) {
				f.fakeSecretLister.Add(secret)
				f.fakeAPIServiceLister.err = errors.New("some-error")
			},
			err: errors.New("some-error"),
		},
		{
			name: "does NOT update anything if they match",
			setup: func(t *testing.T, f *fakes) {
				f.fakeSecretLister.Add(secret)
				f.fakeAPIServiceLister.Add(apiService)

				f.fakeAPIServiceClientSet.PrependReactor("update", "apiservices",
					func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						t.Fail()
						return false, nil, nil
					},
				)
			},
		},
		{
			name: "updating API Service Client fails",
			setup: func(t *testing.T, f *fakes) {

				// Make a copy and altere the secret to ensure they aren't the
				// same and require an update.
				copiedSecret := secret.DeepCopy()
				copiedSecret.Data[certresources.CACert] = []byte("different")

				f.fakeSecretLister.Add(copiedSecret)
				f.fakeAPIServiceLister.Add(apiService)
				f.fakeAPIServiceClientSet.PrependReactor("update", "apiservices",
					func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, errors.New("some-error")
					},
				)
			},
			err: errors.New("some-error"),
		},
		{
			name: "success",
			setup: func(t *testing.T, f *fakes) {

				// Make a copy and altere the secret to ensure they aren't the
				// same and require an update.
				copiedSecret := secret.DeepCopy()
				copiedSecret.Data[certresources.CACert] = []byte("different")

				f.fakeSecretLister.Add(copiedSecret)
				f.fakeAPIServiceLister.Add(apiService)
				f.fakeAPIServiceClientSet.PrependReactor("update", "apiservices",
					func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						s := action.(ktesting.UpdateActionImpl).Object.(*apiv1.APIService)

						testutil.AssertEqual(t, "name", apiServiceName, s.Name)
						testutil.AssertEqual(t, "caBundle", "different", string(s.Spec.CABundle))

						return true, ret, nil
					},
				)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			f := &fakes{
				fakeSecretLister:     &fakeSecretLister{},
				fakeAPIServiceLister: &fakeAPIServiceLister{},

				fakeAPIServiceClientSet: apiserviceclientfake.NewSimpleClientset(),
				fakeClientSet:           fake.NewSimpleClientset(),
			}

			// Ensure the informer copies weren't altered.
			defer func() {
				f.fakeSecretLister.AssertCacheIsPreserved(t)
				f.fakeAPIServiceLister.AssertCacheIsPreserved(t)
			}()

			if tc.setup != nil {
				tc.setup(t, f)
			}

			r := Reconciler{
				Base: &reconciler.Base{
					SecretLister:  f.fakeSecretLister,
					KubeClientSet: f.fakeClientSet,
				},
				apiServiceLister:    f.fakeAPIServiceLister,
				apiServiceClientSet: f.fakeAPIServiceClientSet,
			}

			err := r.Reconcile(context.Background(), "doesn't matter")
			testutil.AssertErrorsEqual(t, tc.err, err)
		})
	}
}
