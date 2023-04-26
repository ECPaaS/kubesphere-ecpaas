/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com.tw
*/

package vpcnetwork

import (
	"k8s.io/apimachinery/pkg/runtime"

	vpcv1 "kubesphere.io/api/vpc/v1"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type vpcnetworkGetter struct {
	sharedInformers informers.SharedInformerFactory
}

func New(sharedInformers informers.SharedInformerFactory) v1alpha3.Interface {
	return &vpcnetworkGetter{sharedInformers: sharedInformers}
}

func (d *vpcnetworkGetter) Get(_, name string) (runtime.Object, error) {
	return d.sharedInformers.K8s().V1().VPCNetworks().Lister().Get(name)
}

func (d *vpcnetworkGetter) List(_ string, query *query.Query) (*api.ListResult, error) {

	vpcnetworks, err := d.sharedInformers.K8s().V1().VPCNetworks().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, vpcnetwork := range vpcnetworks {
		result = append(result, vpcnetwork)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *vpcnetworkGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftVpcnetwork, ok := left.(*vpcv1.VPCNetwork)
	if !ok {
		return false
	}

	rightVpcnetwork, ok := right.(*vpcv1.VPCNetwork)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftVpcnetwork.ObjectMeta, rightVpcnetwork.ObjectMeta, field)
}

func (d *vpcnetworkGetter) filter(object runtime.Object, filter query.Filter) bool {
	role, ok := object.(*vpcv1.VPCNetwork)

	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(role.ObjectMeta, filter)
}
