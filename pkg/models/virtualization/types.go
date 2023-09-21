/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com
*/

package virtualization

type ImageSpec struct {
	Name string `json:"name" description:"Image name"`
	Size string `json:"size" description:"Image size, range from 10Gi to 80Gi"`
}

type GuestSpec struct {
	Username string `json:"username" description:"Guest operating system username"`
	Password string `json:"password" description:"Guest operating system password"`
}

type AddDiskSpec struct {
	Size string `json:"size" description:"Disk size, range from 10Gi to 80Gi"`
}

type MountDiskSpec struct {
	Name      string `json:"name" description:"Disk name"`
	Namespace string `json:"namespace" description:"Disk namespace"`
}

type VirtualMachine struct {
	AliasName   string          `json:"alias_name" description:"Virtual machine name"`
	CpuCores    int32           `json:"cpu_cores" description:"Virtual machine cpu cores, range from 1 to 4"`
	Memory      string          `json:"memory" description:"Virtual machine memory size, range from 1Gi to 8Gi"`
	Description string          `json:"description,omitempty" description:"Virtual machine description"`
	AddDisk     []AddDiskSpec   `json:"add_disk,omitempty" description:"Add new disk for virtual machine"`
	MountDisk   []MountDiskSpec `json:"mount_disk,omitempty" description:"Mount disk for virtual machine"`
	Image       *ImageSpec      `json:"image,omitempty" description:"Virtual machine image source"`
	Guest       *GuestSpec      `json:"guest,omitempty" description:"Virtual machine guest operating system"`
}

type VMStatus struct {
	Ready bool   `json:"ready" description:"Virtual machine is ready or not"`
	State string `json:"state" description:"Virtual machine state"`
}

type VirtualMachineResponse struct {
	Name           string   `json:"name" description:"Virtual machine name"`
	AliasName      string   `json:"alias_name" description:"Virtual machine alias name"`
	Namespace      string   `json:"namespace" description:"Virtual machine namespace"`
	Description    string   `json:"description" description:"Virtual machine description"`
	SystemDiskSize string   `json:"system_disk_size" description:"Virtual machine system disk size"`
	CpuCores       int32    `json:"cpu_cores" description:"Virtual machine cpu cores"`
	Memory         string   `json:"memory" description:"Virtual machine memory size"`
	Disks          []string `json:"disks" description:"Virtual machine disks"`
	Status         VMStatus `json:"status" description:"Virtual machine status"`
}

type ListVirtualMachineResponse struct {
	TotalCount int                      `json:"total_count" description:"Total number of virtual machines"`
	Items      []VirtualMachineResponse `json:"items" description:"List of virtual machines"`
}
