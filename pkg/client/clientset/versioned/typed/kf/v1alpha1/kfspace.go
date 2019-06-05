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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"time"

	v1alpha1 "github.com/GoogleCloudPlatform/kf/pkg/apis/kf/v1alpha1"
	scheme "github.com/GoogleCloudPlatform/kf/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// KfSpacesGetter has a method to return a KfSpaceInterface.
// A group's client should implement this interface.
type KfSpacesGetter interface {
	KfSpaces() KfSpaceInterface
}

// KfSpaceInterface has methods to work with KfSpace resources.
type KfSpaceInterface interface {
	Create(*v1alpha1.KfSpace) (*v1alpha1.KfSpace, error)
	Update(*v1alpha1.KfSpace) (*v1alpha1.KfSpace, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.KfSpace, error)
	List(opts v1.ListOptions) (*v1alpha1.KfSpaceList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.KfSpace, err error)
	KfSpaceExpansion
}

// kfSpaces implements KfSpaceInterface
type kfSpaces struct {
	client rest.Interface
}

// newKfSpaces returns a KfSpaces
func newKfSpaces(c *KfV1alpha1Client) *kfSpaces {
	return &kfSpaces{
		client: c.RESTClient(),
	}
}

// Get takes name of the kfSpace, and returns the corresponding kfSpace object, and an error if there is any.
func (c *kfSpaces) Get(name string, options v1.GetOptions) (result *v1alpha1.KfSpace, err error) {
	result = &v1alpha1.KfSpace{}
	err = c.client.Get().
		Resource("kfspaces").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of KfSpaces that match those selectors.
func (c *kfSpaces) List(opts v1.ListOptions) (result *v1alpha1.KfSpaceList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.KfSpaceList{}
	err = c.client.Get().
		Resource("kfspaces").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested kfSpaces.
func (c *kfSpaces) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("kfspaces").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a kfSpace and creates it.  Returns the server's representation of the kfSpace, and an error, if there is any.
func (c *kfSpaces) Create(kfSpace *v1alpha1.KfSpace) (result *v1alpha1.KfSpace, err error) {
	result = &v1alpha1.KfSpace{}
	err = c.client.Post().
		Resource("kfspaces").
		Body(kfSpace).
		Do().
		Into(result)
	return
}

// Update takes the representation of a kfSpace and updates it. Returns the server's representation of the kfSpace, and an error, if there is any.
func (c *kfSpaces) Update(kfSpace *v1alpha1.KfSpace) (result *v1alpha1.KfSpace, err error) {
	result = &v1alpha1.KfSpace{}
	err = c.client.Put().
		Resource("kfspaces").
		Name(kfSpace.Name).
		Body(kfSpace).
		Do().
		Into(result)
	return
}

// Delete takes name of the kfSpace and deletes it. Returns an error if one occurs.
func (c *kfSpaces) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("kfspaces").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *kfSpaces) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("kfspaces").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched kfSpace.
func (c *kfSpaces) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.KfSpace, err error) {
	result = &v1alpha1.KfSpace{}
	err = c.client.Patch(pt).
		Resource("kfspaces").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
