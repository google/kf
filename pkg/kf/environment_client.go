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

	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/envutil"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
)

// EnvironmentClient interacts with app's environment variables.
type EnvironmentClient interface {
	// List shows all the names and values of the environment variables for an
	// app.
	List(appName string, opts ...ListEnvOption) (map[string]string, error)

	// Set sets the given environment variables.
	Set(appName string, values map[string]string, opts ...SetEnvOption) error

	// Unset unsets the given environment variables.
	Unset(appName string, names []string, opts ...UnsetEnvOption) error
}

// environmentClient interacts with an apps environment variables. It should
// be created via NewEnvironmentClient.
type environmentClient struct {
	l AppLister
	c cserving.ServingV1alpha1Interface
}

// NewEnvironmentClient creates a new EnvironmentClient.
func NewEnvironmentClient(l AppLister, c cserving.ServingV1alpha1Interface) EnvironmentClient {
	return &environmentClient{
		l: l,
		c: c,
	}
}

// List fetches the environment variables for an app.
func (c *environmentClient) List(appName string, opts ...ListEnvOption) (map[string]string, error) {
	if appName == "" {
		return nil, errors.New("invalid app name")
	}
	cfg := ListEnvOptionDefaults().Extend(opts).toConfig()

	s, err := c.fetchService(cfg.Namespace, appName)
	if err != nil {
		return nil, err
	}

	return envutil.EnvVarsToMap(envutil.GetServiceEnvVars(s)), nil
}

// Set sets an environment variables for an app.
func (c *environmentClient) Set(appName string, values map[string]string, opts ...SetEnvOption) error {
	if appName == "" {
		return errors.New("invalid app name")
	}
	cfg := SetEnvOptionDefaults().Extend(opts).toConfig()

	s, err := c.fetchService(cfg.Namespace, appName)
	if err != nil {
		return err
	}

	newValues := envutil.EnvVarsToMap(envutil.GetServiceEnvVars(s))
	for k, v := range values {
		newValues[k] = v
	}

	envutil.SetServiceEnvVars(s, envutil.MapToEnvVars(newValues))
	if _, err := c.c.Services(cfg.Namespace).Update(s); err != nil {
		return err
	}

	return nil
}

// Unset removes environment variables for an app.
func (c *environmentClient) Unset(appName string, names []string, opts ...UnsetEnvOption) error {
	if appName == "" {
		return errors.New("invalid app name")
	}
	cfg := UnsetEnvOptionDefaults().Extend(opts).toConfig()

	s, err := c.fetchService(cfg.Namespace, appName)
	if err != nil {
		return err
	}

	newValues := envutil.RemoveEnvVars(names, envutil.GetServiceEnvVars(s))
	envutil.SetServiceEnvVars(s, newValues)
	if _, err := c.c.Services(cfg.Namespace).Update(s); err != nil {
		return err
	}

	return nil
}

func (c *environmentClient) fetchService(namespace, appName string) (*serving.Service, error) {
	return ExtractOneService(c.l.List(
		WithListNamespace(namespace),
		WithListAppName(appName),
	))
}
