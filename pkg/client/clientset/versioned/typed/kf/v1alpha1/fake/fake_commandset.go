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

package fake

import (
	v1alpha1 "github.com/GoogleCloudPlatform/kf/pkg/apis/kf/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeCommandSets implements CommandSetInterface
type FakeCommandSets struct {
	Fake *FakeKfV1alpha1
	ns   string
}

var commandsetsResource = schema.GroupVersionResource{Group: "kf.dev", Version: "v1alpha1", Resource: "commandsets"}

var commandsetsKind = schema.GroupVersionKind{Group: "kf.dev", Version: "v1alpha1", Kind: "CommandSet"}

// Get takes name of the commandSet, and returns the corresponding commandSet object, and an error if there is any.
func (c *FakeCommandSets) Get(name string, options v1.GetOptions) (result *v1alpha1.CommandSet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(commandsetsResource, c.ns, name), &v1alpha1.CommandSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CommandSet), err
}

// List takes label and field selectors, and returns the list of CommandSets that match those selectors.
func (c *FakeCommandSets) List(opts v1.ListOptions) (result *v1alpha1.CommandSetList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(commandsetsResource, commandsetsKind, c.ns, opts), &v1alpha1.CommandSetList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.CommandSetList{ListMeta: obj.(*v1alpha1.CommandSetList).ListMeta}
	for _, item := range obj.(*v1alpha1.CommandSetList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested commandSets.
func (c *FakeCommandSets) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(commandsetsResource, c.ns, opts))

}

// Create takes the representation of a commandSet and creates it.  Returns the server's representation of the commandSet, and an error, if there is any.
func (c *FakeCommandSets) Create(commandSet *v1alpha1.CommandSet) (result *v1alpha1.CommandSet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(commandsetsResource, c.ns, commandSet), &v1alpha1.CommandSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CommandSet), err
}

// Update takes the representation of a commandSet and updates it. Returns the server's representation of the commandSet, and an error, if there is any.
func (c *FakeCommandSets) Update(commandSet *v1alpha1.CommandSet) (result *v1alpha1.CommandSet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(commandsetsResource, c.ns, commandSet), &v1alpha1.CommandSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CommandSet), err
}

// Delete takes name of the commandSet and deletes it. Returns an error if one occurs.
func (c *FakeCommandSets) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(commandsetsResource, c.ns, name), &v1alpha1.CommandSet{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCommandSets) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(commandsetsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.CommandSetList{})
	return err
}

// Patch applies the patch and returns the patched commandSet.
func (c *FakeCommandSets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CommandSet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(commandsetsResource, c.ns, name, pt, data, subresources...), &v1alpha1.CommandSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CommandSet), err
}
