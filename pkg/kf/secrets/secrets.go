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

package secrets

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

//go:generate go run ../internal/tools/option-builder/option-builder.go options.yml

// KClientFactory builds a Kuberentes client.
type KClientFactory func() (kubernetes.Interface, error)

// ClientInterface is a client capable of interacting with Kubernetes secrets.
type ClientInterface interface {
	// Create creates a Kubernetes secret with the given name and contents.
	Create(name string, options ...CreateOption) error

	// Get retrieves a Kubernetes secret by its given name.
	Get(name string, options ...GetOption) (*corev1.Secret, error)

	// Delete removes a Kubernetes secret with the given name.
	Delete(name string, options ...DeleteOption) error

	// AddLabels adds labels to an existing secret.
	AddLabels(name string, labels map[string]string, options ...AddLabelsOption) error

	// List returns a list of secrets optionally filtered by their labels.
	List(options ...ListOption) ([]corev1.Secret, error)
}

// NewClient creates a new client capable of interacting with Kubernetes secrets.
func NewClient(kclient KClientFactory) ClientInterface {
	return &Client{
		createK8sClient: kclient,
	}
}

// Client is an implementation of Interface that works with the Service Catalog.
type Client struct {
	createK8sClient KClientFactory
}

// Create creates a Kubernetes secret with the given name and contents.
func (c *Client) Create(name string, options ...CreateOption) error {
	config := CreateOptionDefaults().Extend(options).toConfig()

	k8sClient, err := c.createK8sClient()
	if err != nil {
		return err
	}

	secret := &corev1.Secret{StringData: config.StringData, Data: config.Data}

	secret.Name = name
	secret.Kind = "Secret"
	secret.Namespace = config.Namespace
	secret.APIVersion = "v1"
	secret.Labels = config.Labels

	if _, err := k8sClient.CoreV1().Secrets(config.Namespace).Create(secret); err != nil {
		return err
	}

	return nil
}

// Get retrieves a Kubernetes secret by its given name.
func (c *Client) Get(name string, options ...GetOption) (*corev1.Secret, error) {
	config := GetOptionDefaults().Extend(options).toConfig()

	k8sClient, err := c.createK8sClient()
	if err != nil {
		return nil, err
	}

	secret, err := k8sClient.CoreV1().Secrets(config.Namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return secret, nil
}

// Delete removes a Kubernetes secret with the given name.
func (c *Client) Delete(name string, options ...DeleteOption) error {
	config := DeleteOptionDefaults().Extend(options).toConfig()

	k8sClient, err := c.createK8sClient()
	if err != nil {
		return err
	}

	return k8sClient.
		CoreV1().
		Secrets(config.Namespace).
		Delete(name, nil)
}

// AddLabels adds labels to an existing secret.
func (c *Client) AddLabels(name string, newLabels map[string]string, options ...AddLabelsOption) error {
	config := AddLabelsOptionDefaults().Extend(options).toConfig()

	k8sClient, err := c.createK8sClient()
	if err != nil {
		return err
	}

	secret, err := k8sClient.CoreV1().Secrets(config.Namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return err
	}

	secret.Labels = labels.Merge(secret.Labels, newLabels)

	if _, err := k8sClient.CoreV1().Secrets(config.Namespace).Update(secret); err != nil {
		return err
	}

	return nil
}

// List returns a list of secrets optionally filtered by their labels.
func (c *Client) List(options ...ListOption) ([]corev1.Secret, error) {
	config := ListOptionDefaults().Extend(options).toConfig()
	k8sClient, err := c.createK8sClient()
	if err != nil {
		return nil, err
	}

	secrets, err := k8sClient.CoreV1().Secrets(config.Namespace).List(v1.ListOptions{
		LabelSelector: config.LabelSelector,
	})

	if err != nil {
		return nil, err
	}

	return secrets.Items, nil
}
