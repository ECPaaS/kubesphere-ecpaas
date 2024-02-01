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
	v1beta2 "kubesphere.io/api/types/v1beta2"
)

// FakeFederatedNotificationReceivers implements FederatedNotificationReceiverInterface
type FakeFederatedNotificationReceivers struct {
	Fake *FakeTypesV1beta2
}

var federatednotificationreceiversResource = schema.GroupVersionResource{Group: "types.kubefed.io", Version: "v1beta2", Resource: "federatednotificationreceivers"}

var federatednotificationreceiversKind = schema.GroupVersionKind{Group: "types.kubefed.io", Version: "v1beta2", Kind: "FederatedNotificationReceiver"}

// Get takes name of the federatedNotificationReceiver, and returns the corresponding federatedNotificationReceiver object, and an error if there is any.
func (c *FakeFederatedNotificationReceivers) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta2.FederatedNotificationReceiver, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(federatednotificationreceiversResource, name), &v1beta2.FederatedNotificationReceiver{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta2.FederatedNotificationReceiver), err
}

// List takes label and field selectors, and returns the list of FederatedNotificationReceivers that match those selectors.
func (c *FakeFederatedNotificationReceivers) List(ctx context.Context, opts v1.ListOptions) (result *v1beta2.FederatedNotificationReceiverList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(federatednotificationreceiversResource, federatednotificationreceiversKind, opts), &v1beta2.FederatedNotificationReceiverList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta2.FederatedNotificationReceiverList{ListMeta: obj.(*v1beta2.FederatedNotificationReceiverList).ListMeta}
	for _, item := range obj.(*v1beta2.FederatedNotificationReceiverList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested federatedNotificationReceivers.
func (c *FakeFederatedNotificationReceivers) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(federatednotificationreceiversResource, opts))
}

// Create takes the representation of a federatedNotificationReceiver and creates it.  Returns the server's representation of the federatedNotificationReceiver, and an error, if there is any.
func (c *FakeFederatedNotificationReceivers) Create(ctx context.Context, federatedNotificationReceiver *v1beta2.FederatedNotificationReceiver, opts v1.CreateOptions) (result *v1beta2.FederatedNotificationReceiver, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(federatednotificationreceiversResource, federatedNotificationReceiver), &v1beta2.FederatedNotificationReceiver{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta2.FederatedNotificationReceiver), err
}

// Update takes the representation of a federatedNotificationReceiver and updates it. Returns the server's representation of the federatedNotificationReceiver, and an error, if there is any.
func (c *FakeFederatedNotificationReceivers) Update(ctx context.Context, federatedNotificationReceiver *v1beta2.FederatedNotificationReceiver, opts v1.UpdateOptions) (result *v1beta2.FederatedNotificationReceiver, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(federatednotificationreceiversResource, federatedNotificationReceiver), &v1beta2.FederatedNotificationReceiver{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta2.FederatedNotificationReceiver), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeFederatedNotificationReceivers) UpdateStatus(ctx context.Context, federatedNotificationReceiver *v1beta2.FederatedNotificationReceiver, opts v1.UpdateOptions) (*v1beta2.FederatedNotificationReceiver, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(federatednotificationreceiversResource, "status", federatedNotificationReceiver), &v1beta2.FederatedNotificationReceiver{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta2.FederatedNotificationReceiver), err
}

// Delete takes name of the federatedNotificationReceiver and deletes it. Returns an error if one occurs.
func (c *FakeFederatedNotificationReceivers) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(federatednotificationreceiversResource, name, opts), &v1beta2.FederatedNotificationReceiver{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeFederatedNotificationReceivers) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(federatednotificationreceiversResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1beta2.FederatedNotificationReceiverList{})
	return err
}

// Patch applies the patch and returns the patched federatedNotificationReceiver.
func (c *FakeFederatedNotificationReceivers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta2.FederatedNotificationReceiver, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(federatednotificationreceiversResource, name, pt, data, subresources...), &v1beta2.FederatedNotificationReceiver{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta2.FederatedNotificationReceiver), err
}
