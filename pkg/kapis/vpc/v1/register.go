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
		Returns(http.StatusOK, api.StatusOK, []vpc.VPCNetwork{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcNetworkTag}))

	webservice.Route(webservice.GET("/vpcnetwork/{vpcnetwork}").
		To(handler.GetVpcNetwork).
		Param(webservice.PathParameter("vpcnetwork", "vpcnetwork name")).
		Doc("Get vpcnetwork resources").
		Returns(http.StatusOK, api.StatusOK, vpc.VPCNetworkBase{}).
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
		Returns(http.StatusBadRequest, "Invalid format", BadRequestError{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcNetworkTag}))

	webservice.Route(webservice.PUT("/vpcnetwork/{workspace}/{vpcnetwork}").
		To(handler.UpdateVpcNetwork).
		Param(webservice.PathParameter("workspace", "workspace name")).
		Param(webservice.PathParameter("vpcnetwork", "vpcnetwork name")).
		Reads(vpc.VPCNetworkBase{}).
		Doc("Update vpcnetwork").
		Returns(http.StatusOK, api.StatusOK, vpc.VPCNetworkBase{}).
		Returns(http.StatusBadRequest, "Invalid format", BadRequestError{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcNetworkTag}))

	webservice.Route(webservice.PATCH("/vpcnetwork/{vpcnetwork}").
		To(handler.PatchVpcNetwork).
		Param(webservice.PathParameter("vpcnetwork", "vpcnetwork name")).
		Reads(vpc.VPCNetworkBase{}).
		Doc("Patch vpcnetwork").
		Returns(http.StatusOK, api.StatusOK, vpc.VPCNetworkBase{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcNetworkTag}))

	webservice.Route(webservice.DELETE("/vpcnetwork/{vpcnetwork}").
		To(handler.DeleteVpcNetwork).
		Param(webservice.PathParameter("vpcnetwork", "vpcnetwork name")).
		Doc("Delete vpcnetwork").
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcNetworkTag}))

	// VPC Subnet
	webservice.Route(webservice.GET("/vpcsubnets").
		To(handler.ListVpcSubnet).
		Doc("List all vpcsubnet resources").
		Returns(http.StatusOK, api.StatusOK, []vpc.VPCSubnet{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcSubnetTag}))

	webservice.Route(webservice.GET("/vpcnetwork/{vpcnetwork}/vpcsubnets").
		To(handler.ListVpcSubnetWithinVpcNetwork).
		Param(webservice.PathParameter("vpcnetwork", "vpcnetwork name")).
		Doc("List all vpcsubnet resource within vpcnetwork").
		Returns(http.StatusOK, api.StatusOK, []vpc.VPCSubnet{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcSubnetTag}))

	webservice.Route(webservice.GET("/vpcsubnet/{namespace}/{vpcsubnet}").
		To(handler.GetVpcSubnet).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("vpcsubnet", "vpcsubnet name")).
		Doc("Get vpcsubnet resources").
		Returns(http.StatusOK, api.StatusOK, vpc.VPCSubnetBase{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcSubnetTag}))

	webservice.Route(webservice.POST("/vpcsubnet").
		To(handler.CreateVpcSubnet).
		Reads(vpc.VPCSubnet{}).
		Doc("Create vpcsubnet").
		Returns(http.StatusOK, api.StatusOK, vpc.VPCSubnet{}).
		Returns(http.StatusBadRequest, "Invalid format", BadRequestError{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcSubnetTag}))

	webservice.Route(webservice.PUT("/vpcsubnet/{namespace}/{vpcsubnet}").
		To(handler.UpdateVpcSubnet).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("vpcsubnet", "vpcsubnet name")).
		Reads(vpc.VPCSubnetBase{}).
		Doc("Update vpcsubnet").
		Returns(http.StatusOK, api.StatusOK, vpc.VPCSubnetBase{}).
		Returns(http.StatusBadRequest, "Invalid format", BadRequestError{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcSubnetTag}))

	webservice.Route(webservice.DELETE("/vpcsubnet/{namespace}/{vpcsubnet}").
		To(handler.DeleteVpcSubnet).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("vpcsubnet", "vpcsubnet name")).
		Doc("Delete vpcsubnet").
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VpcSubnetTag}))
	container.Add(webservice)

	return nil
}
