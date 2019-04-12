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

package kf

import (
	"errors"

	k8smeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deleter deletes a deployed application. It should be created via
// NewDeleter.
type Deleter struct {
	f ServingFactory
}

// NewDeleter created a new Deleter.
func NewDeleter(f ServingFactory) *Deleter {
	return &Deleter{
		f: f,
	}
}

// Delete deletes a deployed application.
func (d *Deleter) Delete(appName string, opts ...DeleteOption) error {
	cfg := DeleteOptionDefaults().Extend(opts).toConfig()

	if appName == "" {
		return errors.New("invalid app name")
	}

	client, err := d.f()
	if err != nil {
		return err
	}

	propPolicy := k8smeta.DeletePropagationForeground
	if err := client.Services(cfg.Namespace).Delete(appName, &k8smeta.DeleteOptions{
		PropagationPolicy: &propPolicy,
	}); err != nil {
		return err
	}

	return nil
}
