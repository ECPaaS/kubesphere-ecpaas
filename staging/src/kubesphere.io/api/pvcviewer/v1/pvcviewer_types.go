/*
 Copyright(c) 2025-present Accton. All rights reserved. www.accton.com.tw
 */

package v1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// PVCViewerSpec defines the desired state of PVCViewer
type PVCViewerSpec struct {
	// Defines the PVC we want to edit
	// +required
	PVC string `json:"pvc"`

	// TODO infer default values if PodSpec is empty
	// Specifies the deployment's pod spec that will be created	by the operator.
	// +optional
	PodSpec corev1.PodSpec `json:"podSpec,omitempty"`

	// Specifies custom networking specification for the underlying
	// Service and VirtualService resources
	// +optional
	Networking Networking `json:"networking,omitempty"`

	// If set to true, the controller detects RWO-Volumes referred to by
	// the Pod and uses affinities to schedule the PVCViewer to nodes
	// where the volume is currently mounted. This enables the PVCViewer
	// to access RWO-Volumes, even though they might already be mounted.
	// +kubebuilder:default:=false
	RWOScheduling bool `json:"rwoScheduling"`
}

type Networking struct {
	// Specifies the application's target port used by the Deployment's Service.
	// +optional
	TargetPort intstr.IntOrString `json:"targetPort"`
}

// PVCViewerStatus defines the observed state of PVCViewer
type PVCViewerStatus struct {
	// Conditions will be a mirror of the underlying Pod’s Conditions
	Conditions []appsv1.DeploymentCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// Ready defines if the viewer is ready to be used,
	// i.e. Replicas==ReadyReplicas
	// +kubebuilder:default:=false
	Ready bool `json:"ready"`

	// Display the service IP and node port of the PVC viewer
	ServiceIP *string `json:"serviceIp,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PVCViewer is the Schema for the pvcviewers API
type PVCViewer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PVCViewerSpec   `json:"spec,omitempty"`
	Status PVCViewerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PVCViewerList contains a list of PVCViewer
type PVCViewerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PVCViewer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PVCViewer{}, &PVCViewerList{})
}