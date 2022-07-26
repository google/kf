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

package apps

import (
	"context"
	"io"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	cv1alpha1 "github.com/google/kf/v2/pkg/client/kf/clientset/versioned/typed/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/builds"
	"github.com/google/kf/v2/pkg/kf/logs"
)

// ClientExtension holds additional functions that should be exposed by client.
type ClientExtension interface {
	DeployLogsForApp(ctx context.Context, out io.Writer, app *v1alpha1.App) error
	DeployLogs(ctx context.Context, out io.Writer, appName, resourceVersion, namespace string, noStart bool) error
	Restart(ctx context.Context, namespace, name string) error
	Restage(ctx context.Context, namespace, name string) (*v1alpha1.App, error)
}

type appsClient struct {
	buildsClient builds.Client
	coreClient
	appTailer logs.Tailer
}

// NewClient creates a new application client.
func NewClient(
	kclient cv1alpha1.AppsGetter,
	buildsClient builds.Client,
	appTailer logs.Tailer) Client {
	return &appsClient{
		coreClient: coreClient{
			kclient: kclient,
		},
		buildsClient: buildsClient,
		appTailer:    appTailer,
	}
}

// Restart causes the controller to create a new revision for the knative
// service.
func (ac *appsClient) Restart(ctx context.Context, namespace, name string) error {
	_, err := ac.coreClient.Transform(ctx, namespace, name, func(app *v1alpha1.App) error {
		app.Spec.Template.UpdateRequests++
		return nil
	})

	return err
}

// Restage causes the controller to create a new build and then deploy the
// resulting container.
func (ac *appsClient) Restage(ctx context.Context, namespace, name string) (app *v1alpha1.App, err error) {
	return ac.coreClient.Transform(ctx, namespace, name, func(app *v1alpha1.App) error {
		app.Spec.Build.UpdateRequests++
		return nil
	})
}
