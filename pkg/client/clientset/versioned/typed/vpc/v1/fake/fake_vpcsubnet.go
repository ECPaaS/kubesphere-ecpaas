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
	vpcv1 "kubesphere.io/api/vpc/v1"
)

// FakeVPCSubnets implements VPCSubnetInterface
type FakeVPCSubnets struct {
	Fake *FakeK8sV1
	ns   string
}

var vpcsubnetsResource = schema.GroupVersionResource{Group: "k8s.ovn.org", Version: "v1", Resource: "vpcsubnets"}

var vpcsubnetsKind = schema.GroupVersionKind{Group: "k8s.ovn.org", Version: "v1", Kind: "VPCSubnet"}

// Get takes name of the vPCSubnet, and returns the corresponding vPCSubnet object, and an error if there is any.
func (c *FakeVPCSubnets) Get(ctx context.Context, name string, options v1.GetOptions) (result *vpcv1.VPCSubnet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(vpcsubnetsResource, c.ns, name), &vpcv1.VPCSubnet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*vpcv1.VPCSubnet), err
}

// List takes label and field selectors, and returns the list of VPCSubnets that match those selectors.
func (c *FakeVPCSubnets) List(ctx context.Context, opts v1.ListOptions) (result *vpcv1.VPCSubnetList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(vpcsubnetsResource, vpcsubnetsKind, c.ns, opts), &vpcv1.VPCSubnetList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &vpcv1.VPCSubnetList{ListMeta: obj.(*vpcv1.VPCSubnetList).ListMeta}
	for _, item := range obj.(*vpcv1.VPCSubnetList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested vPCSubnets.
func (c *FakeVPCSubnets) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(vpcsubnetsResource, c.ns, opts))

}

// Create takes the representation of a vPCSubnet and creates it.  Returns the server's representation of the vPCSubnet, and an error, if there is any.
func (c *FakeVPCSubnets) Create(ctx context.Context, vPCSubnet *vpcv1.VPCSubnet, opts v1.CreateOptions) (result *vpcv1.VPCSubnet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(vpcsubnetsResource, c.ns, vPCSubnet), &vpcv1.VPCSubnet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*vpcv1.VPCSubnet), err
}

// Update takes the representation of a vPCSubnet and updates it. Returns the server's representation of the vPCSubnet, and an error, if there is any.
func (c *FakeVPCSubnets) Update(ctx context.Context, vPCSubnet *vpcv1.VPCSubnet, opts v1.UpdateOptions) (result *vpcv1.VPCSubnet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(vpcsubnetsResource, c.ns, vPCSubnet), &vpcv1.VPCSubnet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*vpcv1.VPCSubnet), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeVPCSubnets) UpdateStatus(ctx context.Context, vPCSubnet *vpcv1.VPCSubnet, opts v1.UpdateOptions) (*vpcv1.VPCSubnet, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(vpcsubnetsResource, "status", c.ns, vPCSubnet), &vpcv1.VPCSubnet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*vpcv1.VPCSubnet), err
}

// Delete takes name of the vPCSubnet and deletes it. Returns an error if one occurs.
func (c *FakeVPCSubnets) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(vpcsubnetsResource, c.ns, name), &vpcv1.VPCSubnet{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeVPCSubnets) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(vpcsubnetsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &vpcv1.VPCSubnetList{})
	return err
}

// Patch applies the patch and returns the patched vPCSubnet.
func (c *FakeVPCSubnets) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *vpcv1.VPCSubnet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(vpcsubnetsResource, c.ns, name, pt, data, subresources...), &vpcv1.VPCSubnet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*vpcv1.VPCSubnet), err
}
