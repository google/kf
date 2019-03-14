package secrets

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Get(name string, options ...GetOption) (map[string][]byte, error)

	// Delete removes a Kubernetes secret with the given name.
	Delete(name string, options ...DeleteOption) error
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

	if _, err := k8sClient.CoreV1().Secrets(config.Namespace).Create(secret); err != nil {
		return err
	}

	return nil
}

// Get retrieves a Kubernetes secret by its given name.
func (c *Client) Get(name string, options ...GetOption) (map[string][]byte, error) {
	config := GetOptionDefaults().Extend(options).toConfig()

	k8sClient, err := c.createK8sClient()
	if err != nil {
		return nil, err
	}

	secret, err := k8sClient.CoreV1().Secrets(config.Namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return secret.Data, nil
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
