/*
Copyright(c) 2025-present Accton. All rights reserved. www.accton.com.tw
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Spec defines the desired state of PVCCloneRequest.
type Spec struct {
	SourcePVCName      string            `json:"sourcePVCName,omitempty"`
	SourcePVCNamespace string            `json:"sourcePVCNamespace,omitempty"`
	TargetPVCName      string            `json:"targetPVCName,omitempty"`
	TargetPVCNamespace string            `json:"targetPVCNamespace,omitempty"`
	StorageClass       string            `json:"storageClass,omitempty"`
	Size               int               `json:"size,omitempty"`
	AccessModes        []string          `json:"accessModes,omitempty"`
	NodeSelector       map[string]string `json:"nodeSelector,omitempty"`
}

// Status defines the observed state of PVCCloneRequest.
type Status struct {
	Phase string `json:"phase,omitempty"`
}

// +kubebuilder:subresource:status
// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PVCCloneRequest is the Schema for the PVCCloneRequest API.
type PVCCloneRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Spec   `json:"spec,omitempty"`
	Status Status `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PVCCloneRequestList contains a list of PVCCloneRequest.
type PVCCloneRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PVCCloneRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PVCCloneRequest{}, &PVCCloneRequestList{})
}
