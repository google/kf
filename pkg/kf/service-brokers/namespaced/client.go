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

package namespaced

import (
	kf "github.com/google/kf/v2/pkg/client/kf/clientset/versioned/typed/kf/v1alpha1"
)

// ClientExtension holds additional functions that should be exposed by Client.
type ClientExtension interface {
}

// NewClient creates a new ServiceBroker client.
func NewClient(kclient kf.KfV1alpha1Interface) Client {
	return &coreClient{
		kclient: kclient,
	}
}
