/*
Copyright(c) 2025-present Accton. All rights reserved. www.accton.com
*/

package pvc

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	pvcv1 "kubesphere.io/api/pvc/v1"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
)

type Interface interface {
	ClonePVC(namespace string, name string, ui_pvc *PVCCloneRequest) (*PVCNameResponse, error)
}

type pvcCloneOperator struct {
	ksClient  kubesphere.Interface
	k8sClient kubernetes.Interface
}

func New(ksclient kubesphere.Interface, k8sclient kubernetes.Interface) Interface {
	return &pvcCloneOperator{
		ksClient:   ksclient,
		k8sClient:  k8sclient,
	}
}

func (o *pvcCloneOperator) ClonePVC(namespace string, name string, ui_pvc *PVCCloneRequest) (*PVCNameResponse, error) {
	klog.V(2).Infof("Cloning PVC: \"%s\" in \"%s\" namespace", name, namespace)
	// Check source PVC existence
	_, err := o.k8sClient.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting source PVC \"%s\" in \"%s\": %s", name, namespace, err.Error())
	}
	// Check target PVC existence
	_, err = o.k8sClient.CoreV1().PersistentVolumeClaims(ui_pvc.Namespace).Get(context.Background(), ui_pvc.NewName, metav1.GetOptions{})
	if err == nil {
		return nil, fmt.Errorf("error target PVC \"%s\" in \"%s\" already exists", ui_pvc.NewName, ui_pvc.Namespace)
	}
	
	// Create PVCCloneRequest
	request := &pvcv1.PVCCloneRequest{}
	request.Name = ui_pvc.NewName + "-request"
	request.Namespace = ui_pvc.Namespace // target PVC namespace
	request.Spec.SourcePVCName = name
	request.Spec.SourcePVCNamespace = namespace
	request.Spec.TargetPVCName = ui_pvc.NewName
	request.Spec.TargetPVCNamespace = ui_pvc.Namespace
	request.Spec.StorageClass = ui_pvc.StorageClass
	request.Spec.Size = ui_pvc.Size
	request.Spec.AccessModes = ui_pvc.AccessModes
	request.Spec.NodeSelector = ConvertNodeSelectorToMap(ui_pvc.NodeSelector)
	_, err = o.ksClient.PvcV1().PVCCloneRequests(ui_pvc.Namespace).Create(context.Background(), request, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return &PVCNameResponse{NewName: ui_pvc.NewName}, nil
}
