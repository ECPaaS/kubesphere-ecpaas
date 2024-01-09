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
	v1beta1 "kubesphere.io/api/types/v1beta1"
)

// FakeFederatedStatefulSets implements FederatedStatefulSetInterface
type FakeFederatedStatefulSets struct {
	Fake *FakeTypesV1beta1
	ns   string
}

var federatedstatefulsetsResource = schema.GroupVersionResource{Group: "types.kubefed.io", Version: "v1beta1", Resource: "federatedstatefulsets"}

var federatedstatefulsetsKind = schema.GroupVersionKind{Group: "types.kubefed.io", Version: "v1beta1", Kind: "FederatedStatefulSet"}

// Get takes name of the federatedStatefulSet, and returns the corresponding federatedStatefulSet object, and an error if there is any.
func (c *FakeFederatedStatefulSets) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.FederatedStatefulSet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(federatedstatefulsetsResource, c.ns, name), &v1beta1.FederatedStatefulSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.FederatedStatefulSet), err
}

// List takes label and field selectors, and returns the list of FederatedStatefulSets that match those selectors.
func (c *FakeFederatedStatefulSets) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.FederatedStatefulSetList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(federatedstatefulsetsResource, federatedstatefulsetsKind, c.ns, opts), &v1beta1.FederatedStatefulSetList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.FederatedStatefulSetList{ListMeta: obj.(*v1beta1.FederatedStatefulSetList).ListMeta}
	for _, item := range obj.(*v1beta1.FederatedStatefulSetList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested federatedStatefulSets.
func (c *FakeFederatedStatefulSets) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(federatedstatefulsetsResource, c.ns, opts))

}

// Create takes the representation of a federatedStatefulSet and creates it.  Returns the server's representation of the federatedStatefulSet, and an error, if there is any.
func (c *FakeFederatedStatefulSets) Create(ctx context.Context, federatedStatefulSet *v1beta1.FederatedStatefulSet, opts v1.CreateOptions) (result *v1beta1.FederatedStatefulSet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(federatedstatefulsetsResource, c.ns, federatedStatefulSet), &v1beta1.FederatedStatefulSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.FederatedStatefulSet), err
}

// Update takes the representation of a federatedStatefulSet and updates it. Returns the server's representation of the federatedStatefulSet, and an error, if there is any.
func (c *FakeFederatedStatefulSets) Update(ctx context.Context, federatedStatefulSet *v1beta1.FederatedStatefulSet, opts v1.UpdateOptions) (result *v1beta1.FederatedStatefulSet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(federatedstatefulsetsResource, c.ns, federatedStatefulSet), &v1beta1.FederatedStatefulSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.FederatedStatefulSet), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeFederatedStatefulSets) UpdateStatus(ctx context.Context, federatedStatefulSet *v1beta1.FederatedStatefulSet, opts v1.UpdateOptions) (*v1beta1.FederatedStatefulSet, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(federatedstatefulsetsResource, "status", c.ns, federatedStatefulSet), &v1beta1.FederatedStatefulSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.FederatedStatefulSet), err
}

// Delete takes name of the federatedStatefulSet and deletes it. Returns an error if one occurs.
func (c *FakeFederatedStatefulSets) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(federatedstatefulsetsResource, c.ns, name, opts), &v1beta1.FederatedStatefulSet{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeFederatedStatefulSets) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(federatedstatefulsetsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1beta1.FederatedStatefulSetList{})
	return err
}

// Patch applies the patch and returns the patched federatedStatefulSet.
func (c *FakeFederatedStatefulSets) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.FederatedStatefulSet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(federatedstatefulsetsResource, c.ns, name, pt, data, subresources...), &v1beta1.FederatedStatefulSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.FederatedStatefulSet), err
}
