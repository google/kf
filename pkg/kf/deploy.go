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

	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/envutil"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
)

// AppLister lists the deployed apps.
type AppLister interface {
	// List lists the deployed apps.
	List(opts ...ListOption) ([]serving.Service, error)
}

// Deployer deploys an image to Knative. It will either create or update an
// existing service. It should be created via NewDeployer().
type Deployer interface {
	Deploy(s serving.Service, opts ...DeployOption) (serving.Service, error)
}

type deployer struct {
	client      cserving.ServingV1alpha1Interface
	appLister   AppLister
	envInjector SystemEnvInjectorInterface
}

// NewDeployer creates a new Deployer.
func NewDeployer(
	l AppLister,
	c cserving.ServingV1alpha1Interface,
	sei SystemEnvInjectorInterface,
) Deployer {
	return &deployer{
		appLister:   l,
		client:      c,
		envInjector: sei,
	}
}

// Deploy deploys an image to Knative.
func (d *deployer) Deploy(s serving.Service, opts ...DeployOption) (serving.Service, error) {
	namespace := DeployOptionDefaults().Extend(opts).Namespace()
	services, err := d.appLister.List(
		WithListAppName(s.Name),
		WithListNamespace(namespace),
	)
	if err != nil {
		return serving.Service{}, fmt.Errorf("failed to list apps: %s", err)
	}

	f := d.client.Services(namespace).Create
	for _, service := range services {
		if service.Name == s.Name {
			f = d.client.Services(namespace).Update
			s = d.mergeServices(s, service)
			break
		}
	}

	if err := d.envInjector.InjectSystemEnv(&s); err != nil {
		return serving.Service{}, fmt.Errorf("failed to inject system environment variables: %s", err)
	}

	service, err := f(&s)
	if err != nil {
		return serving.Service{}, fmt.Errorf("failed to create service: %s", err)
	}
	return *service, nil
}

func (d *deployer) mergeServices(newService, oldService serving.Service) serving.Service {
	newService.ResourceVersion = oldService.ResourceVersion
	newEnvs := envutil.GetServiceEnvVars(&newService)
	oldEnvs := envutil.GetServiceEnvVars(&oldService)
	envutil.SetServiceEnvVars(&newService, envutil.DeduplicateEnvVars(append(oldEnvs, newEnvs...)))
	return newService
}
