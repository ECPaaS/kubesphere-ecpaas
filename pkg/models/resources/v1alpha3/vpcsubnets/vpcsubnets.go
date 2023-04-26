/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com.tw
*/

package vpcsubnets

import (
	"k8s.io/apimachinery/pkg/runtime"

	vpcv1 "kubesphere.io/api/vpc/v1"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type vpcsubnetGetter struct {
	sharedInformers informers.SharedInformerFactory
}

func New(sharedInformers informers.SharedInformerFactory) v1alpha3.Interface {
	return &vpcsubnetGetter{sharedInformers: sharedInformers}
}

func (d *vpcsubnetGetter) Get(namespace, name string) (runtime.Object, error) {
	return d.sharedInformers.K8s().V1().VPCSubnets().Lister().VPCSubnets(namespace).Get(name)
}

func (d *vpcsubnetGetter) List(_ string, query *query.Query) (*api.ListResult, error) {

	vpcsubnets, err := d.sharedInformers.K8s().V1().VPCSubnets().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, vpcsubnet := range vpcsubnets {
		result = append(result, vpcsubnet)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *vpcsubnetGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftVpcsubnet, ok := left.(*vpcv1.VPCSubnet)
	if !ok {
		return false
	}

	rightVpcsubnet, ok := right.(*vpcv1.VPCSubnet)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftVpcsubnet.ObjectMeta, rightVpcsubnet.ObjectMeta, field)
}

func (d *vpcsubnetGetter) filter(object runtime.Object, filter query.Filter) bool {
	role, ok := object.(*vpcv1.VPCSubnet)

	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(role.ObjectMeta, filter)
}
