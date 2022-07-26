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

	apiserviceclientset "github.com/google/kf/v2/pkg/client/kube-aggregator/clientset/versioned"
	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/system"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiregistrationlisters "k8s.io/kube-aggregator/pkg/client/listers/apiregistration/v1"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	certresources "knative.dev/pkg/webhook/certificates/resources"
)

const (
	// SecretName is the name of the v1.Secret that holds the certs.
	SecretName = "upload-api-server-secret"

	apiServiceName = "v1alpha1.upload.kf.dev"
)

// Reconciler implements a controller.Reconciler.
type Reconciler struct {
	*reconciler.Base

	apiServiceLister    apiregistrationlisters.APIServiceLister
	apiServiceClientSet apiserviceclientset.Interface
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

// Reconcile is called by knative/pkg when a new event is observed by one of
// the watchers in the controller.
//
// This controller is responsible for one thing... Keeping the API server
// secret and the caBundle field in the APIService in sync.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	// XXX: We don't care about the namespace or name. We know which resources
	// we plan on reconciling on and the informers are filtering on those.

	return r.reconcileCerts(ctx)
}

func (r *Reconciler) reconcileCerts(ctx context.Context) (err error) {
	logger := logging.FromContext(ctx)

	// Look for the secret. If it doesn't exist, we won't create one. It is
	// just an error.
	secret, err := r.SecretLister.
		Secrets(system.Namespace()).
		Get(SecretName)
	if err != nil {
		logger.Warnf("failed to get Secret: %v", err)
		return err
	}

	// Next look for the API service. If it doesn't exist, we won't create
	// one. It is just an error.
	apiService, err := r.apiServiceLister.Get(apiServiceName)
	if err != nil {
		logger.Warnf("failed to get API Service: %v", err)
		return err
	}

	// API Service
	{
		// Ensure that the Secret's CA and the API Service CA bundle match.
		if string(apiService.Spec.CABundle) == string(secret.Data[certresources.CACert]) {
			// They match, move on.
			logger.Infof("CABundle and the CACert match. Not making any updates.")
			return nil
		}

		logger.Infof("detected a mismatch in the CABundle and the CACert, updating")

		// Don't modify the informer's copy.
		desired := apiService.DeepCopy()

		desired.Spec.CABundle = secret.Data[certresources.CACert]

		if _, err := r.apiServiceClientSet.
			ApiregistrationV1().
			APIServices().
			Update(ctx, desired, metav1.UpdateOptions{}); err != nil {

			logger.Warnf("failed to update API Service: %v", err)
			return err
		}

		logger.Infof("successfully updated API Service's CABundle")
	}

	// Success!
	return nil
}
