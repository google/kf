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

package quotas

import (
	cv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// ClientExtension holds additional functions that should be exposed by client.
type ClientExtension interface {
}

// NewClient creates a new space client.
func NewClient(kclient cv1.ResourceQuotasGetter) Client {
	return &coreClient{
		kclient: kclient,
		upsertMutate: MutatorList{
			LabelSetMutator(map[string]string{"app.kubernetes.io/managed-by": "kf"}),
		},
		membershipValidator: LabelEqualsPredicate("app.kubernetes.io/managed-by", "kf"),
	}
}
