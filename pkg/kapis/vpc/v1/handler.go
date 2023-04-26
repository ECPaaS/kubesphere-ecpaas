/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com.tw
*/

package v1

import (
	"fmt"
	"net"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
	vpcv1 "kubesphere.io/api/vpc/v1"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/informers"

	"kubesphere.io/kubesphere/pkg/models/vpc"
	servererr "kubesphere.io/kubesphere/pkg/server/errors"

	// vpclister "kubesphere.io/kubesphere/pkg/client/listers/vpc/v1"
	"k8s.io/client-go/kubernetes"
)

type handler struct {
	vpc vpc.Interface
	// vpcLister vpclister.VPCNetworkLister
}

func newHandler(factory informers.InformerFactory, k8sclient kubernetes.Interface, ksclient kubesphere.Interface) *handler {
	return &handler{
		vpc: vpc.New(factory, k8sclient, ksclient),
	}
}

func (h *handler) ListVpcNetwork(request *restful.Request, response *restful.Response) {

	queryParam := query.ParseQueryParameter(request)
	vpcnetworks, err := h.vpc.ListVpcNetwork(queryParam)

	if err != nil {
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		} else {
			api.HandleInternalError(response, request, err)
			return
		}
	}

	response.WriteAsJson(vpcnetworks)
}

func (h *handler) GetVpcNetwork(request *restful.Request, response *restful.Response) {

	vpcnetwork := request.PathParameter("vpcnetwork")
	vpc, err := h.vpc.GetVpcNetwork(vpcnetwork)

	if err != nil {
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		} else {
			api.HandleInternalError(response, request, err)
			return
		}
	}

	response.WriteAsJson(vpc)
}

func (h *handler) CreateVpcNetwork(request *restful.Request, response *restful.Response) {
	workspaceName := request.PathParameter("workspace")
	var vpcnetwork vpcv1.VPCNetwork

	err := request.ReadEntity(&vpcnetwork)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	err = validationVPCNetwork(vpcnetwork)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.vpc.CreateVpcNetwork(workspaceName, &vpcnetwork)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		if errors.IsForbidden(err) {
			api.HandleForbidden(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(created)
}

func validationVPCNetwork(vpcnetwork vpcv1.VPCNetwork) error {

	_, _, err := net.ParseCIDR(vpcnetwork.Spec.CIDR)

	if vpcnetwork.Spec.SubnetLength < 0 || vpcnetwork.Spec.SubnetLength > 32 {
		err = errors.NewBadRequest("invalid subnet length, should be 0-32")
	}
	return err
}

func (h *handler) UpdateVpcNetwork(request *restful.Request, response *restful.Response) {
	workspaceName := request.PathParameter("workspace")
	vpcnetworkName := request.PathParameter("vpcnetwork")
	var vpcnetwork vpcv1.VPCNetwork

	err := request.ReadEntity(&vpcnetwork)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	err = validationVPCNetwork(vpcnetwork)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	if vpcnetworkName != vpcnetwork.Name {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", vpcnetwork.Name, vpcnetworkName)
		klog.Errorf("%+v", err)
		api.HandleBadRequest(response, request, err)
		return
	}

	updated, err := h.vpc.UpdateVpcNetwork(workspaceName, &vpcnetwork)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		if errors.IsForbidden(err) {
			api.HandleForbidden(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(updated)
}

func (h *handler) PatchVpcNetwork(request *restful.Request, response *restful.Response) {
	vpcnetworkName := request.PathParameter("vpcnetwork")

	var vpcnetwork vpcv1.VPCNetwork
	err := request.ReadEntity(&vpcnetwork)
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	err = validationVPCNetwork(vpcnetwork)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	patched, err := h.vpc.PatchVpcNetwork(vpcnetworkName, &vpcnetwork)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		if errors.IsForbidden(err) {
			api.HandleForbidden(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(patched)
}

func (h *handler) DeleteVpcNetwork(request *restful.Request, response *restful.Response) {
	vpcnetwork := request.PathParameter("vpcnetwork")

	err := h.vpc.DeleteVpcNetwork(vpcnetwork)

	if err != nil {
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *handler) ListVpcSubnet(request *restful.Request, response *restful.Response) {

	queryParam := query.ParseQueryParameter(request)
	vpcsubnets, err := h.vpc.ListVpcSubnet(queryParam)

	if err != nil {
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		} else {
			api.HandleInternalError(response, request, err)
			return
		}
	}

	response.WriteAsJson(vpcsubnets)
}

func (h *handler) GetVpcSubnet(request *restful.Request, response *restful.Response) {

	vpcsubnet := request.PathParameter("vpcsubnet")
	namespace := request.PathParameter("namespace")
	vpc, err := h.vpc.GetVpcSubnet(namespace, vpcsubnet)

	if err != nil {
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		} else {
			api.HandleInternalError(response, request, err)
			return
		}
	}

	response.WriteAsJson(vpc)
}

func (h *handler) CreateVpcSubnet(request *restful.Request, response *restful.Response) {

	var vpcsubnet vpcv1.VPCSubnet

	err := request.ReadEntity(&vpcsubnet)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	err = validationVPCSubnet(vpcsubnet)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.vpc.CreateVpcSubnet(&vpcsubnet)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		if errors.IsForbidden(err) {
			api.HandleForbidden(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(created)
}

func validationVPCSubnet(vpcsubnet vpcv1.VPCSubnet) error {

	_, _, err := net.ParseCIDR(vpcsubnet.Spec.CIDR)

	return err
}

func (h *handler) UpdateVpcSubnet(request *restful.Request, response *restful.Response) {

	vpcsubnetName := request.PathParameter("vpcsubnet")
	var vpcsubnet vpcv1.VPCSubnet

	err := request.ReadEntity(&vpcsubnet)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	err = validationVPCSubnet(vpcsubnet)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	if vpcsubnetName != vpcsubnet.Name {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", vpcsubnet.Name, vpcsubnetName)
		klog.Errorf("%+v", err)
		api.HandleBadRequest(response, request, err)
		return
	}

	updated, err := h.vpc.UpdateVpcSubnet(&vpcsubnet)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		if errors.IsForbidden(err) {
			api.HandleForbidden(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(updated)
}

func (h *handler) PatchVpcSubnet(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	vpcsubnetName := request.PathParameter("vpcsubnet")

	var vpcsubnet vpcv1.VPCSubnet
	err := request.ReadEntity(&vpcsubnet)
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	err = validationVPCSubnet(vpcsubnet)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	patched, err := h.vpc.PatchVpcSubnet(namespace, vpcsubnetName, &vpcsubnet)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		if errors.IsForbidden(err) {
			api.HandleForbidden(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(patched)
}

func (h *handler) DeleteVpcSubnet(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	vpcsubnet := request.PathParameter("vpcsubnet")

	err := h.vpc.DeleteVpcSubnet(namespace, vpcsubnet)

	if err != nil {
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}
