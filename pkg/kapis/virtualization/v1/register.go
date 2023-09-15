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
	virtz "kubesphere.io/kubesphere/pkg/models/virtualization"
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
		Reads(virtz.VirtualMachine{}).
		Doc("Create virtual machine").
		Returns(http.StatusOK, api.StatusOK, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMahcineTag}))

	webservice.Route(webservice.DELETE("/namespace/{namespace}/virtualmachine/{virtualmachine}").
		To(handler.DeleteVirtualMachine).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("virtualmachine", "virtual machine name")).
		Doc("Delete virtual machine").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.VirtualMahcineTag}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}))

	container.Add(webservice)

	return nil
}
