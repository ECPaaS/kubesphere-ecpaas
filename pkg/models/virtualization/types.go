/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com
*/

package virtualization

const (
	vmNamePrefix         = "vm-"   // vm: virtual machine
	diskVolumeNamePrefix = "disk-" // disk: disk volume
	diskVolumeNewPrefix  = "new-"
	imageNamePrefix      = "image-"
)

// Virtual Machine
type VirtualMachineRequest struct {
	Name        string             `json:"name" description:"Virtual machine name"`
	CpuCores    uint32             `json:"cpu_cores" default:"1" description:"Virtual machine cpu cores, range from 1 to 4"`
	Memory      string             `json:"memory" default:"1Gi" description:"Virtual machine memory size, range from 1Gi to 8Gi"`
	Description string             `json:"description" default:"" description:"Virtual machine description"`
	Image       *ImageInfoResponse `json:"image" description:"Virtual machine image source"`
	Disk        []DiskSpec         `json:"disk,omitempty" description:"Virtual machine disks"`
	Guest       *GuestSpec         `json:"guest,omitempty" description:"Virtual machine guest operating system"`
}

type DiskSpec struct {
	Action      string `json:"action" description:"Disk action, the value is 'add' or 'mount'"`
	Name        string `json:"name,omitempty" description:"Disk name"`
	ID          string `json:"id,omitempty" description:"Disk id"`
	Description string `json:"description,omitempty" default:"" description:"Disk description"`
	Namespace   string `json:"namespace,omitempty" description:"Disk namespace"`
	Type        string `json:"type" description:"Disk type, the value is 'system' or 'data'"`
	Size        string `json:"size" default:"20Gi" description:"Disk size, range from 10Gi to 500Gi"`
}

type GuestSpec struct {
	Username string `json:"username" default:"root" description:"Guest operating system username"`
	Password string `json:"password" default:"123456" description:"Guest operating system password"`
}

type ModifyVirtualMachineRequest struct {
	//TODO: support dynamic mount disk
	Name        string `json:"name,omitempty" description:"Virtual machine name"`
	CpuCores    uint32 `json:"cpu_cores,omitempty" default:"1" description:"Virtual machine cpu cores, range from 1 to 4"`
	Memory      string `json:"memory,omitempty" default:"1Gi" description:"Virtual machine memory size, range from 1Gi to 8Gi"`
	Description string `json:"description,omitempty" description:"Virtual machine description"`
}

type VirtualMachineResponse struct {
	ID          string             `json:"id" description:"Virtual machine id"`
	Name        string             `json:"name" description:"Virtual machine name"`
	Namespace   string             `json:"namespace" description:"Virtual machine namespace"`
	Description string             `json:"description" description:"Virtual machine description"`
	CpuCores    uint32             `json:"cpu_cores" description:"Virtual machine cpu cores"`
	Memory      string             `json:"memory" description:"Virtual machine memory size"`
	Image       *ImageInfoResponse `json:"image" description:"Virtual machine image source"`
	Disks       []DiskResponse     `json:"disks" description:"Virtual machine disks"`
	Status      VMStatus           `json:"status" description:"Virtual machine status"`
}

type VirtualMachineIDResponse struct {
	ID string `json:"id" description:"virtual machine id"`
}

type ImageIDResponse struct {
	ID string `json:"id" description:"image id"`
}

type DiskIDResponse struct {
	ID string `json:"id" description:"disk id"`
}

type VMStatus struct {
	Ready bool   `json:"ready" description:"Virtual machine is ready or not"`
	State string `json:"state" description:"Virtual machine state"`
}

type ListVirtualMachineResponse struct {
	TotalCount int                      `json:"total_count" description:"Total number of virtual machines"`
	Items      []VirtualMachineResponse `json:"items" description:"List of virtual machines"`
}

// Disk
type DiskRequest struct {
	Name        string `json:"name" description:"Disk name"`
	Description string `json:"description" default:"" description:"Disk description"`
	Size        string `json:"size" default:"20Gi" description:"Disk size, range from 10Gi to 500Gi"`
}

type ModifyDiskRequest struct {
	Name        string `json:"name,omitempty" description:"Disk name"`
	Description string `json:"description,omitempty" default:"" description:"Disk description"`
	Size        string `json:"size,omitempty" default:"20Gi" description:"Disk size, range from 10Gi to 500Gi, the size only can be increased."`
}

