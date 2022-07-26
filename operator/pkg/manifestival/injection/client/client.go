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

package client

// Manually created to emulate injection-gen.
// You can (and probably will need to when it changes) edit.

import (
	context "context"

	mfc "github.com/manifestival/client-go-client"
	mf "github.com/manifestival/manifestival"
	rest "k8s.io/client-go/rest"
	injection "knative.dev/pkg/injection"
	logging "knative.dev/pkg/logging"
)

func init() {
	injection.Default.RegisterClient(withClient)
}

// Key is used as the key for associating information with a context.Context.
type Key struct{}

func withClient(ctx context.Context, cfg *rest.Config) context.Context {
	return context.WithValue(ctx, Key{}, getManifestClientOrDie(ctx, cfg))
}

func getManifestClientOrDie(ctx context.Context, cfg *rest.Config) mf.Client {
	logger := logging.FromContext(ctx)
	mfClient, err := mfc.NewClient(cfg)
	if err != nil {
		logger.Fatal("error creating Manifest Client: ", err)
	}
	return mfClient
}

// Get extracts the mf.Client client from the context.
func Get(ctx context.Context) mf.Client {
	untyped := ctx.Value(Key{})
	if untyped == nil {
		logging.FromContext(ctx).Panic(
			"Unable to fetch mf.Client from context.")
	}
	return untyped.(mf.Client)
}
