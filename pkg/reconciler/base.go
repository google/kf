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

package reconciler

import (
	"context"
	"fmt"

	kfclientset "github.com/google/kf/pkg/client/clientset/versioned"
	kfscheme "github.com/google/kf/pkg/client/clientset/versioned/scheme"
	kfclient "github.com/google/kf/pkg/client/injection/client"
	knativeclientset "github.com/google/kf/third_party/knative-serving/pkg/client/clientset/versioned"
	knativeclient "github.com/google/kf/third_party/knative-serving/pkg/client/injection/client"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	informerscorev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1listers "k8s.io/client-go/listers/core/v1"
	sharedclientset "knative.dev/pkg/client/clientset/versioned"
	sharedclient "knative.dev/pkg/client/injection/client"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/injection/clients/kubeclient"
	namespaceinformer "knative.dev/pkg/injection/informers/kubeinformers/corev1/namespace"
	secretinformer "knative.dev/pkg/injection/informers/kubeinformers/corev1/secret"
	"knative.dev/pkg/kmp"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/logging/logkey"
)

// Base implements the core controller logic, given a Reconciler.
type Base struct {
	// KubeClientSet allows us to talk to the k8s for core APIs
	KubeClientSet kubernetes.Interface

	// SharedClientSet allows us to configure shared objects
	SharedClientSet sharedclientset.Interface

	// KfClientSet allows us to configure Kf objects
	KfClientSet kfclientset.Interface

	// ServingClientSet allows us to configure Knative Serving objects
	ServingClientSet knativeclientset.Interface

	// ConfigMapWatcher allows us to watch for ConfigMap changes.
	ConfigMapWatcher configmap.Watcher

	// NamespaceLister allows us to list Namespaces. We use this to check for
	// terminating namespaces.
	NamespaceLister v1listers.NamespaceLister

	// SecretLister allows us to list Secrets.
	SecretLister v1listers.SecretLister

	// SecretInformer allows us to AddEventHandlers for Secrets.
	SecretInformer informerscorev1.SecretInformer
}

// ConfigStore is a minimal interface to the config stores used by our controllers.
type ConfigStore interface {
	ToContext(ctx context.Context) context.Context
}

// NewBase instantiates a new instance of Base implementing
// the common & boilerplate code between our reconcilers.
func NewBase(ctx context.Context, cmw configmap.Watcher) *Base {
	kubeClient := kubeclient.Get(ctx)
	nsInformer := namespaceinformer.Get(ctx)
	secretInformer := secretinformer.Get(ctx)

	base := &Base{
		KubeClientSet:    kubeClient,
		SharedClientSet:  sharedclient.Get(ctx),
		KfClientSet:      kfclient.Get(ctx),
		ServingClientSet: knativeclient.Get(ctx),
		ConfigMapWatcher: cmw,
		SecretLister:     secretInformer.Lister(),
		SecretInformer:   secretInformer,

		NamespaceLister: nsInformer.Lister(),
	}

	return base
}

func NewControllerLogger(ctx context.Context, resource string) *zap.SugaredLogger {
	return logging.FromContext(ctx).
		Named(resource).
		With(logkey.ControllerType, resource)
}

// IsNamespaceTerminating returns true if the namespace is marked as terminating
// and false if the state is unknown or not terminating.
func (base *Base) IsNamespaceTerminating(namespace string) bool {
	ns, err := base.NamespaceLister.Get(namespace)
	if err != nil || ns == nil {
		return false
	}

	return ns.Status.Phase == corev1.NamespaceTerminating
}

func (b *Base) ReconcileSecret(
	ctx context.Context,
	desired *corev1.Secret,
	actual *corev1.Secret,
) (*corev1.Secret, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	semanticEqual := equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, actual.ObjectMeta.Labels)
	semanticEqual = semanticEqual && equality.Semantic.DeepEqual(desired.Data, actual.Data)
	semanticEqual = semanticEqual && equality.Semantic.DeepEqual(desired.Type, actual.Type)

	if semanticEqual {
		return actual, nil
	}

	diff, err := kmp.SafeDiff(desired.Labels, actual.Labels)
	if err != nil {
		return nil, fmt.Errorf("failed to diff secret: %v", err)
	}
	logger.Debug("Secret.Labelsdiff:", diff)

	diff, err = kmp.SafeDiff(desired.Data, actual.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to diff secret: %v", err)
	}
	logger.Debug("Secret.Data diff:", diff)

	diff, err = kmp.SafeDiff(desired.Type, actual.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to diff secret: %v", err)
	}
	logger.Debug("Secret.Type diff:", diff)

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Data = desired.Data
	existing.Type = desired.Type
	return b.KubeClientSet.
		CoreV1().
		Secrets(existing.Namespace).
		Update(existing)
}

func init() {
	// Add serving types to the default Kubernetes Scheme so Events can be
	// logged for serving types.
	kfscheme.AddToScheme(scheme.Scheme)
}
