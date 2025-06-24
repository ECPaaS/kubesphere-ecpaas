/*
Copyright(c) 2025-present Accton. All rights reserved. www.accton.com
*/

package v1

import (
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	ui_autoscaler "kubesphere.io/kubesphere/pkg/models/autoscaler"
)

type handler struct {
	autoscaler ui_autoscaler.Interface
}

func newHandler(ksclient kubesphere.Interface, k8sclient kubernetes.Interface) handler {
	return handler{
		autoscaler: ui_autoscaler.New(ksclient, k8sclient),
	}
}

// HPA

func (h *handler) CreateHPA(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")

	var ui_hpa ui_autoscaler.HpaRequest
	err := req.ReadEntity(&ui_hpa)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Validate request
	if !isValidHpaRequest(&ui_hpa , resp) {
		return
	}

	// Create HPA
	ui_hpaName, err := h.autoscaler.CreateHPA(namespace, ui_hpa.ResourceName, &ui_hpa)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(ui_hpaName)
}

func (h *handler) UpdateHPA(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	name := req.PathParameter("name")
	var ui_hpa ui_autoscaler.ModifyHpaRequest
	err := req.ReadEntity(&ui_hpa)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Get HPA first
	/*_, err = h.autoscaler.GetHPA(namespace, name)
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}*/

	// Validate request
	if !isValidHpaModifyRequest(&ui_hpa, resp) {
		return
	}

	// Update HPA
	hpaResponse, err := h.autoscaler.UpdateHPA(namespace, name, &ui_hpa)
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(hpaResponse)
}

func (h *handler) GetHPA(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	name := req.PathParameter("name")

	hpaResponse, err := h.autoscaler.GetHPA(namespace, name)
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(hpaResponse)
}

func (h *handler) DeleteHPA(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	name := req.PathParameter("name")

	err := h.autoscaler.DeleteHPA(namespace, name)
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(http.StatusOK)
}

// VPA

func (h *handler) CreateVPA(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	var ui_vpa ui_autoscaler.VpaRequest
	err := req.ReadEntity(&ui_vpa)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Validate request
	if !isValidVpaRequest(&ui_vpa , resp) {
		return
	}

	// Create HPA
	ui_vpaName, err := h.autoscaler.CreateVPA(namespace, ui_vpa.ResourceName, &ui_vpa)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(ui_vpaName)
}

func (h *handler) UpdateVPA(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	name := req.PathParameter("name")
	var ui_vpa ui_autoscaler.ModifyVpaRequest
	err := req.ReadEntity(&ui_vpa)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Get VPA first
	/*_, err = h.autoscaler.GetVPA(namespace, name)
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}*/

	// Validate request
	if !isValidVpaModifyRequest(&ui_vpa, resp) {
		return
	}

	// Update VPA
	vpaResponse, err := h.autoscaler.UpdateVPA(namespace, name, &ui_vpa)
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(vpaResponse)
}

func (h *handler) GetVPA(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	name := req.PathParameter("name")

	vpaResponse, err := h.autoscaler.GetVPA(namespace, name)
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(vpaResponse)
}

func (h *handler) DeleteVPA(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	name := req.PathParameter("name")

	err := h.autoscaler.DeleteVPA(namespace, name)
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(http.StatusOK)
}
