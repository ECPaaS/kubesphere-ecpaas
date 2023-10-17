/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com
*/

package virtualization

import (
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	ui_virtz "kubesphere.io/kubesphere/pkg/models/virtualization"
)

const (
	GroupName = "virtualization.ecpaas.io"
)

var vmPutNotes = `Any parameters which are not provied will not be changed.
When the cpu cores or memory parameter changed, the virtual machine need be restarted.`

var diskPutNotes = `Any parameters which are not provied will not be changed.`

var imagePutNotes = `Any parameters which are not provied will not be changed.`

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

func AddToContainer(container *restful.Container, ksclient kubesphere.Interface, k8sclient kubernetes.Interface) error {
	webservice := runtime.NewWebService(GroupVersion)
	handler := newHandler(ksclient, k8sclient)

	vmPutNotes = strings.ReplaceAll(vmPutNotes, "\n", " ")
	diskPutNotes = strings.ReplaceAll(diskPutNotes, "\n", " ")

	webservice.Route(webservice.POST("/namespace/{namespace}/virtualmachine").
		To(handler.CreateVirtualMahcine).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Reads(ui_virtz.VirtualMachineRequest{}).
		Doc("Create virtual machine").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.VirtualMachineIDResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMachineTag}))

	webservice.Route(webservice.PUT("/namespace/{namespace}/virtualmachine/{id}").
		To(handler.UpdateVirtualMahcine).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "virtual machine id")).
		Reads(ui_virtz.ModifyVirtualMachineRequest{}).
		Doc("Update virtual machine").
		Notes(vmPutNotes).
		Returns(http.StatusOK, api.StatusOK, nil).
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
		Returns(http.StatusOK, api.StatusOK, nil))

	webservice.Route(webservice.POST("/namespace/{namespace}/disk").
		To(handler.CreateDisk).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Reads(ui_virtz.DiskRequest{}).
		Doc("Create disk").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.DiskIDResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DiskTag}))

	webservice.Route(webservice.PUT("/namespace/{namespace}/disk/{id}").
		To(handler.UpdateDisk).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "disk id")).
		Reads(ui_virtz.ModifyDiskRequest{}).
		Doc("Update disk").
		Notes(diskPutNotes).
		Returns(http.StatusOK, api.StatusOK, nil).
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
		Returns(http.StatusOK, api.StatusOK, nil))

	webservice.Route(webservice.POST("/namespace/{namespace}/image").
		To(handler.CreateImage).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Reads(ui_virtz.ImageRequest{}).
		Doc("Create image").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.ImageIDResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ImageTag}))

	webservice.Route(webservice.PUT("/namespace/{namespace}/image/{id}").
		To(handler.UpdateImage).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "image id")).
		Reads(ui_virtz.ModifyImageRequest{}).
		Doc("Update image").
		Notes(imagePutNotes).
		Returns(http.StatusOK, api.StatusOK, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ImageTag}))

	webservice.Route(webservice.GET("/namespace/{namespace}/image/{id}").
		To(handler.GetImage).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "image id")).
		Doc("Get image").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.ImageResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ImageTag}))

	webservice.Route(webservice.GET("/namespace/{namespace}/image").
		To(handler.ListImageWithNamespace).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Doc("List all image with namespace").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.ListImageResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ImageTag}))

	webservice.Route(webservice.GET("/image").
		To(handler.ListImage).
		Doc("List all image").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.ListImageResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ImageTag}))

	webservice.Route(webservice.DELETE("/namespace/{namespace}/image/{id}").
		To(handler.DeleteImage).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "image id")).
		Doc("Delete image").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ImageTag}).
		Returns(http.StatusOK, api.StatusOK, nil))

	container.Add(webservice)

	return nil
}
