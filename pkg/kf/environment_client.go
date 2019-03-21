package kf

import (
	"errors"
	"fmt"

	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// EnvironmentClient interacts with an apps environment variables. It should
// be created via NewEnvironmentClient.
type EnvironmentClient struct {
	l AppLister
	f ServingFactory
}

// NewEnvironmentClient creates a new EnvironmentClient.
func NewEnvironmentClient(l AppLister, f ServingFactory) *EnvironmentClient {
	return &EnvironmentClient{
		l: l,
		f: f,
	}
}

// Get fetches the current value of an environment variable.
func (c *EnvironmentClient) List(appName string, opts ...ListEnvOption) (map[string]string, error) {
	if appName == "" {
		return nil, errors.New("invalid app name")
	}
	cfg := ListEnvOptionDefaults().Extend(opts).toConfig()

	s, err := c.fetchService(cfg.Namespace, appName)
	if err != nil {
		return nil, err
	}

	results := map[string]string{}
	for _, env := range s.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Env {
		results[env.Name] = env.Value
	}

	return results, err
}

// Get fetches the current value of an environment variable.
func (c *EnvironmentClient) Set(appName string, values map[string]string, opts ...SetEnvOption) error {
	if appName == "" {
		return errors.New("invalid app name")
	}
	cfg := SetEnvOptionDefaults().Extend(opts).toConfig()

	client, err := c.f()
	if err != nil {
		return err
	}

	s, err := c.fetchService(cfg.Namespace, appName)
	if err != nil {
		return err
	}

	newValues := c.dedupeEnvs(
		values,
		s.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Env,
	)

	s.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Env = c.mapToEnvs(newValues)
	if _, err := client.Services(cfg.Namespace).Update(&s); err != nil {
		return err
	}

	return nil
}

func (c *EnvironmentClient) dedupeEnvs(values map[string]string, envs []corev1.EnvVar) map[string]string {
	// Create a new map so that we can prioritize the new values over the
	// existing.
	newValues := map[string]string{}
	for _, e := range envs {
		newValues[e.Name] = e.Value
	}
	for n, v := range values {
		newValues[n] = v
	}

	return newValues
}

func (c *EnvironmentClient) mapToEnvs(values map[string]string) []corev1.EnvVar {
	var envs []corev1.EnvVar
	for n, v := range values {
		envs = append(envs, corev1.EnvVar{Name: n, Value: v})
	}
	return envs
}

func (c *EnvironmentClient) fetchService(namespace, appName string) (serving.Service, error) {
	services, err := c.l.List(
		WithListNamespace(namespace),
		WithListAppName(appName),
	)
	if err != nil {
		return serving.Service{}, err
	}

	if len(services) != 1 {
		return serving.Service{}, fmt.Errorf("unknown app: '%s'", appName)
	}

	return services[0], nil
}
