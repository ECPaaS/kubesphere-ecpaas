/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com.tw
*/

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kvapi "kubevirt.io/api/core/v1"
)

const (
	VirtualMachineFinalizer = "finalizers.virtualization.ecpaas.io/virtualmachine"

	VirtualizationBootOrder          = "virtualization.ecpaas.io/bootorder"
	VirtualizationDiskType           = "virtualization.ecpaas.io/disk-type"
	VirtualizationImageInfo          = "virtualization.ecpaas.io/image-info"
	VirtualizationAliasName          = "virtualization.ecpaas.io/alias-name"
	VirtualizationCpuCores           = "virtualization.ecpaas.io/cpu-cores"
	VirtualizationImageMemory        = "virtualization.ecpaas.io/image-memory"
	VirtualizationImageStorage       = "virtualization.ecpaas.io/image-storage"
	VirtualizationOSFamily           = "virtualization.ecpaas.io/os-family"
	VirtualizationOSVersion          = "virtualization.ecpaas.io/os-version"
	VirtualizationOSPlatform         = "virtualization.ecpaas.io/os-platform"
	VirtualizationDescription        = "virtualization.ecpaas.io/description"
	VirtualizationSystemDiskSize     = "virtualization.ecpaas.io/system-disk-size"
	VirtualizationSystemDiskName     = "virtualization.ecpaas.io/system-disk-name"
	VirtualizationDiskVolumeOwner    = "virtualization.ecpaas.io/disk-volume-owner"
	VirtualizationLastDiskVolumes    = "virtualization.ecpaas.io/last-disk-volumes"
	VirtualizationImageType          = "virtualization.ecpaas.io/image-type"
	VirtualizationDiskMode           = "virtualization.ecpaas.io/disk-mode"
	VirtualizationDiskMinioImageName = "virtualization.ecpaas.io/disk-minio-image-name"
	VirtualizationDiskMediaType      = "virtualization.ecpaas.io/disk-media-type"
	VirtualizationDiskHotpluggable   = "virtualization.ecpaas.io/disk-hotpluggable"
)

const (
	VirtualMachineRunStrategyAlways = "always"
	VirtualMachineRunStrategyHalted = "halted"
)

type ResourceRequirements struct {
	// Requests is a description of the initial vmi resources.
	// Valid resource keys are "memory" and "cpu".
	// +optional
	Requests v1.ResourceList `json:"requests,omitempty"`
	// Limits describes the maximum amount of compute resources allowed.
	// Valid resource keys are "memory" and "cpu".
	// +optional
	Limits v1.ResourceList `json:"limits,omitempty"`
}

type CPU struct {
	Cores uint32 `json:"cores,omitempty"`
}

type DomainSpec struct {
	CPU CPU `json:"cpu,omitempty"`
	// Machine type.
	// +optional
	Machine *kvapi.Machine `json:"machine,omitempty"`
	// Features like acpi, apic, hyperv, smm.
	// +optional
	Features *kvapi.Features `json:"features,omitempty"`
	// Devices allows adding disks, network interfaces, and others
	Devices   kvapi.Devices        `json:"devices,omitempty"`
	Resources ResourceRequirements `json:"resources,omitempty"`
}

type Multus struct {
	NetworkName string `json:"networkName,omitempty"`
}

type Network struct {
	Name          string `json:"name"`
	NetworkSource `json:",inline"`
}

type NetworkSource struct {
	Pod    *PodNetwork    `json:"pod,omitempty"`
	Multus *MultusNetwork `json:"multus,omitempty"`
}

type PodNetwork struct {
	VMNetworkCIDR     string `json:"vmNetworkCIDR,omitempty"`
	VMIPv6NetworkCIDR string `json:"vmIPv6NetworkCIDR,omitempty"`
}

type MultusNetwork struct {
	NetworkName string `json:"networkName"`
	Default     bool   `json:"default,omitempty"`
}

type Hardware struct {
	Domain           DomainSpec     `json:"domain,omitempty"`
	EvictionStrategy string         `json:"evictionStrategy,omitempty"`
	Hostname         string         `json:"hostname,omitempty"`
	Networks         []Network      `json:"networks,omitempty"`
	Volumes          []kvapi.Volume `json:"volumes,omitempty"`
}

// VirtualMachineSpec defines the desired state of VirtualMachine
type VirtualMachineSpec struct {
	// DiskVolumeTemplate is the name of the DiskVolumeTemplate.
	DiskVolumeTemplates []DiskVolume `json:"diskVolumeTemplates,omitempty"`
	// DiskVolume is the name of the DiskVolume.
	DiskVolumes []string `json:"diskVolumes,omitempty"`
	// Hardware is the hardware of the VirtualMachine.
	Hardware Hardware `json:"hardware,omitempty"`
	// RunStrategy is the run strategy of the VirtualMachine.
	RunStrategy string `json:"runStrategy,omitempty"`
}

// +kubebuilder:resource:shortName={ksvm,ksvms}
// +kubebuilder:subresource:status
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualMachine runs a vm at a given name.
type VirtualMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtualMachineSpec         `json:"spec,omitempty"`
	Status kvapi.VirtualMachineStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualMachineList contains a list of VirtualMachine
type VirtualMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtualMachine{}, &VirtualMachineList{})
}
