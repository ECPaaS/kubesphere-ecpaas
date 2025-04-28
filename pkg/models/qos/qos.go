/*
Copyright(c) 2025-present Accton. All rights reserved. www.accton.com.tw
*/

package qos

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
)

type Interface interface {
	ListDSCP() (*DscpList, error)
	GetDSCP(namespace string) (*Dscp, error)
	CreateDSCP(namespace string, dscp int64) error
	UpdateDSCP(namespace string, dscp int64) error
	DeleteDSCP(namespace string) error
}

type qosOperator struct {
	ksclient  kubesphere.Interface
	k8sclient kubernetes.Interface
	dynamic   dynamic.Interface
}

func New(k8sclient kubernetes.Interface, ksclient kubesphere.Interface, dynamic dynamic.Interface) Interface {
	return &qosOperator{
		k8sclient: k8sclient,
		ksclient:  ksclient,
		dynamic:   dynamic,
	}
}

const (
	Group   = "k8s.ovn.org"
	Version = "v1"
	Kind    = "egressqoses"
)

func (h *qosOperator) ListDSCP() (*DscpList, error) {

	gvr := schema.GroupVersionResource{
		Group:    Group,
		Version:  Version,
		Resource: Kind, // "egressqoses"
	}

	list, err := h.dynamic.Resource(gvr).Namespace("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list EgressQoS: %v", err)
	}

	dscpList := DscpList{
		Items: []Dscp{},
	}

	for _, item := range list.Items {
		namespace, found, err := unstructured.NestedString(item.Object, "metadata", "namespace")
		if err != nil || !found {
			continue
		}

		egressList, found, err := unstructured.NestedSlice(item.Object, "spec", "egress")
		if err != nil || !found {
			continue
		}

		for _, e := range egressList {
			egressEntry, ok := e.(map[string]interface{})
			if !ok {
				continue
			}

			dscpVal, found, err := unstructured.NestedInt64(egressEntry, "dscp")
			if err != nil || !found {
				continue
			}

			dscpList.Items = append(dscpList.Items, Dscp{
				Namespace: namespace,
				Dscp:      dscpVal,
			})
		}
	}
	dscpList.TotalCount = len(list.Items)

	return &dscpList, nil
}

func (h *qosOperator) GetDSCP(namespace string) (*Dscp, error) {
	gvr := schema.GroupVersionResource{
		Group:    Group,
		Version:  Version,
		Resource: Kind,
	}

	resource, err := h.dynamic.Resource(gvr).Namespace(namespace).Get(context.TODO(), "default", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get EgressQoS with namespace %v: %w", namespace, err)
	}

	egressList, found, err := unstructured.NestedSlice(resource.Object, "spec", "egress")
	if err != nil || !found || len(egressList) == 0 {
		return nil, fmt.Errorf("egress not found: %w", err)
	}

	egressEntry, ok := egressList[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid egress format: %w", err)
	}

	dscpVal, found, err := unstructured.NestedInt64(egressEntry, "dscp")
	if err != nil || !found {
		return nil, fmt.Errorf("dscp not found: %w", err)
	}

	return &Dscp{
		Namespace: namespace,
		Dscp:      dscpVal,
	}, nil
}

func (h *qosOperator) CreateDSCP(namespace string, dscp int64) error {
	gvr := schema.GroupVersionResource{
		Group:    Group,
		Version:  Version,
		Resource: Kind,
	}

	egressQoS := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", Group, Version),
			"kind":       "EgressQoS",
			"metadata": map[string]interface{}{
				"name":      "default", // 固定 name
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"egress": []interface{}{
					map[string]interface{}{
						"dscp": dscp,
					},
				},
			},
		},
	}

	_, err := h.dynamic.Resource(gvr).Namespace(namespace).Create(context.TODO(), egressQoS, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create EgressQoS: %w", err)
	}

	return nil
}

func (h *qosOperator) UpdateDSCP(namespace string, dscp int64) error {
	gvr := schema.GroupVersionResource{
		Group:    Group,
		Version:  Version,
		Resource: Kind,
	}

	// Get the existing resource
	resource, err := h.dynamic.Resource(gvr).Namespace(namespace).Get(context.TODO(), "default", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get EgressQoS with namespace %v: for update: %w", namespace, err)
	}

	// Update spec.egress[0].dscp
	egressList := []interface{}{map[string]interface{}{
		"dscp": dscp,
	}}

	if err := unstructured.SetNestedSlice(resource.Object, egressList, "spec", "egress"); err != nil {
		return fmt.Errorf("failed to set dscp: %w", err)
	}

	// Apply the update
	_, err = h.dynamic.Resource(gvr).Namespace(namespace).Update(context.TODO(), resource, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update EgressQoS: %w", err)
	}

	return nil
}

func (h *qosOperator) DeleteDSCP(namespace string) error {
	gvr := schema.GroupVersionResource{
		Group:    Group,
		Version:  Version,
		Resource: Kind,
	}

	err := h.dynamic.Resource(gvr).Namespace(namespace).Delete(context.TODO(), "default", metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete EgressQoS: %w", err)
	}

	return nil
}
