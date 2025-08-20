/*
Copyright(c) 2025-present Accton. All rights reserved. www.accton.com
*/

package v1

import (
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/kapis/util"
	ui_pvc "kubesphere.io/kubesphere/pkg/models/pvc"
)

const (
	GroupName = "pvc.ecpaas.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

func AddToContainer(container *restful.Container, ksclient kubesphere.Interface, k8sclient kubernetes.Interface) error {
	webservice := runtime.NewWebService(GroupVersion)
	handler := newHandler(ksclient, k8sclient)

	// PVC clone
	webservice.Route(webservice.POST("/clone/namespaces/{namespace}/pvc/{name}").
		To(handler.ClonePVC).
		Param(webservice.PathParameter("namespace", "PVC namespace")).
		Param(webservice.PathParameter("name", "PVC name")).
		Reads(ui_pvc.PVCCloneRequest{}).
		Doc("Clone PVC").
		Returns(http.StatusOK, api.StatusOK, ui_pvc.PVCNameResponse{}).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.PVCTag}))

	container.Add(webservice)

	return nil
}
