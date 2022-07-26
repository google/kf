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

package dynamichelper

// Manually created to emulate injection-gen.
// You can (and probably will need to when it changes) edit.

import (
	context "context"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	rest "k8s.io/client-go/rest"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	injection "knative.dev/pkg/injection"
	dynamicclient "knative.dev/pkg/injection/clients/dynamicclient"
	logging "knative.dev/pkg/logging"
)

func init() {
	injection.Default.RegisterClient(withClient)
}

// An Interface wraps a restmapper and a dynamic.Interface so callers do not need
// to understand the nuances of GVK/GVR.
type Interface interface {
	dynamic.Interface
	KindFor(resource schema.GroupVersionResource) (schema.GroupVersionKind, error)
	RESTMapping(gk schema.GroupKind, versions ...string) (m *meta.RESTMapping, err error)
	Lookup(kind schema.GroupKind, namespace string, versions ...string) (dynamic.ResourceInterface, error)
}

// Key is used as the key for associating information with a context.Context.
type Key struct{}

func withClient(ctx context.Context, cfg *rest.Config) context.Context {
	return context.WithValue(ctx, Key{}, CreateDynamicHelper(ctx, kubeclient.Get(ctx), dynamicclient.Get(ctx)))
}

// Get returns DynamicHelper from ctx.
func Get(ctx context.Context) Interface {
	untyped := ctx.Value(Key{})
	if untyped == nil {
		logging.FromContext(ctx).Panic(
			"Unable to fetch dynamicclient from context.")
	}
	return untyped.(Interface)
}
