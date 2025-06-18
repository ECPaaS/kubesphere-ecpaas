/*
Copyright(c) 2025-present Accton. All rights reserved. www.accton.com.tw
*/

package v1alpha1

import (
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/kapis/util"
	"kubesphere.io/kubesphere/pkg/models/qos"
)

const (
	GroupName = "qos.ecpaas.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

func Resource(resource string) schema.GroupResource {
	return GroupVersion.WithResource(resource).GroupResource()
}

func AddToContainer(container *restful.Container, ksclient kubesphere.Interface, k8sclient kubernetes.Interface, dynamic dynamic.Interface) error {
	webservice := runtime.NewWebService(GroupVersion)
	handler := newHandler(k8sclient, ksclient, dynamic)

	// GET (List All)
	webservice.Route(webservice.GET("/ovnk/dscp").
		To(handler.ListDscp).
		Doc("List all of the namespaces DSCP value ").
		Returns(http.StatusOK, api.StatusOK, qos.DscpList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.QoSTag}))

	// POST (Create)
	webservice.Route(webservice.POST("/ovnk/namespaces/{namespace}/dscp").
		To(handler.CreateDscp).
		Doc("Create a DSCP value with namespaces").
		Param(webservice.PathParameter("namespace", "Namespace to create")).
		Reads(qos.DscpWithoutNs{}).
		Returns(http.StatusOK, api.StatusOK, qos.Dscp{}).
		Returns(http.StatusBadRequest, api.StatusBadRequest, nil).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.QoSTag}))

	// PUT (Update)
	webservice.Route(webservice.PUT("/ovnk/namespaces/{namespace}/dscp").
		To(handler.UpdateDscp).
		Doc("Update a DSCP value with namespaces").
		Param(webservice.PathParameter("namespace", "Namespace to update")).
		Reads(qos.DscpWithoutNs{}).
		Returns(http.StatusOK, api.StatusOK, qos.Dscp{}).
		Returns(http.StatusBadRequest, api.StatusBadRequest, util.BadRequestError{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.QoSTag}))

	// GET (Single)
	webservice.Route(webservice.GET("/ovnk/namespaces/{namespace}/dscp").
		To(handler.GetDscp).
		Doc("Get a DSCP value with namespaces").
		Param(webservice.PathParameter("namespace", "Namespace to fetch")).
		Returns(http.StatusOK, api.StatusOK, qos.Dscp{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.QoSTag}))

	// DELETE
	webservice.Route(webservice.DELETE("/ovnk/namespaces/{namespace}/dscp").
		To(handler.DeleteDscp).
		Doc("Delete a DSCP value with namespaces").
		Param(webservice.PathParameter("namespace", "Namespace to delete")).
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.QoSTag}))

	container.Add(webservice)

	return nil
}
