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

package featureflag

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/reconciler"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/controller"
)

type Reconciler struct {
	*reconciler.Base

	namespaceLister v1listers.NamespaceLister
	configStore     *config.Store
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	ctx = r.configStore.ToContext(ctx)

	// Make sure we only reconcile the Kf namespace.
	if err := validateKfNamespaceName(key); err != nil {
		return err
	}

	ns, nsErr := r.namespaceLister.Get(v1alpha1.KfNamespace)
	if nsErr != nil {
		return nsErr
	}

	if _, err := r.reconcileFeatureFlags(ctx, ns); err != nil {
		return err
	}
	return nil
}

func validateKfNamespaceName(key string) error {
	namespace, _, err := cache.SplitMetaNamespaceKey(key)

	switch {
	case err != nil:
		return err
	case namespace != v1alpha1.KfNamespace:
		return fmt.Errorf("invalid namespace %q queued, expect only %q", namespace, v1alpha1.KfNamespace)
	}

	return nil
}

func (r *Reconciler) reconcileFeatureFlags(ctx context.Context, namespace *v1.Namespace) (*v1.Namespace, error) {
	// Don't modify the informers copy.
	existing := namespace.DeepCopy()
	featureFlagsMap := config.FeatureFlagToggles{}

	// Reconcile route services feature flag
	configDefaults, err := config.FromContext(ctx).Defaults()
	if err != nil {
		return nil, err
	}

	for k, v := range configDefaults.FeatureFlags {
		featureFlagsMap[k] = v
	}

	// Consolidate all feature flags into one annotation in Kf namespace
	marshaledMap, err := json.Marshal(featureFlagsMap)
	if err != nil {
		return nil, err
	}
	existing.ObjectMeta.Annotations[v1alpha1.FeatureFlagsAnnotation] = string(marshaledMap)

	return r.KubeClientSet.CoreV1().Namespaces().Update(ctx, existing, metav1.UpdateOptions{})
}
