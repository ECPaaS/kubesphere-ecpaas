/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com.tw
*/

package v1

import (
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	vpc "kubesphere.io/kubesphere/pkg/models/vpc"
)

const (
	GroupName       = "k8s.ovn.org"
	ExampleJsonPath = "./pkg/kapis/vpc/v1/definition/"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

func Resource(resource string) schema.GroupResource {
	return GroupVersion.WithResource(resource).GroupResource()
}

type BadRequestError struct {
	Reason string `json:"reason"`
}

func AddToContainer(container *restful.Container, factory informers.InformerFactory, k8sclient kubernetes.Interface, ksclient kubesphere.Interface) error {
	webservice := runtime.NewWebService(GroupVersion)
	handler := newHandler(factory, k8sclient, ksclient)

	webservice.Route(webservice.GET("/vpcnetworks").
		To(handler.ListVpcNetwork).
		Doc("List all vpcnetwork resources").
		Returns(http.StatusOK, api.StatusOK, []vpc.VPCNetworkResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcNetworkTag}))

	webservice.Route(webservice.GET("/vpcnetwork/{name}").
		To(handler.GetVpcNetwork).
		Param(webservice.PathParameter("name", "vpcnetwork name")).
		Doc("Get vpcnetwork resources").
		Returns(http.StatusOK, api.StatusOK, vpc.VPCNetworkResponse{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcNetworkTag}))

	webservice.Route(webservice.GET("/vpcnetwork/gatewayChassisNode").
		To(handler.GetGatewayChassisNode).
		Doc("List available gateway chassis nodes").
		Returns(http.StatusOK, api.StatusOK, []vpc.GatewayChassisNode{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcNetworkTag}))

	webservice.Route(webservice.POST("/vpcnetwork/{workspace}").
		To(handler.CreateVpcNetwork).
		Param(webservice.PathParameter("workspace", "workspace name")).
		Reads(vpc.VPCNetwork{}).
		Doc("Create vpcnetwork").
		Returns(http.StatusOK, api.StatusOK, vpc.VPCNetwork{}).
		Returns(http.StatusBadRequest, api.StatusBadRequest, BadRequestError{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcNetworkTag}))

	webservice.Route(webservice.PUT("/vpcnetwork/{name}").
		To(handler.UpdateVpcNetwork).
		Param(webservice.PathParameter("name", "vpcnetwork name")).
		Reads(vpc.VPCNetworkBase{}).
		Doc("Update vpcnetwork").
		Returns(http.StatusOK, api.StatusOK, vpc.VPCNetwork{}).
		Returns(http.StatusBadRequest, api.StatusBadRequest, BadRequestError{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcNetworkTag}))

	webservice.Route(webservice.PATCH("/vpcnetwork/{name}").
		To(handler.PatchVpcNetwork).
		Param(webservice.PathParameter("name", "vpcnetwork name")).
		Reads(vpc.VPCNetworkPatch{}).
		Doc("Patch vpcnetwork").
		Returns(http.StatusOK, api.StatusOK, vpc.VPCNetworkPatch{}).
		Returns(http.StatusBadRequest, api.StatusBadRequest, BadRequestError{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcNetworkTag}))

	webservice.Route(webservice.DELETE("/vpcnetwork/{name}").
		To(handler.DeleteVpcNetwork).
		Param(webservice.PathParameter("name", "vpcnetwork name")).
		Doc("Delete vpcnetwork").
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcNetworkTag}))

	// VPC Subnet
	webservice.Route(webservice.GET("/vpcsubnets").
		To(handler.ListVpcSubnet).
		Doc("List all vpcsubnet resources").
		Returns(http.StatusOK, api.StatusOK, []vpc.VPCSubnetResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcSubnetTag}))

	webservice.Route(webservice.GET("/vpcsubnets/vpcnetwork/{name}").
		To(handler.ListVpcSubnetWithinVpcNetwork).
		Param(webservice.PathParameter("name", "vpcnetwork name")).
		Doc("List all vpcsubnet resource within vpcnetwork").
		Returns(http.StatusOK, api.StatusOK, []vpc.VPCSubnetResponse{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcSubnetTag}))

	webservice.Route(webservice.GET("/vpcsubnets/{namespace}").
		To(handler.ListVpcSubnetWithinNamespace).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Doc("List all vpcsubnet within the same namespace.").
		Returns(http.StatusOK, api.StatusOK, []vpc.VPCSubnetResponse{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcSubnetTag}))

	webservice.Route(webservice.GET("/vpcsubnet/{namespace}/{name}").
		To(handler.GetVpcSubnet).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("name", "vpcsubnet name")).
		Doc("Get vpcsubnet resources").
		Returns(http.StatusOK, api.StatusOK, vpc.VPCSubnetResponse{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcSubnetTag}))

	webservice.Route(webservice.POST("/vpcsubnet").
		To(handler.CreateVpcSubnet).
		Reads(vpc.VPCSubnet{}).
		Doc("Create vpcsubnet").
		Returns(http.StatusOK, api.StatusOK, vpc.VPCSubnet{}).
		Returns(http.StatusBadRequest, api.StatusBadRequest, BadRequestError{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcSubnetTag}))

	webservice.Route(webservice.PUT("/vpcsubnet/{namespace}/{name}").
		To(handler.UpdateVpcSubnet).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("name", "vpcsubnet name")).
		Reads(vpc.VPCSubnetBase{}).
		Doc("Update vpcsubnet").
		Returns(http.StatusOK, api.StatusOK, vpc.VPCSubnet{}).
		Returns(http.StatusBadRequest, api.StatusBadRequest, BadRequestError{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcSubnetTag}))

	webservice.Route(webservice.PATCH("/vpcsubnet/{namespace}/{name}").
		To(handler.PatchVpcSubnet).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("name", "vpcsubnet name")).
		Reads(vpc.VPCSubnetPatch{}).
		Doc("Patch vpcsubnet").
		Returns(http.StatusOK, api.StatusOK, vpc.VPCSubnetPatch{}).
		Returns(http.StatusBadRequest, api.StatusBadRequest, BadRequestError{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcSubnetTag}))

	webservice.Route(webservice.DELETE("/vpcsubnet/{namespace}/{name}").
		To(handler.DeleteVpcSubnet).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("name", "vpcsubnet name")).
		Doc("Delete vpcsubnet").
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcSubnetTag}))

	container.Add(webservice)

	return nil
}
