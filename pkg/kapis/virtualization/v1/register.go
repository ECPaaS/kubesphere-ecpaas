package virtualization

import (
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/server/errors"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	ui_virtz "kubesphere.io/kubesphere/pkg/models/virtualization"
)

const (
	GroupName = "virtualization.ecpaas.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

func AddToContainer(container *restful.Container, ksclient kubesphere.Interface) error {
	webservice := runtime.NewWebService(GroupVersion)
	handler := newHandler(ksclient)

	webservice.Route(webservice.POST("/namespace/{namespace}/virtualmachine").
		To(handler.CreateVirtualMahcine).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Reads(ui_virtz.VirtualMachineRequest{}).
		Doc("Create virtual machine").
		Returns(http.StatusOK, api.StatusOK, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMachineTag}))

	webservice.Route(webservice.PUT("/namespace/{namespace}/virtualmachine/{id}").
		To(handler.UpdateVirtualMahcine).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "virtual machine id")).
		Reads(ui_virtz.VirtualMachineRequest{}).
		Doc("Update virtual machine").
		Notes("Any parameters which are not provied will not be changed.").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.VirtualMachineRequest{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMachineTag}))

	webservice.Route(webservice.GET("/namespace/{namespace}/virtualmachine/{id}").
		To(handler.GetVirtualMachine).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "virtual machine id")).
		Doc("Get virtual machine").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.VirtualMachineResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMachineTag}))

	webservice.Route(webservice.GET("/namespace/{namespace}/virtualmachine").
		To(handler.ListVirtualMachineWithNamespace).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Doc("List all virtual machine with namespace").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.ListVirtualMachineResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMachineTag}))

	webservice.Route(webservice.GET("/virtualmachine").
		To(handler.ListVirtualMachine).
		Doc("List all virtual machine").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.ListVirtualMachineResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMachineTag}))

	webservice.Route(webservice.DELETE("/namespace/{namespace}/virtualmachine/{id}").
		To(handler.DeleteVirtualMachine).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "virtual machine id")).
		Doc("Delete virtual machine").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMachineTag}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}))

	webservice.Route(webservice.POST("/namespace/{namespace}/disk").
		To(handler.CreateDisk).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Reads(ui_virtz.DiskRequest{}).
		Doc("Create disk").
		Returns(http.StatusOK, api.StatusOK, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DiskTag}))

	webservice.Route(webservice.PUT("/namespace/{namespace}/disk/{id}").
		To(handler.UpdateDisk).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "disk id")).
		Reads(ui_virtz.DiskRequest{}).
		Doc("Update disk").
		Notes("Any parameters which are not provied will not be changed.").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.DiskRequest{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DiskTag}))

	webservice.Route(webservice.GET("/namespace/{namespace}/disk/{id}").
		To(handler.GetDisk).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "disk id")).
		Doc("Get disk").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.DiskResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DiskTag}))

	webservice.Route(webservice.GET("/namespace/{namespace}/disk").
		To(handler.ListDiskWithNamespace).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Doc("List all disk with namespace").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.ListDiskResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DiskTag}))

	webservice.Route(webservice.GET("/disk").
		To(handler.ListDisk).
		Doc("List all disk").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.ListDiskResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DiskTag}))

	webservice.Route(webservice.DELETE("/namespace/{namespace}/disk/{id}").
		To(handler.DeleteDisk).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "disk id")).
		Doc("Delete disk").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DiskTag}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}))

	container.Add(webservice)

	return nil
}
