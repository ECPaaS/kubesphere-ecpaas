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
	v1alpha1 "kubesphere.io/api/virtualization/v1alpha1"
)

// FakeDiskVolumes implements DiskVolumeInterface
type FakeDiskVolumes struct {
	Fake *FakeVirtualizationV1alpha1
	ns   string
}

var diskvolumesResource = schema.GroupVersionResource{Group: "virtualization.ecpaas.io", Version: "v1alpha1", Resource: "diskvolumes"}

var diskvolumesKind = schema.GroupVersionKind{Group: "virtualization.ecpaas.io", Version: "v1alpha1", Kind: "DiskVolume"}

// Get takes name of the diskVolume, and returns the corresponding diskVolume object, and an error if there is any.
func (c *FakeDiskVolumes) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.DiskVolume, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(diskvolumesResource, c.ns, name), &v1alpha1.DiskVolume{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DiskVolume), err
}

// List takes label and field selectors, and returns the list of DiskVolumes that match those selectors.
func (c *FakeDiskVolumes) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.DiskVolumeList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(diskvolumesResource, diskvolumesKind, c.ns, opts), &v1alpha1.DiskVolumeList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.DiskVolumeList{ListMeta: obj.(*v1alpha1.DiskVolumeList).ListMeta}
	for _, item := range obj.(*v1alpha1.DiskVolumeList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested diskVolumes.
func (c *FakeDiskVolumes) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(diskvolumesResource, c.ns, opts))

}

// Create takes the representation of a diskVolume and creates it.  Returns the server's representation of the diskVolume, and an error, if there is any.
func (c *FakeDiskVolumes) Create(ctx context.Context, diskVolume *v1alpha1.DiskVolume, opts v1.CreateOptions) (result *v1alpha1.DiskVolume, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(diskvolumesResource, c.ns, diskVolume), &v1alpha1.DiskVolume{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DiskVolume), err
}

// Update takes the representation of a diskVolume and updates it. Returns the server's representation of the diskVolume, and an error, if there is any.
func (c *FakeDiskVolumes) Update(ctx context.Context, diskVolume *v1alpha1.DiskVolume, opts v1.UpdateOptions) (result *v1alpha1.DiskVolume, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(diskvolumesResource, c.ns, diskVolume), &v1alpha1.DiskVolume{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DiskVolume), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeDiskVolumes) UpdateStatus(ctx context.Context, diskVolume *v1alpha1.DiskVolume, opts v1.UpdateOptions) (*v1alpha1.DiskVolume, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(diskvolumesResource, "status", c.ns, diskVolume), &v1alpha1.DiskVolume{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DiskVolume), err
}

// Delete takes name of the diskVolume and deletes it. Returns an error if one occurs.
func (c *FakeDiskVolumes) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(diskvolumesResource, c.ns, name), &v1alpha1.DiskVolume{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeDiskVolumes) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(diskvolumesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.DiskVolumeList{})
	return err
}

// Patch applies the patch and returns the patched diskVolume.
func (c *FakeDiskVolumes) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.DiskVolume, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(diskvolumesResource, c.ns, name, pt, data, subresources...), &v1alpha1.DiskVolume{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DiskVolume), err
}
