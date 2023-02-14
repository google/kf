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

package space

import (
	"context"
	"fmt"
	"time"

	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/apis/networking"
	spaceinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/space"
	kflisters "github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1"
	networkpolicyinformer "github.com/google/kf/v2/pkg/client/kube/injection/informers/networking/v1/networkpolicy"
	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/reconciler/build/config"
	"github.com/google/kf/v2/pkg/reconciler/reconcilerutil"
	"github.com/google/kf/v2/pkg/system"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	deploymentinformer "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	configmapinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/configmap"
	namespaceinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/namespace"
	serviceinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/service"
	serviceaccountinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/serviceaccount"
	clusterroleinformer "knative.dev/pkg/client/injection/kube/informers/rbac/v1/clusterrole"
	clusterrolebindinginformer "knative.dev/pkg/client/injection/kube/informers/rbac/v1/clusterrolebinding"
	roleinformer "knative.dev/pkg/client/injection/kube/informers/rbac/v1/role"
	rolebindinginformer "knative.dev/pkg/client/injection/kube/informers/rbac/v1/rolebinding"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/clients/dynamicclient"
)

const (
	kccProjectAnnotation = "cnrm.cloud.google.com/project-id"
)

// NewController creates a new controller capable of reconciling Kf Spaces.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := reconciler.NewControllerLogger(ctx, "spaces.kf.dev")

	// Get informers off context
	nsInformer := namespaceinformer.Get(ctx)
	spaceInformer := spaceinformer.Get(ctx)
	serviceAccountInformer := serviceaccountinformer.Get(ctx)
	serviceInformer := serviceinformer.Get(ctx)
	npInformer := networkpolicyinformer.Get(ctx)
	deploymentInformer := deploymentinformer.Get(ctx)
	rolebindingInformer := rolebindinginformer.Get(ctx)
	roleInformer := roleinformer.Get(ctx)
	clusterRoleInformer := clusterroleinformer.Get(ctx)
	clusterRoleBindingInformer := clusterrolebindinginformer.Get(ctx)
	configMapInformer := configmapinformer.Get(ctx)

	// Dynamic client.
	dynamicClient := dynamicclient.Get(ctx)
	dynamicInformer := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		dynamicClient,
		time.Minute,
		v1alpha1.KfNamespace,
		nil,
	)

	// Setup a GSA Policy informer with the dynamic client.
	gsaPoliciesGVR, _ := schema.ParseResourceArg("iampolicies.v1beta1.iam.cnrm.cloud.google.com")
	gsaPolicyInformer := dynamicInformer.ForResource(*gsaPoliciesGVR)
	go gsaPolicyInformer.Informer().Run(ctx.Done())

	// Create reconciler
	c := &Reconciler{
		Base:                     reconciler.NewBase(ctx, cmw),
		spaceLister:              spaceInformer.Lister(),
		namespaceLister:          nsInformer.Lister(),
		serviceAccountLister:     serviceAccountInformer.Lister(),
		serviceLister:            serviceInformer.Lister(),
		networkPolicyLister:      npInformer.Lister(),
		deploymentLister:         deploymentInformer.Lister(),
		roleBindingLister:        rolebindingInformer.Lister(),
		roleLister:               roleInformer.Lister(),
		clusterRoleLister:        clusterRoleInformer.Lister(),
		clusterRoleBindingLister: clusterRoleBindingInformer.Lister(),
		gsaPolicyLister:          gsaPolicyInformer.Lister(),
		configMapLister:          configMapInformer.Lister(),
		iamClientSet:             dynamicClient.Resource(*gsaPoliciesGVR),
	}

	impl := controller.NewContext(ctx, c, controller.ControllerOptions{
		WorkQueueName: "Spaces",
		Logger:        logger,
		Reporter:      &reconcilerutil.StructuredStatsReporter{Logger: logger},
	})

	logger.Info("Setting up event handlers")
	// Watch for changes in sub-resources so we can sync accordingly
	spaceInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	// Set up all owned resources to be triggered only based on the controller.
	for _, informer := range []cache.SharedIndexInformer{
		nsInformer.Informer(),
		serviceAccountInformer.Informer(),
		npInformer.Informer(),
		c.SecretInformer.Informer(),
		roleInformer.Informer(),
		rolebindingInformer.Informer(),
		clusterRoleInformer.Informer(),
		clusterRoleBindingInformer.Informer(),
	} {
		informer.AddEventHandler(cache.FilteringResourceEventHandler{
			FilterFunc: controller.FilterControllerGVK(v1alpha1.SchemeGroupVersion.WithKind("Space")),
			Handler:    controller.HandleAll(impl.EnqueueControllerOf),
		})
	}

	// Watch for any IAM policy changes in the Kf namespace.
	gsaPolicyInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			// We can act on all the IAM policy changes in the namespace.
			return true
		},
		Handler: &pickASpaceHandler{
			c.spaceLister,
			logger,
			impl.Enqueue,
		},
	})

	// Watch the istio-system namespace to see if the revision of ASM/Istio
	// has been updated.
	impl.FilteredGlobalResync(
		controller.FilterWithName(networking.IstioNamespace),
		nsInformer.Informer(),
	)

	// Re-sync all spaces when the ingress updates
	impl.FilteredGlobalResync(
		func(obj interface{}) bool {
			if object, ok := obj.(metav1.Object); ok {
				selector := system.ClusterIngressSelector()
				return selector.Matches(labels.Set(object.GetLabels()))
			}
			return false
		},
		serviceInformer.Informer(),
	)

	logger.Info("Setting up ConfigMap receivers")
	configsToResync := []interface{}{
		&config.SecretsConfig{},
		&kfconfig.DefaultsConfig{},
	}
	resync := configmap.TypeFilter(configsToResync...)(func(string, interface{}) {
		impl.GlobalResync(spaceInformer.Informer())
	})
	configStore := config.NewStore(logger.Named("secrets-config-store"), resync)
	configStore.WatchConfigs(cmw)
	c.configStore = configStore

	kfConfigStore := kfconfig.NewStore(logger.Named("kf-config-store"), resync)
	kfConfigStore.WatchConfigs(cmw)
	c.kfConfigStore = kfConfigStore

	return impl
}

