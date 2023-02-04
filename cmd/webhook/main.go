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

package main

import (
	"context"

	kfvalidation "github.com/google/kf/v2/pkg/admission/validation"
	apiconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	appinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/app"
	serviceinstanceinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/serviceinstance"
	serviceinstancebindinginformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/serviceinstancebinding"
	spaceinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/space"
	buildconfig "github.com/google/kf/v2/pkg/reconciler/build/config"
	v1 "k8s.io/api/admission/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	"knative.dev/pkg/webhook"
	"knative.dev/pkg/webhook/certificates"
	"knative.dev/pkg/webhook/configmaps"
	"knative.dev/pkg/webhook/resourcesemantics"
	"knative.dev/pkg/webhook/resourcesemantics/defaulting"
	"knative.dev/pkg/webhook/resourcesemantics/validation"
)

var types = map[schema.GroupVersionKind]resourcesemantics.GenericCRD{
	v1alpha1.SchemeGroupVersion.WithKind("Space"):                  &v1alpha1.Space{},
	v1alpha1.SchemeGroupVersion.WithKind("App"):                    &v1alpha1.App{},
	v1alpha1.SchemeGroupVersion.WithKind("Build"):                  &v1alpha1.Build{},
	v1alpha1.SchemeGroupVersion.WithKind("Route"):                  &v1alpha1.Route{},
	v1alpha1.SchemeGroupVersion.WithKind("ClusterServiceBroker"):   &v1alpha1.ClusterServiceBroker{},
	v1alpha1.SchemeGroupVersion.WithKind("ServiceBroker"):          &v1alpha1.ServiceBroker{},
	v1alpha1.SchemeGroupVersion.WithKind("ServiceInstance"):        &v1alpha1.ServiceInstance{},
	v1alpha1.SchemeGroupVersion.WithKind("ServiceInstanceBinding"): &v1alpha1.ServiceInstanceBinding{},
	autoscaling.SchemeGroupVersion.WithKind("Scale"):               &v1alpha1.Scale{},
	v1alpha1.SchemeGroupVersion.WithKind("Task"):                   &v1alpha1.Task{},
	v1alpha1.SchemeGroupVersion.WithKind("TaskSchedule"):           &v1alpha1.TaskSchedule{},
	v1alpha1.SchemeGroupVersion.WithKind("SourcePackage"):          &v1alpha1.SourcePackage{},
}

var callbacks = map[schema.GroupVersionKind]validation.Callback{
	v1alpha1.SchemeGroupVersion.WithKind("ClusterServiceBroker"):   validation.NewCallback(kfvalidation.ClusterServiceBrokerValidationCallback, v1.Delete),
	v1alpha1.SchemeGroupVersion.WithKind("ServiceBroker"):          validation.NewCallback(kfvalidation.ServiceBrokerValidationCallback, v1.Delete),
	v1alpha1.SchemeGroupVersion.WithKind("ServiceInstance"):        validation.NewCallback(kfvalidation.ServiceInstanceValidationCallback, v1.Delete),
	v1alpha1.SchemeGroupVersion.WithKind("ServiceInstanceBinding"): validation.NewCallback(kfvalidation.ServiceInstanceBindingValidationCallback, v1.Create, v1.Update),
	v1alpha1.SchemeGroupVersion.WithKind("App"):                    validation.NewCallback(kfvalidation.AppValidationCallback, v1.Create, v1.Update),
	v1alpha1.SchemeGroupVersion.WithKind("Route"):                  validation.NewCallback(kfvalidation.RouteValidationCallback, v1.Create),
}

func newDefaultingAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	// Decorate contexts with the current state of the config.
	store := apiconfig.NewStore(logging.FromContext(ctx).Named("config-store"))
	store.WatchConfigs(cmw)

	return defaulting.NewAdmissionController(ctx,

		// Name of the resource webhook.
		"webhook.kf.dev",

		// The path on which to serve the webhook.
		"/defaulting",

		// The resources to validate and default.
		types,

		// A function that infuses the context passed to Validate/SetDefaults
		// with custom metadata.
		func(ctx context.Context) context.Context {
			return store.ToContext(ctx)
		},

		// Whether to disallow unknown fields.
		true,
	)
}

func newValidationAdmissionController(controllerCtx context.Context, cmw configmap.Watcher) *controller.Impl {
	// Decorate contexts with the current state of the config.
	store := apiconfig.NewStore(logging.FromContext(controllerCtx).Named("config-store"))
	store.WatchConfigs(cmw)

	// The context passed to the controller is different from the per-request context that is passed to validation requests.
	// Explicitly add the informers to the context passed to requests.
	serviceInstanceBindingInformer := serviceinstancebindinginformer.Get(controllerCtx)
	spaceInformer := spaceinformer.Get(controllerCtx)
	appInformer := appinformer.Get(controllerCtx)
	serviceInstanceInformer := serviceinstanceinformer.Get(controllerCtx)
	return validation.NewAdmissionController(controllerCtx,

		// Name of the resource webhook.
		"validation.webhook.kf.dev",

		// The path on which to serve the webhook.
		"/resource-validation",

		// The resources to validate and default.
		types,

		// A function that infuses the context passed to Validate/SetDefaults
		// with custom metadata.
		func(ctx context.Context) context.Context {
			// Add informers to the context used in validation requests.
			ctx = context.WithValue(ctx, kfvalidation.ServiceInstanceBindingInformerKey{}, serviceInstanceBindingInformer)
			ctx = context.WithValue(ctx, kfvalidation.SpaceInformerKey{}, spaceInformer)
			ctx = context.WithValue(ctx, kfvalidation.AppInformerKey{}, appInformer)
			ctx = context.WithValue(ctx, kfvalidation.ServiceInstanceInformerKey{}, serviceInstanceInformer)
			return store.ToContext(ctx)
		},

		// Whether to disallow unknown fields.
		true,

		// Callback functions that are called after initial validation.
		callbacks,
	)
}

func newConfigValidationController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return configmaps.NewAdmissionController(ctx,

		// Name of the configmap webhook.
		"config.webhook.kf.dev",

		// The path on which to serve the webhook.
		"/config-validation",

		// The configmaps to validate.
		configmap.Constructors{
			apiconfig.DefaultsConfigName:  apiconfig.NewDefaultsConfigFromConfigMap,
			buildconfig.SecretsConfigName: buildconfig.NewSecretsConfigFromConfigMap,
			metrics.ConfigMapName():       metrics.NewObservabilityConfigFromConfigMap,
			logging.ConfigMapName():       logging.NewConfigFromConfigMap,
		},
	)
}

func main() {
	// Set up a signal context with our webhook options
	ctx := webhook.WithOptions(signals.NewContext(), webhook.Options{
		ServiceName: "webhook",
		Port:        8443,
		SecretName:  "webhook-certs",
	})

	sharedmain.WebhookMainWithContext(ctx, "webhook",
		certificates.NewController,
		newDefaultingAdmissionController,
		newValidationAdmissionController,
		newConfigValidationController,
	)
}
