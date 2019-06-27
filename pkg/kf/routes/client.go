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

package routes

import (
	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	kf "github.com/google/kf/pkg/client/clientset/versioned/typed/kf/v1alpha1"
)

// ClientExtension holds additional functions that should be exposed by client.
type ClientExtension interface {
}

// NewClient creates a new route client.
func NewClient(kclient kf.KfV1alpha1Interface) Client {
	return &coreClient{
		kclient:             kclient,
		membershipValidator: func(_ *v1alpha1.Route) bool { return true },
	}
}
