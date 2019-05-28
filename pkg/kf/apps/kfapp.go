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
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
)

// KfApp provides a facade around Knative services for accessing and mutating its
// values.
type KfApp serving.Service

// GetName retrieves the name of the app.
func (k *KfApp) GetName() string {
	return k.Name
}

// SetName sets the name of the app.
func (k *KfApp) SetName(name string) {
	k.Name = name
}

// ToService casts this alias back into a Service.
func (k *KfApp) ToService() *serving.Service {
	svc := serving.Service(*k)
	return &svc
}

// NewKfApp creates a new KfApp.
func NewKfApp() KfApp {
	return KfApp{}
}
