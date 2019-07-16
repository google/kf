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
	cv1alpha1 "github.com/google/kf/pkg/client/clientset/versioned/typed/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/sources"
)

// ClientExtension holds additional functions that should be exposed by client.
type ClientExtension interface {
	DeleteInForeground(namespace string, name string) error
	Push(appName, srcImageName string, opts ...PushOption) error
}

type appsClient struct {
	sourcesClient sources.Client
	coreClient
}

// NewClient creates a new space client.
func NewClient(kclient cv1alpha1.AppsGetter, sourcesClient sources.Client) Client {
	return &appsClient{
		coreClient: coreClient{
			kclient: kclient,
			upsertMutate: MutatorList{
				LabelSetMutator(map[string]string{"app.kubernetes.io/managed-by": "kf"}),
			},
			membershipValidator: AllPredicate(), // all spaces can be managed by Kf
		},
		sourcesClient: sourcesClient,
	}
}

// DeleteInForeground causes the deletion to happen in the foreground for
// a client. kf uses this to display correct lifecycle info.
func (ac *appsClient) DeleteInForeground(namespace string, name string) error {
	return ac.coreClient.Delete(namespace, name, WithDeleteForegroundDeletion(true))
}
