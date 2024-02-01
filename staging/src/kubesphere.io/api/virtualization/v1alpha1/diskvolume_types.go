/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com.tw
*/

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type DataVolumeBlankImage struct{}

// DataVolumeSourceImage provides the parameters to create a Data Volume from an existing Image
type DataVolumeSourceImage struct {
	// The namespace of the source Image
	Namespace string `json:"namespace"`
	// The name of the source Image
	Name string `json:"name"`
}

type DiskVolumeSource struct {
	Blank *DataVolumeBlankImage  `json:"blank,omitempty"`
	Image *DataVolumeSourceImage `json:"image,omitempty"`
}

// DiskVolumeSpec defines the desired state of DiskVolume
type DiskVolumeSpec struct {
	PVCName string `json:"pvcName,omitempty"`
	// Resources represents the minimum resources the volume should have.
	Resources ResourceRequirements `json:"resources,omitempty"`
	// Source is the source of the volume.
	Source DiskVolumeSource `json:"source"`
}

// DiskVolumeStatus defines the observed state of DiskVolume
type DiskVolumeStatus struct {
	Created bool   `json:"created,omitempty"`
	Owner   string `json:"owner,omitempty"`
	Ready   bool   `json:"ready,omitempty"`
	target  string `json:"target,omitempty"`
}

// +kubebuilder:subresource:status
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DiskVolume is the Schema for the diskvolumes API
type DiskVolume struct {
	metav1.TypeMeta `json:",inline"`
	// +kubebuilder:pruning:PreserveUnknownFields
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DiskVolumeSpec   `json:"spec,omitempty"`
	Status DiskVolumeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualMachineList contains a list of VirtualMachine
type DiskVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DiskVolume `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DiskVolume{}, &DiskVolumeList{})
}