var _ reconcilerutil.HealthChecker = &Reconciler{}

// Healthy implements HealthChecker.
func (r *Reconciler) Healthy(ctx context.Context) error {
	ctx = r.configStore.ToContext(ctx)
	ctx = r.kfConfigStore.ToContext(ctx)

	configDefaults, err := kfconfig.FromContext(ctx).Defaults()
	if err != nil {
		return err
	}

	if configDefaults.FeatureFlags.AppDevExperienceBuilds().IsEnabled() {
		// All of this unnecessary when AppDevExperience Builds is being used.
		return nil
	}

	secretsConfig, err := config.FromContext(ctx).Secrets()
	if err != nil {
		return err
	}
	if secretsConfig.GoogleServiceAccount == "" {
		// GSA isn't set, so just move on.
		return nil
	}

	// Check Kf namespace KCC annotation.
	ns, err := r.namespaceLister.Get(v1alpha1.KfNamespace)
	if err != nil {
		// Return all errors from this. The namespace not being found doesn't
		// make sense, as this controller should reside in the namespace.
		return err
	}

	if project := ns.Annotations[kccProjectAnnotation]; project != secretsConfig.GoogleProjectID {
		return fmt.Errorf("Kf namespace (%q) is missing expected cnrm annotation: %s: %s. Current value is: %s", v1alpha1.KfNamespace, kccProjectAnnotation, secretsConfig.GoogleProjectID, project)
	}

	return nil
}

// pickASpaceHandler implements ResourceEventHandler. It will always just pick
// a random Space.
type pickASpaceHandler struct {
	spaceLister kflisters.SpaceLister
	logger      *zap.SugaredLogger
	enqueue     func(interface{})
}

// OnAdd implements ResourceEventHandler.
func (h *pickASpaceHandler) OnAdd(_ interface{}) {
	// Given each Space uses a single IAM policy, we want to enqueue all the
	// Spaces so each Space can get its Status updated.
	spaces, err := h.spaceLister.List(labels.Everything())
	if err != nil {
		h.logger.Warn("failed to list Spaces: %v", err)
		return
	}

	for _, space := range spaces {
		h.enqueue(space)
	}
}

// OnUpdate implements ResourceEventHandler.
func (h *pickASpaceHandler) OnUpdate(_, _ interface{}) {
	h.OnAdd(nil)
}

// OnDelete implements ResourceEventHandler.
func (h *pickASpaceHandler) OnDelete(_ interface{}) {
	h.OnAdd(nil)
}
