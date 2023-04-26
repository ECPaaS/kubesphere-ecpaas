/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com.tw
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindVpcSubnet     = "VPCSubnet"
	ResourceSingularVpcSubnet = "vpcsubnet"
	ResourcePluralVpcSubnets  = "vpcsubnets"
	VpcSubnetLabel            = "k8s.ovn.org/vpcsubnet"
)

// VPCSubnetSpec defines the desired state of VPCSubnet
type VPCSubnetSpec struct {
	// vpc subnet private segment address space
	CIDR string `json:"cidr"`
	// +kubebuilder:validation:Required
	// vpc network name
	Vpc string `json:"vpc"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +k8s:openapi-gen=true
// VPCSubnet is the Schema for the vpcsubnets API
type VPCSubnet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VPCSubnetSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// VPCSubnetList contains a list of VPCSubnet
type VPCSubnetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VPCSubnet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VPCSubnet{}, &VPCSubnetList{})
}
