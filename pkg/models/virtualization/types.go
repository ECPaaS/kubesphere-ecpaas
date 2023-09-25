/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com
*/

package virtualization

const (
	diskVolumeNamePrefix = "disk-" // disk: disk volume
)

type ImageInfo struct {
	Name      string `json:"name" description:"Image name"`
	Namespace string `json:"namespace" description:"Image namespace"`
	System    string `json:"system" description:"Image system"`
	Version   string `json:"version" description:"Image version"`
	AliasName string `json:"aliasName" description:"Image alias name"`
	ImageSize string `json:"imageSize" description:"Image size"`
	Cpu       string `json:"cpu" description:"cpu used by image"`
	Memory    string `json:"memory" description:"memory used by image"`
}

type ImageSpec struct {
	Name string `json:"name,omitempty" description:"Image name"`
	Size string `json:"size" default:"20Gi" description:"Image size, range from 10Gi to 80Gi"`
}

type GuestSpec struct {
	Username string `json:"username,omitempty" description:"Guest operating system username"`
	Password string `json:"password,omitempty" description:"Guest operating system password"`
}

type DiskSpec struct {
	Name      string `json:"name,omitempty" description:"Disk name"`
	ID        string `json:"id,omitempty" description:"Disk id"`
	Namespace string `json:"namespace,omitempty" description:"Disk namespace"`
	Type      string `json:"type,omitempty" description:"Disk type, system or data"`
	Size      string `json:"size,omitempty" default:"20Gi" description:"Disk size, range from 10Gi to 500Gi"`
}

type VirtualMachine struct {
	Name        string     `json:"name" description:"Virtual machine name"`
	CpuCores    uint32     `json:"cpu_cores" default:"1" description:"Virtual machine cpu cores, range from 1 to 4"`
	Memory      string     `json:"memory" default:"1Gi" description:"Virtual machine memory size, range from 1Gi to 8Gi"`
	Description string     `json:"description,omitempty" description:"Virtual machine description"`
	Image       *ImageSpec `json:"image" description:"Virtual machine image source"`
	Disk        []DiskSpec `json:"disk,omitempty" description:"Virtual machine disks"`
	Guest       *GuestSpec `json:"guest,omitempty" description:"Virtual machine guest operating system"`
}

type VMStatus struct {
	Ready bool   `json:"ready" description:"Virtual machine is ready or not"`
	State string `json:"state" description:"Virtual machine state"`
}

type VirtualMachineResponse struct {
	Name        string     `json:"name" description:"Virtual machine name"`
	ID          string     `json:"id" description:"Virtual machine id"`
	Namespace   string     `json:"namespace" description:"Virtual machine namespace"`
	Description string     `json:"description" description:"Virtual machine description"`
	CpuCores    uint32     `json:"cpu_cores" description:"Virtual machine cpu cores"`
	Memory      string     `json:"memory" description:"Virtual machine memory size"`
	Image       *ImageSpec `json:"image" description:"Virtual machine image source"`
	Disks       []DiskSpec `json:"disks" description:"Virtual machine disks"`
	Status      VMStatus   `json:"status" description:"Virtual machine status"`
}

type ListVirtualMachineResponse struct {
	TotalCount int                      `json:"total_count" description:"Total number of virtual machines"`
	Items      []VirtualMachineResponse `json:"items" description:"List of virtual machines"`
}
