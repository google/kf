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
	"github.com/google/kf/pkg/kf/sources"
	"github.com/google/kf/pkg/kf/systemenvinjector"
)

// ClientExtension holds additional functions that should be exposed by client.
type ClientExtension interface {
	DeleteInForeground(namespace string, name string) error

	// DeployLogs writes the logs for the build and deploy stage to the given
	// out.  The method exits once the logs are done streaming.
	DeployLogs(out io.Writer, appName, resourceVersion, namespace string, noStart bool) error
	Restart(namespace, name string) error
	Restage(namespace, name string) error
}

type appsClient struct {
	sourcesClient sources.Client
	coreClient
}

// NewClient creates a new application client.
func NewClient(
	kclient cv1alpha1.AppsGetter,
	envInjector systemenvinjector.SystemEnvInjectorInterface,
	sourcesClient sources.Client) Client {
	return &appsClient{
		coreClient: coreClient{
			kclient: kclient,
			upsertMutate: MutatorList{
				LabelSetMutator(map[string]string{"app.kubernetes.io/managed-by": "kf"}),
			},
			membershipValidator: AllPredicate(), // all apps can be managed by Kf
		},
		sourcesClient: sourcesClient,
	}
}

// DeleteInForeground causes the deletion to happen in the foreground for
// a client. kf uses this to display correct lifecycle info.
func (ac *appsClient) DeleteInForeground(namespace string, name string) error {
	return ac.coreClient.Delete(namespace, name, WithDeleteForegroundDeletion(true))
}

// Restart causes the controller to create a new revision for the knative
// service.
func (ac *appsClient) Restart(namespace, name string) error {
	return ac.coreClient.Transform(namespace, name, func(a *v1alpha1.App) error {
		a.Spec.Template.UpdateRequests++
		return nil
	})
}

// Restage causes the controller to create a new build and then deploy the
// resulting container.
func (ac *appsClient) Restage(namespace, name string) error {
	return ac.coreClient.Transform(namespace, name, func(a *v1alpha1.App) error {
		a.Spec.Source.UpdateRequests++
		return nil
	})
}
