/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com.tw
*/

package v1

import (
	"net"
	"net/http"
	"reflect"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/kapis/validation"
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

	vpcnetwork := request.PathParameter("name")
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

func (h *handler) GetGatewayChassisNode(request *restful.Request, response *restful.Response) {

	chassisNode, err := h.vpc.GetGatewayChassisNode()

	if err != nil {
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		} else {
			api.HandleInternalError(response, request, err)
			return
		}
	}

	response.WriteAsJson(chassisNode)
}

func (h *handler) CreateVpcNetwork(request *restful.Request, response *restful.Response) {
	workspaceName := request.PathParameter("workspace")
	var vpcnetwork vpc.VPCNetwork

	err := request.ReadEntity(&vpcnetwork)

	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	if !validationVPCNetwork(vpcnetwork, response) {
		return
	}

	created, err := h.vpc.CreateVpcNetwork(workspaceName, &vpcnetwork)

	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	response.WriteEntity(created)
}

func (h *handler) UpdateVpcNetwork(request *restful.Request, response *restful.Response) {
	vpcnetworkName := request.PathParameter("name")
	var vpcnetwork vpc.VPCNetworkBase

	err := request.ReadEntity(&vpcnetwork)

	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	if !validationVPCNetworkBase(vpcnetwork, response) {
		return
	}

	updated, err := h.vpc.UpdateVpcNetwork(&vpcnetwork, vpcnetworkName)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(updated)
}

func (h *handler) PatchVpcNetwork(request *restful.Request, response *restful.Response) {
	vpcnetworkName := request.PathParameter("name")

	var vpcnetwork vpc.VPCNetworkPatch
	err := request.ReadEntity(&vpcnetwork)
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	if !validationVPCNetworkPatch(vpcnetwork, response) {
		return
	}

	patched, err := h.vpc.PatchVpcNetwork(vpcnetworkName, &vpcnetwork)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(patched)
}

func (h *handler) DeleteVpcNetwork(request *restful.Request, response *restful.Response) {
	vpcnetwork := request.PathParameter("name")

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

// VPC Subnet
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

func (h *handler) ListVpcSubnetWithinVpcNetwork(request *restful.Request, response *restful.Response) {

	vpcnetwork := request.PathParameter("name")
	queryParam := query.ParseQueryParameter(request)
	vpcsubnets, err := h.vpc.ListVpcSubnetWithinVpcNetwork(vpcnetwork, queryParam)

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

func (h *handler) ListVpcSubnetWithinNamespace(request *restful.Request, response *restful.Response) {

	namespace := request.PathParameter("namespace")
	queryParam := query.ParseQueryParameter(request)
	vpcsubnets, err := h.vpc.ListVpcSubnetWithinNamespace(namespace, queryParam)

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

	vpcsubnetName := request.PathParameter("name")
	namespace := request.PathParameter("namespace")
	vpc, err := h.vpc.GetVpcSubnet(namespace, vpcsubnetName)

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

	var vpcsubnet vpc.VPCSubnet

	err := request.ReadEntity(&vpcsubnet)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	if !validation.IsValidCIDR(vpcsubnet.CIDR, response) {
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

func (h *handler) UpdateVpcSubnet(request *restful.Request, response *restful.Response) {

	vpcsubnetName := request.PathParameter("name")
	namespace := request.PathParameter("namespace")
	var vpcsubnet vpc.VPCSubnetBase

	err := request.ReadEntity(&vpcsubnet)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	if !validation.IsValidCIDR(vpcsubnet.CIDR, response) {
		return
	}

	updated, err := h.vpc.UpdateVpcSubnet(&vpcsubnet, namespace, vpcsubnetName)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(updated)
}

func (h *handler) PatchVpcSubnet(request *restful.Request, response *restful.Response) {
	vpcsubnetName := request.PathParameter("name")
	namespace := request.PathParameter("namespace")
	var vpcsubnet vpc.VPCSubnetPatch
	err := request.ReadEntity(&vpcsubnet)
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	if vpcsubnet.CIDR != "" {
		if !validateCIDR(vpcsubnet.CIDR, response) {
			return
		}
	}

	patched, err := h.vpc.PatchVpcSubnet(&vpcsubnet, namespace, vpcsubnetName)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(patched)
}

func (h *handler) DeleteVpcSubnet(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("name")

	err := h.vpc.DeleteVpcSubnet(namespace, name)

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

func validationVPCNetwork(vpcnetwork vpc.VPCNetwork, resp *restful.Response) bool {

	// name
	if !validation.IsValidString(vpcnetwork.Name, resp) {
		return false
	}

	if !validationVPCNetworkBase(vpcnetwork.VPCNetworkBase, resp) {
		return false
	}

	return true
}

func validationVPCNetworkPatch(vpcnetwork vpc.VPCNetworkPatch, resp *restful.Response) bool {
	reflectType := reflect.TypeOf(vpcnetwork)

	if vpcnetwork.CIDR != "" {
		if !validateCIDR(vpcnetwork.CIDR, resp) {
			return false
		}
	}

	if len(vpcnetwork.GatewayChassis) > 0 {
		if !validateGatewayChassis(vpcnetwork.GatewayChassis, resp) {
			return false
		}
	}

	if len(vpcnetwork.L3Gateways) > 0 {
		if !validateL3Gateway(vpcnetwork.L3Gateways, reflectType, resp) {
			return false
		}
	}

	return true
}

func validationVPCNetworkBase(vpcnetwork vpc.VPCNetworkBase, resp *restful.Response) bool {
	reflectType := reflect.TypeOf(vpcnetwork)

	if !validateCIDR(vpcnetwork.CIDR, resp) {
		return false
	}

	if !validateGatewayChassis(vpcnetwork.GatewayChassis, resp) {
		return false
	}

	if !validateL3Gateway(vpcnetwork.L3Gateways, reflectType, resp) {
		return false
	}

	return true
}

func validateCIDR(cidr string, resp *restful.Response) bool {
	// CIDR
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, validation.BadRequestError{
			Reason: "invalid CIDR: " + err.Error(),
		})
		return false
	}
	return true
}

func validateGatewayChassis(gatewayChassises []vpc.GatewayChassis, resp *restful.Response) bool {
	// gatewayChassis
	for _, gatewayChassis := range gatewayChassises {
		ip := net.ParseIP(gatewayChassis.IP)
		if ip == nil {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, validation.BadRequestError{
				Reason: "invalid IP Address",
			})
			return false
		}
	}
	return true
}

func validateL3Gateway(l3gateways []vpc.L3Gateway, reflectType reflect.Type, resp *restful.Response) bool {
	// L3Gateways
	for _, gateway := range l3gateways {
		// Destination
		if gateway.Destination != "" {
			_, _, err := net.ParseCIDR(gateway.Destination)
			if err != nil {
				resp.WriteHeaderAndEntity(http.StatusBadRequest, validation.BadRequestError{
					Reason: "invalid destination address: " + err.Error(),
				})
				return false
			}
		}
		// Network
		_, _, err := net.ParseCIDR(gateway.Network)
		if err != nil {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, validation.BadRequestError{
				Reason: "invalid network address: " + err.Error(),
			})
			return false
		}
		// Nexthop
		ip := net.ParseIP(gateway.NextHop)
		if ip == nil {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, validation.BadRequestError{
				Reason: "invalid Nexthop IP Address",
			})
			return false
		}
		// VLAN
		if gateway.VLANId != 0 {
			if !validation.IsValidWithinRange(reflectType, int(gateway.VLANId), "VLANId", resp) {
				return false
			}
		}
	}
	return true
}
