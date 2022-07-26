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

package injection

import (
	"context"

	"k8s.io/client-go/rest"
	"knative.dev/pkg/injection"
)

// WithInjection returns a context for the CLI that has the
// knative.dev/pkg/injection.Default setup.
func WithInjection(ctx context.Context, restCfg *rest.Config) context.Context {
	ctx, _ = injection.Default.SetupInformers(ctx, restCfg)
	return ctx
}
