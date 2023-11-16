/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com
*/

package virtualization

import (
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/minio/minio-go/v7"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	ui_virtz "kubesphere.io/kubesphere/pkg/models/virtualization"
)

const (
	GroupName = "virtualization.ecpaas.io"
)

var vmPutNotes = `Any parameters which are not provied will not be changed.
When the cpu cores or memory parameter changed, the virtual machine need be restarted.`

var diskPutNotes = `Any parameters which are not provied will not be changed.`

var imagePutNotes = `Any parameters which are not provied will not be changed.`
var imagePostCloneNotes = `Source image's namespace shall be different from new image's namespace.`

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

func AddToContainer(container *restful.Container, minioClient *minio.Client, ksclient kubesphere.Interface, k8sclient kubernetes.Interface, factory informers.InformerFactory) error {
	webservice := runtime.NewWebService(GroupVersion)
	handler := newHandler(ksclient, k8sclient, factory, minioClient)

	vmPutNotes = strings.ReplaceAll(vmPutNotes, "\n", " ")
	diskPutNotes = strings.ReplaceAll(diskPutNotes, "\n", " ")

	webservice.Route(webservice.POST("/namespaces/{namespace}/virtualmachines").
		To(handler.CreateVirtualMahcine).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Reads(ui_virtz.VirtualMachineRequest{}).
		Doc("Create virtual machine").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.VirtualMachineIDResponse{}).
		Returns(http.StatusForbidden, "Invalid format", BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMachineTag}))

	webservice.Route(webservice.PUT("/namespaces/{namespace}/virtualmachines/{id}").
		To(handler.UpdateVirtualMahcine).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "virtual machine id")).
		Reads(ui_virtz.ModifyVirtualMachineRequest{}).
		Doc("Update virtual machine").
		Notes(vmPutNotes).
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusForbidden, "Invalid format", BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMachineTag}))

	webservice.Route(webservice.GET("/namespaces/{namespace}/virtualmachines/{id}").
		To(handler.GetVirtualMachine).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "virtual machine id")).
		Doc("Get virtual machine").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.VirtualMachineResponse{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMachineTag}))

	webservice.Route(webservice.GET("/namespaces/{namespace}/virtualmachines").
		To(handler.ListVirtualMachineWithNamespace).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Doc("List all virtual machine with namespace").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.ListVirtualMachineResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMachineTag}))

	webservice.Route(webservice.GET("/virtualmachines").
		To(handler.ListVirtualMachine).
		Doc("List all virtual machine").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.ListVirtualMachineResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMachineTag}))

	webservice.Route(webservice.DELETE("/namespaces/{namespace}/virtualmachines/{id}").
		To(handler.DeleteVirtualMachine).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "virtual machine id")).
		Doc("Delete virtual machine").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMachineTag}).
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil))

	webservice.Route(webservice.POST("/namespaces/{namespace}/disks").
		To(handler.CreateDisk).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Reads(ui_virtz.DiskRequest{}).
		Doc("Create disk").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.DiskIDResponse{}).
		Returns(http.StatusForbidden, "Invalid format", BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DiskTag}))

	webservice.Route(webservice.PUT("/namespaces/{namespace}/disks/{id}").
		To(handler.UpdateDisk).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "disk id")).
		Reads(ui_virtz.ModifyDiskRequest{}).
		Doc("Update disk").
		Notes(diskPutNotes).
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusForbidden, "Invalid format", BadRequestError{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DiskTag}))

	webservice.Route(webservice.GET("/namespaces/{namespace}/disks/{id}").
		To(handler.GetDisk).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "disk id")).
		Doc("Get disk").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.DiskResponse{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DiskTag}))

	webservice.Route(webservice.GET("/namespaces/{namespace}/disks").
		To(handler.ListDiskWithNamespace).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Doc("List all disk with namespace").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.ListDiskResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DiskTag}))

	webservice.Route(webservice.GET("/disks").
		To(handler.ListDisk).
		Doc("List all disk").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.ListDiskResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DiskTag}))

	webservice.Route(webservice.DELETE("/namespaces/{namespace}/disks/{id}").
		To(handler.DeleteDisk).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "disk id")).
		Doc("Delete disk").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DiskTag}).
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil))

	webservice.Route(webservice.POST("/namespaces/{namespace}/images").
		To(handler.CreateImage).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Reads(ui_virtz.ImageRequest{}).
		Doc("Create image").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.ImageIDResponse{}).
		Returns(http.StatusForbidden, "Invalid format", BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ImageTag}))

	webservice.Route(webservice.POST("/namespaces/{namespace}/images/clone").
		To(handler.CloneImage).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Reads(ui_virtz.CloneImageRequest{}).
		Doc("Clone image").
		Notes(imagePostCloneNotes).
		Returns(http.StatusOK, api.StatusOK, ui_virtz.ImageIDResponse{}).
		Returns(http.StatusForbidden, "Invalid format", BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ImageTag}))

	webservice.Route(webservice.PUT("/namespaces/{namespace}/images/{id}").
		To(handler.UpdateImage).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "image id")).
		Reads(ui_virtz.ModifyImageRequest{}).
		Doc("Update image").
		Notes(imagePutNotes).
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusForbidden, "Invalid format", BadRequestError{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ImageTag}))

	webservice.Route(webservice.GET("/namespaces/{namespace}/images/{id}").
		To(handler.GetImage).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "image id")).
		Doc("Get image").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.ImageResponse{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ImageTag}))

	webservice.Route(webservice.GET("/namespaces/{namespace}/images").
		To(handler.ListImageWithNamespace).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Doc("List all image with namespace").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.ListImageResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ImageTag}))

	webservice.Route(webservice.GET("/images").
		To(handler.ListImage).
		Doc("List all image").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.ListImageResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ImageTag}))

	webservice.Route(webservice.DELETE("/namespaces/{namespace}/images/{id}").
		To(handler.DeleteImage).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "image id")).
		Doc("Delete image").
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ImageTag}))

	webservice.Route(webservice.GET("virtualization/namespaces/{namespace}/quotas").
		To(handler.handleVirtualizationGetNamespaceQuotas).
		Param(webservice.PathParameter("namespace", "the name of the project")).
		Doc("Get specified namespace's of virtualization resource quota and usage").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.VirtualizationResourceQuota{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ResourceQuotasTag}))

	container.Add(webservice)

	return nil
}
