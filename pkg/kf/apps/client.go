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
	"github.com/google/kf/pkg/kf/systemenvinjector"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
)

// ClientExtension holds additional functions that should be exposed by client.
type ClientExtension interface {
	DeleteInForeground(namespace string, name string) error
}

type appsClient struct {
	coreClient
}

// NewClient creates a new application client.
func NewClient(kclient cserving.ServingV1alpha1Interface, envInjector systemenvinjector.SystemEnvInjectorInterface) Client {
	return &appsClient{
		coreClient{
			kclient: kclient,
			upsertMutate: MutatorList{
				envInjector.InjectSystemEnv,
				LabelSetMutator(map[string]string{"app.kubernetes.io/managed-by": "kf"}),
			},
			membershipValidator: func(_ *serving.Service) bool { return true },
		},
	}
}

// DeleteInForeground causes the deletion to happen in the foreground for
// a client. kf uses this to display correct lifecycle info.
func (ac *appsClient) DeleteInForeground(namespace string, name string) error {
	return ac.coreClient.Delete(namespace, name, WithDeleteForegroundDeletion(true))
}
