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

package fake

// Manually created to emulate injection-gen.

import (
	context "context"

	"kf-operator/pkg/operand/fake"
	client "kf-operator/pkg/operand/injection/ownerinjector"

	"k8s.io/client-go/rest"
	injection "knative.dev/pkg/injection"
	logging "knative.dev/pkg/logging"
)

func init() {
	injection.Fake.RegisterClient(withClient)
}

func withClient(ctx context.Context, cfg *rest.Config) context.Context {
	ctx, _ = With(ctx)
	return ctx
}

// With returns an OwnerInjector and a copy of ctx with that OwnerInjector associated with client Key.
func With(ctx context.Context) (context.Context, fake.OwnerInjector) {
	f := fake.Create()
	return context.WithValue(ctx, client.Key{}, f), f
}

// Get returns OwnerInjector from ctx.
func Get(ctx context.Context) fake.OwnerInjector {
	untyped := ctx.Value(client.Key{})
	if untyped == nil {
		logging.FromContext(ctx).Panic(
			"Unable to fetch fake.OwnerInjector from context.")
	}
	return untyped.(fake.OwnerInjector)
}