type DiskResponse struct {
	ID          string     `json:"id" description:"Disk id"`
	Name        string     `json:"name" description:"Disk name"`
	Namespace   string     `json:"namespace" description:"Disk namespace"`
	Description string     `json:"description" default:"" description:"Disk description"`
	Type        string     `json:"type" description:"Disk type, the value is 'system' or 'data'"`
	Size        string     `json:"size" default:"20Gi" description:"Disk size, range from 10Gi to 500Gi"`
	Status      DiskStatus `json:"status" description:"Disk status"`
}

type DiskStatus struct {
	Ready bool   `json:"ready" description:"Disk is ready or not"`
	Owner string `json:"owner" description:"Disk owner, if empty, means not owned by any virtual machine"`
}

type ListDiskResponse struct {
	TotalCount int            `json:"total_count" description:"Total number of disks"`
	Items      []DiskResponse `json:"items" description:"List of disks"`
}

// Image
type ImageInfo struct {
	ID        string `json:"id" description:"Image id"`
	Name      string `json:"name" description:"Image name"`
	Namespace string `json:"namespace" description:"Image namespace"`
	System    string `json:"system" description:"Image system"`
	Version   string `json:"version" description:"Image version"`
	ImageSize string `json:"imageSize" description:"Image size"`
	Cpu       string `json:"cpu" description:"cpu used by image"`
	Memory    string `json:"memory" description:"memory used by image"`
}

type ImageInfoResponse struct {
	ID   string `json:"id" description:"Image id"`
	Size string `json:"size" default:"20Gi" description:"Image size, range from 10Gi to 80Gi"`
}

type ImageRequest struct {
	Name           string `json:"name" description:"Image name"`
	OSFamily       string `json:"os_family" default:"ubuntu" description:"Image operating system"`
	Version        string `json:"version" default:"20.04_LTS_64bit" description:"Image version"`
	CpuCores       string `json:"cpu_cores" default:"1" description:"Default image cpu cores, range from 1 to 4"`
	Memory         string `json:"memory" default:"1Gi" description:"Default image memory, range from 1Gi to 8Gi"`
	Size           string `json:"size" default:"20Gi" description:"Default image size, range from 10Gi to 80Gi"`
	Description    string `json:"description" default:"" description:"Image description"`
	UploadFileName string `json:"upload_file_name" default:"" description:"File name which created by upload image api"`
	Shared         bool   `json:"shared" default:"false" description:"Image shared or not"`
}

type ModifyImageRequest struct {
	Name        string `json:"name,omitempty" description:"Image name"`
	CpuCores    string `json:"cpu_cores,omitempty" default:"1" description:"Default image cpu cores, range from 1 to 4"`
	Memory      string `json:"memory,omitempty" default:"1Gi" description:"Default image memory, range from 1Gi to 8Gi"`
	Size        string `json:"size,omitempty" default:"20Gi" description:"Default image size, range from 10Gi to 80Gi, the size only can be increased."`
	Description string `json:"description,omitempty" default:"" description:"Image description"`
	Shared      bool   `json:"shared,omitempty" default:"false" description:"Image shared or not"`
}

type ImageResponse struct {
	ID             string      `json:"id" description:"Image id"`
	Name           string      `json:"name" description:"Image name"`
	Namespace      string      `json:"namespace" description:"Image namespace"`
	OSFamily       string      `json:"os_family" default:"ubuntu" description:"Image operating system"`
	Version        string      `json:"version" default:"20.04_LTS_64bit" description:"Image version"`
	CpuCores       string      `json:"cpu_cores" default:"1" description:"Default image cpu cores, range from 1 to 4"`
	Memory         string      `json:"memory" default:"1Gi" description:"Default image memory, range from 1Gi to 8Gi"`
	Size           string      `json:"size" default:"20Gi" description:"Default image size, range from 10Gi to 80Gi"`
	UploadFileName string      `json:"upload_file_name" description:"File name which created by upload image api"`
	Description    string      `json:"description" default:"" description:"Image description"`
	Shared         bool        `json:"shared" default:"false" description:"Image shared or not"`
	Status         ImageStatus `json:"status" description:"Image status"`
}

type ImageStatus struct {
	Ready bool `json:"ready" description:"Image is ready or not"`
}

type ListImageResponse struct {
	TotalCount int             `json:"total_count" description:"Total number of images"`
	Items      []ImageResponse `json:"items" description:"List of images"`
}
