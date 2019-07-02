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
	"fmt"

	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/internal/envutil"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
)

// Deployer deploys an image to Knative. It will either create or update an
// existing service. It should be created via NewDeployer().
type Deployer interface {
	Deploy(s serving.Service, opts ...DeployOption) (*serving.Service, error)
}

type deployer struct {
	appsClient apps.Client
}

// NewDeployer creates a new Deployer.
func NewDeployer(appsClient apps.Client) Deployer {
	return &deployer{
		appsClient: appsClient,
	}
}

// Deploy deploys an image to Knative.
func (d *deployer) Deploy(s serving.Service, opts ...DeployOption) (*serving.Service, error) {
	namespace := DeployOptionDefaults().Extend(opts).Namespace()

	service, err := d.appsClient.Upsert(namespace, &s, d.mergeServices)
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %s", err)
	}
	return service, nil
}

func (*deployer) mergeServices(newService, oldService *serving.Service) *serving.Service {
	newService.ResourceVersion = oldService.ResourceVersion
	newEnvs := envutil.GetServiceEnvVars(newService)
	oldEnvs := envutil.GetServiceEnvVars(oldService)
	envutil.SetServiceEnvVars(newService, envutil.DeduplicateEnvVars(append(oldEnvs, newEnvs...)))
	return newService
}
