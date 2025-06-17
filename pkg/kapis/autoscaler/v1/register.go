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
	"kubesphere.io/kubesphere/pkg/models/autoscaler"
)

const (
	GroupName = "autoscaler.ecpaas.io"
)

var putNotes = `Any parameters which are not provided will not be changed.`

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

func AddToContainer(container *restful.Container, ksclient kubesphere.Interface, k8sclient kubernetes.Interface) error {
	webservice := runtime.NewWebService(GroupVersion)
	handler := newHandler(ksclient, k8sclient)

	// HPA
	webservice.Route(webservice.POST("/namespaces/{namespace}/hpa").
		To(handler.CreateHPA).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Reads(autoscaler.HpaRequest{}).
		Doc("Create HPA").
		Returns(http.StatusOK, api.StatusOK, autoscaler.HpaNameResponse{}).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AutoscalerHpaTag}))

	webservice.Route(webservice.PUT("/namespaces/{namespace}/hpa/{name}").
		To(handler.UpdateHPA).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("name", "hpa name")).
		Reads(autoscaler.ModifyHpaRequest{}).
		Doc("Update HPA").
		Notes(putNotes).
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AutoscalerHpaTag}))

	webservice.Route(webservice.GET("/namespaces/{namespace}/hpa/{name}").
		To(handler.GetHPA).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("name", "hpa name")).
		Doc("Get HPA").
		Returns(http.StatusOK, api.StatusOK, autoscaler.HpaResponse{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AutoscalerHpaTag}))

	webservice.Route(webservice.DELETE("/namespaces/{namespace}/hpa/{name}").
		To(handler.DeleteHPA).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("name", "hpa name")).
		Doc("Delete HPA").
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AutoscalerHpaTag}))

	// VPA
	webservice.Route(webservice.POST("/namespaces/{namespace}/vpa").
		To(handler.CreateVPA).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Reads(autoscaler.VpaRequest{}).
		Doc("Create VPA").
		Returns(http.StatusOK, api.StatusOK, autoscaler.VpaNameResponse{}).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AutoscalerVpaTag}))

	webservice.Route(webservice.PUT("/namespaces/{namespace}/vpa/{name}").
		To(handler.UpdateVPA).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("name", "vpa name")).
		Reads(autoscaler.ModifyVpaRequest{}).
		Doc("Update VPA").
		Notes(putNotes).
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AutoscalerVpaTag}))

	webservice.Route(webservice.GET("/namespaces/{namespace}/vpa/{name}").
		To(handler.GetVPA).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("name", "vpa name")).
		Doc("Get VPA").
		Returns(http.StatusOK, api.StatusOK, autoscaler.VpaResponse{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AutoscalerVpaTag}))

	webservice.Route(webservice.DELETE("/namespaces/{namespace}/vpa/{name}").
		To(handler.DeleteVPA).
		Param(webservice.PathParameter("namespace", "namespace name")).
		Param(webservice.PathParameter("name", "vpa name")).
		Doc("Delete VPA").
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AutoscalerVpaTag}))

	container.Add(webservice)

	return nil
}
