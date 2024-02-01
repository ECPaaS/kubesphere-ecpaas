/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	v2beta2 "kubesphere.io/api/notification/v2beta2"
)

// FakeSilences implements SilenceInterface
type FakeSilences struct {
	Fake *FakeNotificationV2beta2
}

var silencesResource = schema.GroupVersionResource{Group: "notification.kubesphere.io", Version: "v2beta2", Resource: "silences"}

var silencesKind = schema.GroupVersionKind{Group: "notification.kubesphere.io", Version: "v2beta2", Kind: "Silence"}

// Get takes name of the silence, and returns the corresponding silence object, and an error if there is any.
func (c *FakeSilences) Get(ctx context.Context, name string, options v1.GetOptions) (result *v2beta2.Silence, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(silencesResource, name), &v2beta2.Silence{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v2beta2.Silence), err
}

// List takes label and field selectors, and returns the list of Silences that match those selectors.
func (c *FakeSilences) List(ctx context.Context, opts v1.ListOptions) (result *v2beta2.SilenceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(silencesResource, silencesKind, opts), &v2beta2.SilenceList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v2beta2.SilenceList{ListMeta: obj.(*v2beta2.SilenceList).ListMeta}
	for _, item := range obj.(*v2beta2.SilenceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested silences.
func (c *FakeSilences) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(silencesResource, opts))
}

// Create takes the representation of a silence and creates it.  Returns the server's representation of the silence, and an error, if there is any.
func (c *FakeSilences) Create(ctx context.Context, silence *v2beta2.Silence, opts v1.CreateOptions) (result *v2beta2.Silence, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(silencesResource, silence), &v2beta2.Silence{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v2beta2.Silence), err
}

// Update takes the representation of a silence and updates it. Returns the server's representation of the silence, and an error, if there is any.
func (c *FakeSilences) Update(ctx context.Context, silence *v2beta2.Silence, opts v1.UpdateOptions) (result *v2beta2.Silence, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(silencesResource, silence), &v2beta2.Silence{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v2beta2.Silence), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeSilences) UpdateStatus(ctx context.Context, silence *v2beta2.Silence, opts v1.UpdateOptions) (*v2beta2.Silence, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(silencesResource, "status", silence), &v2beta2.Silence{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v2beta2.Silence), err
}

// Delete takes name of the silence and deletes it. Returns an error if one occurs.
func (c *FakeSilences) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(silencesResource, name, opts), &v2beta2.Silence{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSilences) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(silencesResource, listOpts)

	_, err := c.Fake.Invokes(action, &v2beta2.SilenceList{})
	return err
}

// Patch applies the patch and returns the patched silence.
func (c *FakeSilences) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v2beta2.Silence, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(silencesResource, name, pt, data, subresources...), &v2beta2.Silence{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v2beta2.Silence), err
}
