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

import (
	"context"

	"github.com/google/kf/v2/pkg/kf/injection"
	"github.com/google/kf/v2/pkg/kf/injection/clients/tableclient"
	fakeinjection "github.com/google/kf/v2/pkg/kf/injection/fake"
	faketable "github.com/google/kf/v2/pkg/kf/internal/tableclient/fake"
	"k8s.io/client-go/rest"
	kninjection "knative.dev/pkg/injection"
)

func init() {
	kninjection.Fake.RegisterClient(withTableClient)
}

// Get returns the Fake that is injected into the context.
func Get(ctx context.Context) *faketable.FakeInterface {
	return ctx.Value(tableclient.Key{}).(*injection.LazyInit).Create().(*faketable.FakeInterface)
}

func withTableClient(ctx context.Context, cfg *rest.Config) context.Context {
	return context.WithValue(ctx, tableclient.Key{}, injection.NewLazyInit(func() interface{} {
		return faketable.NewFakeInterface(fakeinjection.GetController(ctx))
	}))
}
