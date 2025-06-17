/*
Copyright(c) 2025-present Accton. All rights reserved. www.accton.com
*/

package v1

import (
	"github.com/emicklei/go-restful"
	"k8s.io/client-go/kubernetes"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	//ui_autoscaler "kubesphere.io/kubesphere/pkg/models/autoscaler"
)

type handler struct {
	//autoscaler               ui_autoscaler.Interface
	//minioClient         *minio.Client
	ksClient  kubesphere.Interface
	k8sClient kubernetes.Interface
}

func newHandler(ksclient kubesphere.Interface, k8sclient kubernetes.Interface) handler {
	return handler{
		//autoscaler:               ui_autoscaler.New(ksclient, k8sclient),
		//minioClient:         minioClient,
		ksClient:  ksclient,
		k8sClient: k8sclient,
	}
}

// HPA

func (h *handler) CreateHPA(req *restful.Request, resp *restful.Response) { }

func (h *handler) UpdateHPA(req *restful.Request, resp *restful.Response) { }

func (h *handler) GetHPA(req *restful.Request, resp *restful.Response) { }

func (h *handler) DeleteHPA(req *restful.Request, resp *restful.Response) { }

// VPA

func (h *handler) CreateVPA(req *restful.Request, resp *restful.Response) { }

func (h *handler) UpdateVPA(req *restful.Request, resp *restful.Response) { }

func (h *handler) GetVPA(req *restful.Request, resp *restful.Response) { }

func (h *handler) DeleteVPA(req *restful.Request, resp *restful.Response) { }
