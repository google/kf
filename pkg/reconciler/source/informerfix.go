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

package source

import (
	"context"

	versioned "github.com/knative/build/pkg/client/clientset/versioned"
	externalversions "github.com/knative/build/pkg/client/informers/externalversions"
	v1alpha1 "github.com/knative/build/pkg/client/informers/externalversions/build/v1alpha1"
	rest "k8s.io/client-go/rest"
	controller "knative.dev/pkg/controller"
	injection "knative.dev/pkg/injection"
	logging "knative.dev/pkg/logging"
)

func init() {
	injection.Default.RegisterClient(withClient)
	injection.Default.RegisterInformerFactory(withInformerFactory)
	injection.Default.RegisterInformer(withInformer)
}

// Key is used for associating the Informer inside the context.Context.
type Key struct{}

func withInformer(ctx context.Context) (context.Context, controller.Informer) {
	f := GetInformerFactory(ctx)
	inf := f.Build().V1alpha1().Builds()
	return context.WithValue(ctx, Key{}, inf), inf.Informer()
}

// Get extracts the typed informer from the context.
func GetBuildInformer(ctx context.Context) v1alpha1.BuildInformer {
	untyped := ctx.Value(Key{})
	if untyped == nil {
		logging.FromContext(ctx).Fatalf(
			"Unable to fetch %T from context.", (v1alpha1.BuildInformer)(nil))
	}
	return untyped.(v1alpha1.BuildInformer)
}

// Key is used as the key for associating information with a context.Context.
type FactoryKey struct{}

func withInformerFactory(ctx context.Context) context.Context {
	c := GetInformerClient(ctx)
	return context.WithValue(ctx, FactoryKey{},
		externalversions.NewSharedInformerFactory(c, controller.GetResyncPeriod(ctx)))
}

// Get extracts the InformerFactory from the context.
func GetInformerFactory(ctx context.Context) externalversions.SharedInformerFactory {
	untyped := ctx.Value(FactoryKey{})
	if untyped == nil {
		logging.FromContext(ctx).Fatalf(
			"Unable to fetch %T from context.", (externalversions.SharedInformerFactory)(nil))
	}
	return untyped.(externalversions.SharedInformerFactory)
}

// Key is used as the key for associating information with a context.Context.
type ClientKey struct{}

func withClient(ctx context.Context, cfg *rest.Config) context.Context {
	return context.WithValue(ctx, ClientKey{}, versioned.NewForConfigOrDie(cfg))
}

// Get extracts the versioned.Interface client from the context.
func GetInformerClient(ctx context.Context) versioned.Interface {
	untyped := ctx.Value(ClientKey{})
	if untyped == nil {
		logging.FromContext(ctx).Fatalf(
			"Unable to fetch %T from context.", (versioned.Interface)(nil))
	}
	return untyped.(versioned.Interface)
}
