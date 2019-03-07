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
	cfg := ListOptions(opts).toConfig()
	if cfg.Namespace == "" {
		cfg.Namespace = "default"
	}

	client, err := l.f()
	if err != nil {
		return nil, err
	}

	services, err := client.Services(cfg.Namespace).List(v1.ListOptions{})
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
