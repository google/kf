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

	"github.com/google/kf/v2/pkg/reconciler/apiservercerts"
	"github.com/google/kf/v2/pkg/reconciler/app"
	"github.com/google/kf/v2/pkg/reconciler/build"
	"github.com/google/kf/v2/pkg/reconciler/clusterservicebroker"
	"github.com/google/kf/v2/pkg/reconciler/featureflag"
	"github.com/google/kf/v2/pkg/reconciler/garbagecollector"
	"github.com/google/kf/v2/pkg/reconciler/reconcilerutil"
	"github.com/google/kf/v2/pkg/reconciler/route"
	"github.com/google/kf/v2/pkg/reconciler/servicebroker"
	"github.com/google/kf/v2/pkg/reconciler/serviceinstance"
	"github.com/google/kf/v2/pkg/reconciler/serviceinstancebinding"
	"github.com/google/kf/v2/pkg/reconciler/space"
	"github.com/google/kf/v2/pkg/reconciler/task"
	"github.com/google/kf/v2/pkg/reconciler/taskschedule"
	"knative.dev/pkg/webhook"
	"knative.dev/pkg/webhook/certificates"
)

func main() {
	ctx := webhook.WithOptions(context.Background(), webhook.Options{
		SecretName:  apiservercerts.SecretName,
		ServiceName: "subresource-apiserver.kf.svc",
	})

	reconcilerutil.HealthCheckerMain(ctx, ":10000", "controller",
		// Append all controllers here
		space.NewController,
		build.NewController,
		route.NewController,
		app.NewController,
		serviceinstance.NewController,
		serviceinstancebinding.NewController,
		servicebroker.NewController,
		clusterservicebroker.NewController,
		featureflag.NewController,
		task.NewController,
		taskschedule.NewController,
		certificates.NewController,
		apiservercerts.NewController,
		garbagecollector.NewController,
	)
}
