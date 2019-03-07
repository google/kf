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
	cfg := DeleteOptions(opts).toConfig()

	if cfg.Namespace == ""{
		cfg.Namespace="default"
	}
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
