/*
Copyright(c) 2025-present Accton. All rights reserved. www.accton.com
*/

package v1

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"k8s.io/client-go/kubernetes"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"

	ui_pvc "kubesphere.io/kubesphere/pkg/models/pvc"
)


type pvcHandler struct {
	pvc       ui_pvc.Interface
	k8sClient kubernetes.Interface
}

func newHandler(ksclient kubesphere.Interface, k8sclient kubernetes.Interface) pvcHandler {
	return pvcHandler{
		pvc:       ui_pvc.New(ksclient, k8sclient),
		k8sClient: k8sclient,
	}
}

// Create new clone request
func (h *pvcHandler) ClonePVC(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	name := req.PathParameter("name")

	var ui_pvc ui_pvc.PVCCloneRequest
	err := req.ReadEntity(&ui_pvc)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	if !isValidClonePVCRequest(&ui_pvc , resp) {
		return
	}

	ui_newPVCName, err := h.pvc.ClonePVC(namespace, name, &ui_pvc)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(ui_newPVCName)
}
