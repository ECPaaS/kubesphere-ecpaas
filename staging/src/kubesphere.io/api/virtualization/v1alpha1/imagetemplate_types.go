/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com.tw
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

// ImageTemplateSource defines the source of the image.
type ImageTemplateSource struct {
	HTTP  *cdiv1.DataVolumeSourceHTTP `json:"http,omitempty"`
	Clone *cdiv1.DataVolumeSourcePVC  `json:"clone,omitempty"`
}

type ImageTemplateAttributes struct {
	Public bool `json:"public,omitempty"`
}

// ImageTemplateSpec defines the desired state of ImageTemplate
type ImageTemplateSpec struct {
	Attributes ImageTemplateAttributes `json:"attributes,omitempty"`
	// Resources represents the minimum resources the volume should have.
	Resources ResourceRequirements `json:"resources,omitempty"`
	// Source is the source of the volume.
	Source ImageTemplateSource `json:"source"`
}

// ImageTemplateStatus defines the observed state of ImageTemplate
type ImageTemplateStatus struct {
	Created bool   `json:"created,omitempty"`
	Owner   string `json:"owner,omitempty"`
	Ready   bool   `json:"ready,omitempty"`
	target  string `json:"target,omitempty"`
}

// +kubebuilder:subresource:status
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ImageTemplate is the Schema for the diskvolumes API
type ImageTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImageTemplateSpec   `json:"spec,omitempty"`
	Status ImageTemplateStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ImageTemplateList contains a list of VirtualMachine
type ImageTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImageTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ImageTemplate{}, &ImageTemplateList{})
}
