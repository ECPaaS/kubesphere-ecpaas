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

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	v1 "kubesphere.io/api/vpc/v1"
)

// VPCSubnetLister helps list VPCSubnets.
// All objects returned here must be treated as read-only.
type VPCSubnetLister interface {
	// List lists all VPCSubnets in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.VPCSubnet, err error)
	// VPCSubnets returns an object that can list and get VPCSubnets.
	VPCSubnets(namespace string) VPCSubnetNamespaceLister
	VPCSubnetListerExpansion
}

// vPCSubnetLister implements the VPCSubnetLister interface.
type vPCSubnetLister struct {
	indexer cache.Indexer
}

// NewVPCSubnetLister returns a new VPCSubnetLister.
func NewVPCSubnetLister(indexer cache.Indexer) VPCSubnetLister {
	return &vPCSubnetLister{indexer: indexer}
}

// List lists all VPCSubnets in the indexer.
func (s *vPCSubnetLister) List(selector labels.Selector) (ret []*v1.VPCSubnet, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.VPCSubnet))
	})
	return ret, err
}

// VPCSubnets returns an object that can list and get VPCSubnets.
func (s *vPCSubnetLister) VPCSubnets(namespace string) VPCSubnetNamespaceLister {
	return vPCSubnetNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// VPCSubnetNamespaceLister helps list and get VPCSubnets.
// All objects returned here must be treated as read-only.
type VPCSubnetNamespaceLister interface {
	// List lists all VPCSubnets in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.VPCSubnet, err error)
	// Get retrieves the VPCSubnet from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.VPCSubnet, error)
	VPCSubnetNamespaceListerExpansion
}

// vPCSubnetNamespaceLister implements the VPCSubnetNamespaceLister
// interface.
type vPCSubnetNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all VPCSubnets in the indexer for a given namespace.
func (s vPCSubnetNamespaceLister) List(selector labels.Selector) (ret []*v1.VPCSubnet, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.VPCSubnet))
	})
	return ret, err
}

// Get retrieves the VPCSubnet from the indexer for a given namespace and name.
func (s vPCSubnetNamespaceLister) Get(name string) (*v1.VPCSubnet, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("vpcsubnet"), name)
	}
	return obj.(*v1.VPCSubnet), nil
}
