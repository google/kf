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
	"io"

	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	cv1alpha1 "github.com/google/kf/pkg/client/clientset/versioned/typed/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/algorithms"
	"github.com/google/kf/pkg/kf/sources"
)

// ClientExtension holds additional functions that should be exposed by client.
type ClientExtension interface {
	DeployLogsForApp(out io.Writer, app *v1alpha1.App) error
	DeployLogs(out io.Writer, appName, resourceVersion, namespace string, noStart bool) error
	Restart(namespace, name string) error
	Restage(namespace, name string) (*v1alpha1.App, error)
	BindService(namespace, name string, binding *v1alpha1.AppSpecServiceBinding) (*v1alpha1.App, error)
	UnbindService(namespace, name, bindingName string) (*v1alpha1.App, error)
}

type appsClient struct {
	sourcesClient sources.Client
	coreClient
}

// NewClient creates a new application client.
func NewClient(
	kclient cv1alpha1.AppsGetter,
	sourcesClient sources.Client) Client {
	return &appsClient{
		coreClient: coreClient{
			kclient: kclient,
			upsertMutate: func(app *v1alpha1.App) error {
				// Dedupe Routes
				// TODO(https://github.com/knative/pkg/issues/542): Route
				// already exists and the webhook can't dedupe for us.
				app.Spec.Routes = []v1alpha1.RouteSpecFields(
					algorithms.Dedupe(
						v1alpha1.RouteSpecFieldsSlice(app.Spec.Routes),
					).(v1alpha1.RouteSpecFieldsSlice),
				)

				return nil
			},
		},
		sourcesClient: sourcesClient,
	}
}

// Restart causes the controller to create a new revision for the knative
// service.
func (ac *appsClient) Restart(namespace, name string) error {
	_, err := ac.coreClient.Transform(namespace, name, func(app *v1alpha1.App) error {
		app.Spec.Template.UpdateRequests++
		return nil
	})

	return err
}

// Restage causes the controller to create a new build and then deploy the
// resulting container.
func (ac *appsClient) Restage(namespace, name string) (app *v1alpha1.App, err error) {
	app, err = ac.coreClient.Get(namespace, name)
	if err != nil {
		return
	}

	app.Spec.Source.UpdateRequests++

	return ac.coreClient.Update(namespace, app)
}

// BindService adds the given service binding to the app.
func (ac *appsClient) BindService(namespace, name string, binding *v1alpha1.AppSpecServiceBinding) (app *v1alpha1.App, err error) {
	return ac.coreClient.Transform(namespace, name, func(app *v1alpha1.App) error {
		BindService(app, binding)
		return nil
	})
}

// UnbindService removes the given service binding from the app.
func (ac *appsClient) UnbindService(namespace, name, bindingName string) (app *v1alpha1.App, err error) {
	return ac.coreClient.Transform(namespace, name, func(app *v1alpha1.App) error {
		UnbindService(app, bindingName)
		return nil
	})
}
