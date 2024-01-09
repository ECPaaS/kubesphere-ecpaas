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

// FakeFederatedConfigMaps implements FederatedConfigMapInterface
type FakeFederatedConfigMaps struct {
	Fake *FakeTypesV1beta1
	ns   string
}

var federatedconfigmapsResource = schema.GroupVersionResource{Group: "types.kubefed.io", Version: "v1beta1", Resource: "federatedconfigmaps"}

var federatedconfigmapsKind = schema.GroupVersionKind{Group: "types.kubefed.io", Version: "v1beta1", Kind: "FederatedConfigMap"}

// Get takes name of the federatedConfigMap, and returns the corresponding federatedConfigMap object, and an error if there is any.
func (c *FakeFederatedConfigMaps) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.FederatedConfigMap, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(federatedconfigmapsResource, c.ns, name), &v1beta1.FederatedConfigMap{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.FederatedConfigMap), err
}

// List takes label and field selectors, and returns the list of FederatedConfigMaps that match those selectors.
func (c *FakeFederatedConfigMaps) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.FederatedConfigMapList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(federatedconfigmapsResource, federatedconfigmapsKind, c.ns, opts), &v1beta1.FederatedConfigMapList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.FederatedConfigMapList{ListMeta: obj.(*v1beta1.FederatedConfigMapList).ListMeta}
	for _, item := range obj.(*v1beta1.FederatedConfigMapList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested federatedConfigMaps.
func (c *FakeFederatedConfigMaps) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(federatedconfigmapsResource, c.ns, opts))

}

// Create takes the representation of a federatedConfigMap and creates it.  Returns the server's representation of the federatedConfigMap, and an error, if there is any.
func (c *FakeFederatedConfigMaps) Create(ctx context.Context, federatedConfigMap *v1beta1.FederatedConfigMap, opts v1.CreateOptions) (result *v1beta1.FederatedConfigMap, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(federatedconfigmapsResource, c.ns, federatedConfigMap), &v1beta1.FederatedConfigMap{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.FederatedConfigMap), err
}

// Update takes the representation of a federatedConfigMap and updates it. Returns the server's representation of the federatedConfigMap, and an error, if there is any.
func (c *FakeFederatedConfigMaps) Update(ctx context.Context, federatedConfigMap *v1beta1.FederatedConfigMap, opts v1.UpdateOptions) (result *v1beta1.FederatedConfigMap, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(federatedconfigmapsResource, c.ns, federatedConfigMap), &v1beta1.FederatedConfigMap{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.FederatedConfigMap), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeFederatedConfigMaps) UpdateStatus(ctx context.Context, federatedConfigMap *v1beta1.FederatedConfigMap, opts v1.UpdateOptions) (*v1beta1.FederatedConfigMap, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(federatedconfigmapsResource, "status", c.ns, federatedConfigMap), &v1beta1.FederatedConfigMap{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.FederatedConfigMap), err
}

// Delete takes name of the federatedConfigMap and deletes it. Returns an error if one occurs.
func (c *FakeFederatedConfigMaps) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(federatedconfigmapsResource, c.ns, name, opts), &v1beta1.FederatedConfigMap{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeFederatedConfigMaps) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(federatedconfigmapsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1beta1.FederatedConfigMapList{})
	return err
}

// Patch applies the patch and returns the patched federatedConfigMap.
func (c *FakeFederatedConfigMaps) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.FederatedConfigMap, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(federatedconfigmapsResource, c.ns, name, pt, data, subresources...), &v1beta1.FederatedConfigMap{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.FederatedConfigMap), err
}
