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

package tableclient

import (
	"context"
	"log"

	"github.com/google/kf/v2/pkg/kf/injection"
	"github.com/google/kf/v2/pkg/kf/internal/tableclient"
	"k8s.io/client-go/rest"
	kninjection "knative.dev/pkg/injection"
)

func init() {
	kninjection.Default.RegisterClient(withTableClient)
}

// Key is the key used to store the TableClient on a context. It should not
// be used directly. It is exported for the fake package.
type Key struct{}

// Get returns the TableClient that is injected into the context.
func Get(ctx context.Context) tableclient.Interface {
	return ctx.Value(Key{}).(*injection.LazyInit).Create().(tableclient.Interface)
}

func withTableClient(ctx context.Context, cfg *rest.Config) context.Context {
	return context.WithValue(ctx, Key{}, injection.NewLazyInit(func() interface{} {
		tc, err := tableclient.New(cfg)
		if err != nil {
			log.Fatalf("failed to create TableClient: %v", err)
		}
		return tc
	}))
}
