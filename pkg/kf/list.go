package kf

import (
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ServingFactory returns a client for serving.
type ServingFactory func() (cserving.ServingV1alpha1Interface, error)

// Lister lists deployed applications. It should be created via NewLister.
type Lister struct {
	f ServingFactory
}

// NewLister creates a new Lister.
func NewLister(f ServingFactory) *Lister {
	return &Lister{
		f: f,
	}
}

// List lists the deployed applications for the given namespace.
func (l *Lister) List(opts ...ListOption) ([]serving.Service, error) {
	cfg := ListOptionDefaults().Extend(opts).toConfig()
	client, listOptions, err := l.setup(cfg, opts)
	if err != nil {
		return nil, err
	}

	services, err := client.Services(cfg.Namespace).List(listOptions)
	if err != nil {
		return nil, err
	}
	services.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "knative.dev",
		Version: "v1alpha1",
		Kind:    "Service",
	})

	return services.Items, nil
}

// ListConfigurations lists the deployed Configurations for a given namespace.
func (l *Lister) ListConfigurations(opts ...ListConfigurationsOption) ([]serving.Configuration, error) {
	cfg := ListConfigurationsOptionDefaults().Extend(opts).toConfig()
	client, listOptions, err := l.setupConfiguration(cfg, opts)
	if err != nil {
		return nil, err
	}

	configs, err := client.Configurations(cfg.Namespace).List(listOptions)
	if err != nil {
		return nil, err
	}

	return configs.Items, nil
}

func (l *Lister) setup(cfg listConfig, opts []ListOption) (cserving.ServingV1alpha1Interface, v1.ListOptions, error) {
	client, err := l.f()
	if err != nil {
		return nil, v1.ListOptions{}, err
	}

	listOptions := v1.ListOptions{}
	if cfg.AppName != "" {
		listOptions.FieldSelector = "metadata.name=" + cfg.AppName
	}

	return client, listOptions, nil
}

func (l *Lister) setupConfiguration(cfg listConfigurationsConfig, opts []ListConfigurationsOption) (cserving.ServingV1alpha1Interface, v1.ListOptions, error) {
	client, err := l.f()
	if err != nil {
		return nil, v1.ListOptions{}, err
	}

	listOptions := v1.ListOptions{}
	if cfg.AppName != "" {
		listOptions.FieldSelector = "metadata.name=" + cfg.AppName
	}

	return client, listOptions, nil
}
