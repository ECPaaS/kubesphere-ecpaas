/*
Copyright(c) 2025-present Accton. All rights reserved. www.accton.com.tw
*/

package v1alpha1

import (
	"net/http"
	"reflect"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	validation "kubesphere.io/kubesphere/pkg/kapis/util"
	"kubesphere.io/kubesphere/pkg/models/qos"
)

func newHandler(k8sclient kubernetes.Interface, ksclient kubesphere.Interface, dynamic dynamic.Interface) *handler {
	return &handler{
		qos: qos.New(k8sclient, ksclient, dynamic),
	}
}

type handler struct {
	qos qos.Interface
}

func (h *handler) ListDscp(request *restful.Request, response *restful.Response) {

	listDscp, err := h.qos.ListDSCP()

	if err != nil {
		if errors.IsNotFound(err) {
			klog.Error(err)
			response.WriteHeader(http.StatusNotFound)
			return
		} else {
			klog.Error(err)
		}
	}

	response.WriteEntity(listDscp)
}

func (h *handler) CreateDscp(request *restful.Request, response *restful.Response) {

	namespace := request.PathParameter("namespace")
	var dscpWithoutNs = qos.DscpWithoutNs{}

	err := request.ReadEntity(&dscpWithoutNs)

	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	var dscpType = qos.Dscp{}
	reflectType := reflect.TypeOf(dscpType)
	if !validation.IsValidWithinRange(reflectType, int(dscpWithoutNs.Dscp), "Dscp", response) {
		return
	}

	err = h.qos.CreateDSCP(namespace, dscpWithoutNs.Dscp)
	if err != nil {
		klog.Error(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusOK)
}

func (h *handler) GetDscp(request *restful.Request, response *restful.Response) {

	namespace := request.PathParameter("namespace")
	result, err := h.qos.GetDSCP(namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Error(err)
			response.WriteHeader(http.StatusNotFound)
			return
		}
		klog.Error(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.WriteEntity(result)
}

func (h *handler) UpdateDscp(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")

	var dscpWithoutNs = qos.DscpWithoutNs{}
	err := request.ReadEntity(&dscpWithoutNs)

	var dscpType = qos.Dscp{}
	reflectType := reflect.TypeOf(dscpType)
	if !validation.IsValidWithinRange(reflectType, int(dscpWithoutNs.Dscp), "Dscp", response) {
		return
	}

	if err != nil {
		klog.Error(err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.qos.UpdateDSCP(namespace, dscpWithoutNs.Dscp)
	if err != nil {
		klog.Error(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.WriteEntity(qos.Dscp{Dscp: dscpWithoutNs.Dscp, Namespace: namespace})
}

func (h *handler) DeleteDscp(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")

	err := h.qos.DeleteDSCP(namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Error(err)
			response.WriteHeader(http.StatusNotFound)
			return
		}
		klog.Error(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusNoContent)
}
