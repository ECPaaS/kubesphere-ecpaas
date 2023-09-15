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
		Reads(ui_virtz.VirtualMachine{}).
		Doc("Create virtual machine").
		Returns(http.StatusOK, api.StatusOK, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMahcineTag}))

	webservice.Route(webservice.PUT("/namespace/{namespace}/virtualmachine/{id}").
		To(handler.UpdateVirtualMahcine).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "virtual machine id")).
		Reads(ui_virtz.VirtualMachine{}).
		Doc("Update virtual machine").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.VirtualMachine{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMahcineTag}))

	webservice.Route(webservice.GET("/namespace/{namespace}/virtualmachine/{id}").
		To(handler.GetVirtualMachine).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "virtual machine id")).
		Doc("Get virtual machine").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.VirtualMachineResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMahcineTag}))

	webservice.Route(webservice.GET("/namespace/{namespace}/virtualmachine").
		To(handler.ListVirtualMachineWithNamespace).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Doc("List all virtual machine with namespace").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.ListVirtualMachineResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMahcineTag}))

	webservice.Route(webservice.GET("/virtualmachine").
		To(handler.ListVirtualMachine).
		Doc("List all virtual machine").
		Returns(http.StatusOK, api.StatusOK, ui_virtz.ListVirtualMachineResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMahcineTag}))

	webservice.Route(webservice.DELETE("/namespace/{namespace}/virtualmachine/{id}").
		To(handler.DeleteVirtualMachine).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("id", "virtual machine id")).
		Doc("Delete virtual machine").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMahcineTag}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}))

	container.Add(webservice)

	return nil
}
