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
	Name        string             `json:"name" description:"Virtual machine name. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"16"`
	CpuCores    uint               `json:"cpu_cores" default:"1" description:"Virtual machine cpu cores" minimum:"1" maximum:"4"`
	Memory      uint               `json:"memory" default:"1" description:"Virtual machine memory size, unit is GB" minimum:"1" maximum:"8"`
	Description string             `json:"description" description:"Virtual machine description. Default is empty string." maximum:"128"`
	Image       *ImageInfoResponse `json:"image" description:"Virtual machine image source"`
	Disk        []DiskSpec         `json:"disk,omitempty" description:"Virtual machine disks"`
	Guest       *GuestSpec         `json:"guest,omitempty" description:"Virtual machine guest operating system"`
}

type DiskSpec struct {
	Action    string `json:"action" description:"Disk action, the value is 'add', 'mount' or 'unmount'"`
	ID        string `json:"id,omitempty" description:"Disk id which is got from disk api"`
	Namespace string `json:"namespace,omitempty" description:"Disk namespace"`
	Size      uint   `json:"size,omitempty" default:"20" description:"Disk size, unit is GB." minimum:"10" maximum:"500"`
}

type ModifyDiskSpec struct {
	Action    string `json:"action" description:"Disk action, the value is 'mount' or 'unmount'"`
	ID        string `json:"id" description:"Disk id which is got from disk api"`
	Namespace string `json:"namespace,omitempty" description:"Disk namespace"`
}

type GuestSpec struct {
	Username string `json:"username" default:"root" description:"Guest operating system username"`
	Password string `json:"password" default:"abc1234" description:"Guest operating system password"`
}

type ModifyVirtualMachineRequest struct {
	Name        string           `json:"name,omitempty" description:"Virtual machine name. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"16"`
	CpuCores    uint             `json:"cpu_cores,omitempty" default:"1" description:"Virtual machine cpu cores." minimum:"1" maximum:"4"`
	Memory      uint             `json:"memory,omitempty" default:"1" description:"Virtual machine memory size, unit is GB." minimum:"1" maximum:"8"`
	Disk        []ModifyDiskSpec `json:"disk,omitempty" description:"Virtual machine disks"`
	Description string           `json:"description,omitempty" description:"Virtual machine description. Can be empty string." maximum:"128"`
}

type VirtualMachineResponse struct {
	ID          string             `json:"id" description:"Virtual machine id"`
	Name        string             `json:"name" description:"Virtual machine name"`
	Namespace   string             `json:"namespace" description:"Virtual machine namespace"`
	Description string             `json:"description" description:"Virtual machine description"`
	CpuCores    uint               `json:"cpu_cores" description:"Virtual machine cpu cores"`
	Memory      uint               `json:"memory" description:"Virtual machine memory size"`
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
	Name        string `json:"name" description:"Disk name. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"16"`
	Description string `json:"description" default:"" description:"Disk description. Default is empty string." maximum:"128"`
	Size        uint   `json:"size" default:"20" description:"Disk size, unit is GB." minimum:"10" maximum:"500"`
}

type ModifyDiskRequest struct {
	Name        string `json:"name,omitempty" description:"Disk name. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"16"`
	Description string `json:"description,omitempty" default:"" description:"Disk description. Can be empty string." maximum:"128"`
	Size        uint   `json:"size,omitempty" default:"20" description:"Disk size, unit is GB and the size only can be increased." minimum:"10" maximum:"500"`
}

type DiskResponse struct {
	ID          string     `json:"id" description:"Disk id"`
	Name        string     `json:"name" description:"Disk name"`
	Namespace   string     `json:"namespace" description:"Disk namespace"`
	Description string     `json:"description" default:"" description:"Disk description"`
	Type        string     `json:"type" description:"Disk type, the value is 'system' or 'data'"`
	Size        uint       `json:"size" default:"20" description:"Disk size, unit is GB" minimum:"10" maximum:"500"`
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
	ID        string `json:"id" description:"Image id which is got from image api"`
	Namespace string `json:"namespace" description:"Image namespace"`
	Size      uint   `json:"size" default:"20" description:"Image size, unit is GB." minimum:"10" maximum:"80"`
}

type ImageRequest struct {
	Name           string `json:"name" description:"Image name. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"16"`
	OSFamily       string `json:"os_family" default:"ubuntu" description:"Image operating system"`
	Version        string `json:"version" default:"20.04_LTS_64bit" description:"Image version"`
	CpuCores       uint   `json:"cpu_cores" default:"1" description:"Default image cpu cores" minimum:"1" maximum:"4"`
	Memory         uint   `json:"memory" default:"1" description:"Default image memory, unit is GB." minimum:"1" maximum:"8"`
	Size           uint   `json:"size" default:"20" description:"Default image size, unit is GB." minimum:"10" maximum:"80"`
	Description    string `json:"description" description:"Image description. Default is empty string." maximum:"128"`
	MinioImageName string `json:"minio_image_name" description:"File name which created by minio image api"`
	Shared         bool   `json:"shared" default:"false" description:"Image shared or not"`
}

type CloneImageRequest struct {
	NewImageName         string `json:"new_img_name" description:"New image name. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"16"`
	SourceImageID        string `json:"src_img_id" description:"Source image id which is got from image api"`
	SourceImageNamespace string `json:"src_img_namespace" description:"Source Image namespace"`
}

type ModifyImageRequest struct {
	Name        string `json:"name,omitempty" description:"Image name. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"16"`
	CpuCores    uint   `json:"cpu_cores,omitempty" default:"1" description:"Default image cpu cores" minimum:"1" maximum:"4"`
	Memory      uint   `json:"memory,omitempty" default:"1" description:"Default image memory, unit is GB." minimum:"1" maximum:"8"`
	Size        uint   `json:"size,omitempty" default:"20" description:"Default image size, unit is GB and the size only can be increased." minimum:"10" maximum:"80"`
	Description string `json:"description,omitempty" default:"" description:"Image description. Can be empty string." maximum:"128"`
	Shared      bool   `json:"shared,omitempty" default:"false" description:"Image shared or not"`
}

type ImageResponse struct {
	ID             string      `json:"id" description:"Image id"`
	Name           string      `json:"name" description:"Image name"`
	Namespace      string      `json:"namespace" description:"Image namespace"`
	OSFamily       string      `json:"os_family" default:"ubuntu" description:"Image operating system"`
	Version        string      `json:"version" default:"20.04_LTS_64bit" description:"Image version"`
	CpuCores       uint        `json:"cpu_cores" default:"1" description:"Default image cpu cores" minimum:"1" maximum:"4"`
	Memory         uint        `json:"memory" default:"1" description:"Default image memory, unit is GB" minimum:"1" maximum:"8"`
	Size           uint        `json:"size" default:"20" description:"Default image size, unit is GB" minimum:"10" maximum:"80"`
	MinioImageName string      `json:"minio_image_name" description:"File name which created by minio image api"`
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

type VirtualizationResourceQuota struct {
	Namespace      string `json:"namespace" description:"Resource Quota namespace"`
	Disk           int    `json:"diskCount"`
	File           int    `json:"fileCount"`
	Image          int    `json:"imageCount"`
	VirtualMachine int    `json:"virtualMachineCount"`
}
