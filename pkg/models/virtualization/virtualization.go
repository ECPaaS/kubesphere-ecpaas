/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com
*/

package virtualization

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"kubesphere.io/api/virtualization/v1alpha1"
	kvapi "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/kubecli"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
)

var bucketName = "ecpaas-images"

type Interface interface {
	// VirtualMachine
	CreateVirtualMachine(namespace string, ui_vm *VirtualMachineRequest) (*v1alpha1.VirtualMachine, error)
	GetVirtualMachine(namespace string, name string) (*v1alpha1.VirtualMachine, error)
	UpdateVirtualMachine(namespace string, name string, ui_vm *ModifyVirtualMachineRequest) (*v1alpha1.VirtualMachine, error)
	StartVirtualMachine(namespace string, name string) (*v1alpha1.VirtualMachine, error)
	StopVirtualMachine(namespace string, name string) (*v1alpha1.VirtualMachine, error)
	ListVirtualMachine(namespace string) (*v1alpha1.VirtualMachineList, error)
	DeleteVirtualMachine(namespace string, name string) (*v1alpha1.VirtualMachine, error)
	// Disk
	CreateDisk(namespace string, ui_disk *DiskRequest) (*v1alpha1.DiskVolume, error)
	UpdateDisk(namespace string, name string, ui_disk *ModifyDiskRequest) (*v1alpha1.DiskVolume, error)
	GetDisk(namespace string, name string) (*v1alpha1.DiskVolume, error)
	ListDisk(namespace string) (*v1alpha1.DiskVolumeList, error)
	DeleteDisk(namespace string, name string) (*v1alpha1.DiskVolume, error)
	// Image
	CreateImage(namespace string, ui_image *ImageRequest) (*v1alpha1.ImageTemplate, error)
	CloneImage(namespace string, ui_clone_image *CloneImageRequest) (*v1alpha1.ImageTemplate, error)
	UpdateImage(namespace string, name string, ui_image *ModifyImageRequest) (*v1alpha1.ImageTemplate, error)
	GetImage(namespace string, name string) (*v1alpha1.ImageTemplate, error)
	ListImage(namespace string) (*v1alpha1.ImageTemplateList, error)
	DeleteImage(namespace string, name string) (*v1alpha1.ImageTemplate, error)
}

type virtualizationOperator struct {
	ksclient  kubesphere.Interface
	k8sclient kubernetes.Interface
}

func New(ksclient kubesphere.Interface, k8sclient kubernetes.Interface) Interface {
	return &virtualizationOperator{
		ksclient:  ksclient,
		k8sclient: k8sclient,
	}
}

func (v *virtualizationOperator) CreateVirtualMachine(namespace string, ui_vm *VirtualMachineRequest) (*v1alpha1.VirtualMachine, error) {
	vm := v1alpha1.VirtualMachine{}
	vm_uuid := uuid.New().String()[:8]
	vm.Namespace = namespace

	ApplyVMSpec(ui_vm, &vm, vm_uuid)

	if ui_vm.Image != nil {
		imagetemplate, err := v.ksclient.VirtualizationV1alpha1().ImageTemplates(ui_vm.Image.Namespace).Get(context.Background(), ui_vm.Image.ID, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		imageType := imagetemplate.Labels[v1alpha1.VirtualizationImageType]

		if strings.ToLower(imageType) == "cloud" {
			err = ApplyCloudImageSpec(ui_vm, &vm, imagetemplate, namespace, vm_uuid)
			if err != nil {
				return nil, err
			}
		} else if strings.ToLower(imageType) == "iso" {
			err = ApplyISOImageSpec(ui_vm, &vm, imagetemplate, namespace, vm_uuid)
			if err != nil {
				return nil, err
			}
		}
	}

	err := ApplyVMDiskSpec(ui_vm, &vm)
	if err != nil {
		return nil, err
	}

	for _, disk := range ui_vm.Disk {
		if disk.Action == "mount" {
			diskVolume, err := v.ksclient.VirtualizationV1alpha1().DiskVolumes(namespace).Get(context.Background(), disk.ID, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			if !IsDiskVolumeOwnerLabelEmpty(diskVolume) {
				return nil, fmt.Errorf("disk %s is used by vm %s", disk.ID, diskVolume.Labels[v1alpha1.VirtualizationDiskVolumeOwner])
			}
		}
	}

	v1alpha1VM, err := v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Create(context.Background(), &vm, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return v1alpha1VM, nil
}

func ApplyVMSpec(ui_vm *VirtualMachineRequest, vm *v1alpha1.VirtualMachine, vm_uuid string) {
	vm.Annotations = make(map[string]string)
	vm.Annotations[v1alpha1.VirtualizationAliasName] = ui_vm.Name
	vm.Annotations[v1alpha1.VirtualizationDescription] = ui_vm.Description
	vm.Annotations[v1alpha1.VirtualizationSystemDiskSize] = strconv.FormatUint(uint64(ui_vm.Image.Size), 10)
	vm.Name = vmNamePrefix + vm_uuid

	memory := strconv.FormatUint(uint64(ui_vm.Memory), 10) + "Gi"
	vm.Spec.Hardware.Domain = v1alpha1.DomainSpec{
		CPU: v1alpha1.CPU{
			Cores: uint32(ui_vm.CpuCores),
		},
		Devices: kvapi.Devices{
			Interfaces: []kvapi.Interface{
				{ // network interface
					Name: "default",
					InterfaceBindingMethod: kvapi.InterfaceBindingMethod{
						Masquerade: &kvapi.InterfaceMasquerade{},
					},
				},
			},
		},
		Resources: v1alpha1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceMemory: resource.MustParse(memory),
			},
		},
	}
	vm.Spec.Hardware.Networks = []v1alpha1.Network{
		{
			Name: "default",
			NetworkSource: v1alpha1.NetworkSource{
				Pod: &v1alpha1.PodNetwork{},
			},
		},
	}
	vm.Spec.Hardware.Hostname = strings.ToLower(ui_vm.Name)
	vm.Spec.RunStrategy = v1alpha1.VirtualMachineRunStrategyAlways
}

func ApplyCloudImageSpec(ui_vm *VirtualMachineRequest, vm *v1alpha1.VirtualMachine, imagetemplate *v1alpha1.ImageTemplate, namespace string, vm_uuid string) error {

	imageInfo := ImageInfo{}
	imageInfo.ID = imagetemplate.Name
	imageInfo.Namespace = imagetemplate.Namespace
	// annotations
	imageInfo.Name = imagetemplate.Annotations[v1alpha1.VirtualizationAliasName]
	// labels
	imageInfo.System = imagetemplate.Labels[v1alpha1.VirtualizationOSFamily]
	imageInfo.Version = imagetemplate.Labels[v1alpha1.VirtualizationOSVersion]
	imageInfo.ImageSize = imagetemplate.Labels[v1alpha1.VirtualizationImageStorage]
	imageInfo.Cpu = imagetemplate.Labels[v1alpha1.VirtualizationCpuCores]
	imageInfo.Memory = imagetemplate.Labels[v1alpha1.VirtualizationImageMemory]

	jsonData, err := json.Marshal(imageInfo)
	if err != nil {
		return err
	}

	vm.Annotations[v1alpha1.VirtualizationImageInfo] = string(jsonData)

	size := strconv.FormatUint(uint64(ui_vm.Image.Size), 10) + "Gi"
	vm.Spec.DiskVolumeTemplates = []v1alpha1.DiskVolume{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: diskVolumeNamePrefix + vm_uuid,
				Annotations: map[string]string{
					v1alpha1.VirtualizationAliasName: ui_vm.Name,
				},
				Labels: map[string]string{
					v1alpha1.VirtualizationBootOrder:        "1",
					v1alpha1.VirtualizationDiskType:         "system",
					v1alpha1.VirtualizationDiskHotpluggable: "false",
					v1alpha1.VirtualizationDiskMode:         "rw",
				},
				Namespace: namespace,
			},
			Spec: v1alpha1.DiskVolumeSpec{
				Source: v1alpha1.DiskVolumeSource{
					Image: &v1alpha1.DataVolumeSourceImage{
						Namespace: imagetemplate.Namespace,
						Name:      ui_vm.Image.ID,
					},
				},
				Resources: v1alpha1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: resource.MustParse(size),
					},
				},
			},
		},
	}
	vm.Spec.DiskVolumes = []string{
		diskVolumeNamePrefix + vm_uuid,
	}

	vm.Annotations[v1alpha1.VirtualizationSystemDiskName] = diskVolumeNamePrefix + vm_uuid

	username := "root"
	password := "abc1234"

	if ui_vm.Guest != nil {
		username = ui_vm.Guest.Username
		password = ui_vm.Guest.Password
	}

	userDataString := `#cloud-config
	updates:
	  network:
		when: ['boot']
	timezone: Asia/Taipei
	packages:
	 - cloud-init
	package_update: true
	ssh_pwauth: true
	disable_root: false
	chpasswd: {"list":"` + username + `:` + password + `",expire: False}
	runcmd:
	 - sed -i "/PermitRootLogin/s/^.*$/PermitRootLogin yes/g" /etc/ssh/sshd_config
	 - systemctl restart sshd.service
	 `
	// remote tab character and space
	userDataString = strings.Replace(userDataString, "\t", "", -1)

	userDataBytes := []byte(userDataString)
	encodedBase64userData := base64.StdEncoding.EncodeToString(userDataBytes)

	vm.Spec.Hardware.Volumes = []kvapi.Volume{
		{
			Name: "cloudinitdisk",
			VolumeSource: kvapi.VolumeSource{
				CloudInitNoCloud: &kvapi.CloudInitNoCloudSource{
					UserDataBase64: encodedBase64userData,
				},
			},
		},
	}

	return nil
}

func ApplyISOImageSpec(ui_vm *VirtualMachineRequest, vm *v1alpha1.VirtualMachine, imagetemplate *v1alpha1.ImageTemplate, namespace string, vm_uuid string) error {

	osFamily := imagetemplate.Labels[v1alpha1.VirtualizationOSFamily]

	if strings.ToLower(osFamily) == "windows" {
		err := ApplyWindowsISOImageSpec(ui_vm, vm, imagetemplate, namespace, vm_uuid)
		if err != nil {
			return err
		}
	} else {
		err := ApplyLinuxISOImageSpec(ui_vm, vm, imagetemplate, namespace, vm_uuid)
		if err != nil {
			return err
		}
	}

	return nil
}

func ApplyWindowsISOImageSpec(ui_vm *VirtualMachineRequest, vm *v1alpha1.VirtualMachine, imagetemplate *v1alpha1.ImageTemplate, namespace string, vm_uuid string) error {

	imageInfo := ImageInfo{}
	imageInfo.ID = imagetemplate.Name
	imageInfo.Namespace = imagetemplate.Namespace
	// annotations
	imageInfo.Name = imagetemplate.Annotations[v1alpha1.VirtualizationAliasName]
	// labels
	imageInfo.System = imagetemplate.Labels[v1alpha1.VirtualizationOSFamily]
	imageInfo.Version = imagetemplate.Labels[v1alpha1.VirtualizationOSVersion]
	imageInfo.ImageSize = imagetemplate.Labels[v1alpha1.VirtualizationImageStorage]
	imageInfo.Cpu = imagetemplate.Labels[v1alpha1.VirtualizationCpuCores]
	imageInfo.Memory = imagetemplate.Labels[v1alpha1.VirtualizationImageMemory]

	jsonData, err := json.Marshal(imageInfo)
	if err != nil {
		return err
	}

	vm.Annotations[v1alpha1.VirtualizationImageInfo] = string(jsonData)

	systemSize := strconv.FormatUint(uint64(ui_vm.Image.Size), 10) + "Gi"
	imageSize, _ := strconv.ParseUint(imagetemplate.Labels[v1alpha1.VirtualizationImageStorage], 10, 32)
	cdromSize := strconv.FormatUint(imageSize, 10) + "Gi"

	var spinlocksRetries uint32 = 8191

	vm.Spec.DiskVolumeTemplates = []v1alpha1.DiskVolume{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: diskVolumeNamePrefix + vm_uuid,
				Annotations: map[string]string{
					v1alpha1.VirtualizationAliasName: ui_vm.Name,
				},
				Labels: map[string]string{
					v1alpha1.VirtualizationBootOrder:        "1",
					v1alpha1.VirtualizationDiskType:         "system",
					v1alpha1.VirtualizationDiskHotpluggable: "false",
					v1alpha1.VirtualizationDiskMode:         "rw",
				},
				Namespace: namespace,
			},
			Spec: v1alpha1.DiskVolumeSpec{
				Source: v1alpha1.DiskVolumeSource{
					Blank: &v1alpha1.DataVolumeBlankImage{},
				},
				Resources: v1alpha1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: resource.MustParse(systemSize),
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: diskVolumeNamePrefix + vm_uuid + diskVolumeNameSuffix,
				Annotations: map[string]string{
					v1alpha1.VirtualizationAliasName: ui_vm.Name,
				},
				Labels: map[string]string{
					v1alpha1.VirtualizationBootOrder:          "2",
					v1alpha1.VirtualizationDiskType:           "data",
					v1alpha1.VirtualizationDiskMediaType:      "cdrom",
					v1alpha1.VirtualizationDiskHotpluggable:   "false",
					v1alpha1.VirtualizationDiskMode:           "ro",
					v1alpha1.VirtualizationDiskMinioImageName: imagetemplate.Labels[v1alpha1.VirtualizationDiskMinioImageName],
				},
				Namespace: namespace,
			},
			Spec: v1alpha1.DiskVolumeSpec{
				Source: v1alpha1.DiskVolumeSource{
					Image: &v1alpha1.DataVolumeSourceImage{
						Namespace: imagetemplate.Namespace,
						Name:      ui_vm.Image.ID,
					},
				},
				Resources: v1alpha1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: resource.MustParse(cdromSize),
					},
				},
			},
		},
	}
	vm.Spec.DiskVolumes = []string{
		diskVolumeNamePrefix + vm_uuid,
		diskVolumeNamePrefix + vm_uuid + diskVolumeNameSuffix,
	}

	vm.Spec.Hardware.Domain.Features = &kvapi.Features{
		ACPI: kvapi.FeatureState{},
		APIC: &kvapi.FeatureAPIC{},
		Hyperv: &kvapi.FeatureHyperv{
			Relaxed: &kvapi.FeatureState{},
			VAPIC:   &kvapi.FeatureState{},
			Spinlocks: &kvapi.FeatureSpinlocks{
				Retries: &spinlocksRetries,
			},
		},
	}

	vm.Spec.Hardware.Domain.Devices.Disks = []kvapi.Disk{
		{
			Name: "virtiocontainer",
			DiskDevice: kvapi.DiskDevice{
				CDRom: &kvapi.CDRomTarget{
					Bus: "sata",
				},
			},
		},
	}

	vm.Spec.Hardware.Volumes = []kvapi.Volume{
		{
			Name: "virtiocontainer",
			VolumeSource: kvapi.VolumeSource{
				ContainerDisk: &kvapi.ContainerDiskSource{
					Image: "kubevirt/virtio-container-disk",
				},
			},
		},
	}

	return nil
}

func ApplyLinuxISOImageSpec(ui_vm *VirtualMachineRequest, vm *v1alpha1.VirtualMachine, imagetemplate *v1alpha1.ImageTemplate, namespace string, vm_uuid string) error {
	imageInfo := ImageInfo{}
	imageInfo.ID = imagetemplate.Name
	imageInfo.Namespace = imagetemplate.Namespace
	// annotations
	imageInfo.Name = imagetemplate.Annotations[v1alpha1.VirtualizationAliasName]
	// labels
	imageInfo.System = imagetemplate.Labels[v1alpha1.VirtualizationOSFamily]
	imageInfo.Version = imagetemplate.Labels[v1alpha1.VirtualizationOSVersion]
	imageInfo.ImageSize = imagetemplate.Labels[v1alpha1.VirtualizationImageStorage]
	imageInfo.Cpu = imagetemplate.Labels[v1alpha1.VirtualizationCpuCores]
	imageInfo.Memory = imagetemplate.Labels[v1alpha1.VirtualizationImageMemory]

	jsonData, err := json.Marshal(imageInfo)
	if err != nil {
		return err
	}

	vm.Annotations[v1alpha1.VirtualizationImageInfo] = string(jsonData)

	systemSize := strconv.FormatUint(uint64(ui_vm.Image.Size), 10) + "Gi"
	imageSize, _ := strconv.ParseUint(imagetemplate.Labels[v1alpha1.VirtualizationImageStorage], 10, 32)
	cdromSize := strconv.FormatUint(imageSize, 10) + "Gi"

	vm.Spec.DiskVolumeTemplates = []v1alpha1.DiskVolume{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: diskVolumeNamePrefix + vm_uuid,
				Annotations: map[string]string{
					v1alpha1.VirtualizationAliasName: ui_vm.Name,
				},
				Labels: map[string]string{
					v1alpha1.VirtualizationBootOrder:        "1",
					v1alpha1.VirtualizationDiskType:         "system",
					v1alpha1.VirtualizationDiskHotpluggable: "false",
					v1alpha1.VirtualizationDiskMode:         "rw",
				},
				Namespace: namespace,
			},
			Spec: v1alpha1.DiskVolumeSpec{
				Source: v1alpha1.DiskVolumeSource{
					Blank: &v1alpha1.DataVolumeBlankImage{},
				},
				Resources: v1alpha1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: resource.MustParse(systemSize),
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: diskVolumeNamePrefix + vm_uuid + diskVolumeNameSuffix,
				Annotations: map[string]string{
					v1alpha1.VirtualizationAliasName: ui_vm.Name,
				},
				Labels: map[string]string{
					v1alpha1.VirtualizationBootOrder:        "2",
					v1alpha1.VirtualizationDiskType:         "data",
					v1alpha1.VirtualizationDiskMediaType:    "cdrom",
					v1alpha1.VirtualizationDiskHotpluggable: "false",
					v1alpha1.VirtualizationDiskMode:         "ro",
				},
				Namespace: namespace,
			},
			Spec: v1alpha1.DiskVolumeSpec{
				Source: v1alpha1.DiskVolumeSource{
					Image: &v1alpha1.DataVolumeSourceImage{
						Namespace: imagetemplate.Namespace,
						Name:      ui_vm.Image.ID,
					},
				},
				Resources: v1alpha1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: resource.MustParse(cdromSize),
					},
				},
			},
		},
	}
	vm.Spec.DiskVolumes = []string{
		diskVolumeNamePrefix + vm_uuid,
		diskVolumeNamePrefix + vm_uuid + diskVolumeNameSuffix,
	}

	return nil
}

func ApplyVMDiskSpec(ui_vm *VirtualMachineRequest, vm *v1alpha1.VirtualMachine) error {
	for _, uiDisk := range ui_vm.Disk {
		if uiDisk.Action == "add" {
			ApplyAddDisk(ui_vm, vm, &uiDisk)
		} else if uiDisk.Action == "mount" {
			err := ApplyMountDisk(vm, &uiDisk)
			if err != nil {
				klog.Errorf("mount disk error: %v", err)
				return err
			}
		}
	}

	return nil
}

func ApplyAddDisk(ui_vm *VirtualMachineRequest, vm *v1alpha1.VirtualMachine, uiDisk *DiskSpec) {
	uiDisk.Namespace = vm.Namespace
	diskVolume := AddDiskVolume(ui_vm.Name, uiDisk)
	vm.Spec.DiskVolumeTemplates = append(vm.Spec.DiskVolumeTemplates, diskVolume)
	vm.Spec.DiskVolumes = append(vm.Spec.DiskVolumes, diskVolume.Name)
}

func AddDiskVolume(diskVolumeName string, uiDisk *DiskSpec) v1alpha1.DiskVolume {
	disk_uuid := uuid.New().String()[:8]

	diskVolume := v1alpha1.DiskVolume{}
	diskVolume.Name = diskVolumeNamePrefix + disk_uuid
	diskVolume.Namespace = uiDisk.Namespace
	diskVolume.Annotations = map[string]string{
		v1alpha1.VirtualizationAliasName: diskVolumeName,
	}
	diskVolume.Labels = map[string]string{
		v1alpha1.VirtualizationDiskHotpluggable: "true",
		v1alpha1.VirtualizationDiskType:         "data",
		v1alpha1.VirtualizationDiskMode:         "rw",
	}

	size := strconv.FormatUint(uint64(uiDisk.Size), 10) + "Gi"
	diskVolume.Spec.Source.Blank = &v1alpha1.DataVolumeBlankImage{}
	res := v1.ResourceList{}
	res[v1.ResourceStorage] = resource.MustParse(size)
	diskVolume.Spec.Resources.Requests = res

	return diskVolume
}

func ApplyMountDisk(vm *v1alpha1.VirtualMachine, uiDisk *DiskSpec) error {
	if uiDisk.Namespace != vm.Namespace {
		return fmt.Errorf("disk namespace is not equal to vm namespace")
	}

	vm.Spec.DiskVolumes = append(vm.Spec.DiskVolumes, uiDisk.ID)

	return nil
}

func ApplyUnmountDisk(vm *v1alpha1.VirtualMachine, uiDisk *DiskSpec) error {
	found := false

	for i, disk := range vm.Spec.DiskVolumes {
		if disk == uiDisk.ID {
			found = true
			vm.Spec.DiskVolumes = append(vm.Spec.DiskVolumes[:i], vm.Spec.DiskVolumes[i+1:]...)
		}
	}

	if !found {
		return fmt.Errorf("disk %s not found", uiDisk.ID)
	}

	return nil
}

func IsDiskVolumeOwnerLabelEmpty(diskVolume *v1alpha1.DiskVolume) bool {
	if diskVolume.Labels == nil {
		return true
	}

	if diskVolume.Labels[v1alpha1.VirtualizationDiskVolumeOwner] == "" {
		return true
	}

	return false
}

func ConvertModifyDiskSpecToDiskSpec(modifyDiskSpec *ModifyDiskSpec) *DiskSpec {
	diskSpec := DiskSpec{}

	diskSpec.ID = modifyDiskSpec.ID
	diskSpec.Size = 0
	diskSpec.Action = modifyDiskSpec.Action
	diskSpec.Namespace = modifyDiskSpec.Namespace

	return &diskSpec
}

func (v *virtualizationOperator) UpdateVirtualMachine(namespace string, name string, ui_vm *ModifyVirtualMachineRequest) (*v1alpha1.VirtualMachine, error) {
	vm, err := v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if ui_vm.Name != "" && ui_vm.Name != vm.Annotations[v1alpha1.VirtualizationAliasName] {
		vm.Annotations[v1alpha1.VirtualizationAliasName] = ui_vm.Name
	}

	if ui_vm.Description != nil {
		vm.Annotations[v1alpha1.VirtualizationDescription] = *ui_vm.Description
	}

	if ui_vm.CpuCores != 0 && ui_vm.CpuCores != uint(vm.Spec.Hardware.Domain.CPU.Cores) {
		vm.Spec.Hardware.Domain.CPU.Cores = uint32(ui_vm.CpuCores)
	}

	if ui_vm.Memory != 0 && ui_vm.Memory != uint(vm.Spec.Hardware.Domain.Resources.Requests.Memory().Size()) {
		vm.Spec.Hardware.Domain.Resources.Requests[v1.ResourceMemory] =
			resource.MustParse(strconv.FormatUint(uint64(ui_vm.Memory), 10) + "Gi")
	}

	// TODO: update image size
	// if ui_vm.Image.Size != "" && ui_vm.Image.Size != vm.Annotations[v1alpha1.VirtualizationSystemDiskSize] {
	// 	vm.Annotations[v1alpha1.VirtualizationSystemDiskSize] = ui_vm.Image.Size
	// }

	if ui_vm.Disk != nil {
		for _, uiDisk := range ui_vm.Disk {
			if uiDisk.Action == "mount" {
				err := ApplyMountDisk(vm, ConvertModifyDiskSpecToDiskSpec(&uiDisk))
				if err != nil {
					klog.Errorf("mount disk error: %v", err)
					return nil, err
				}

				diskVolume, err := v.ksclient.VirtualizationV1alpha1().DiskVolumes(namespace).Get(context.Background(), uiDisk.ID, metav1.GetOptions{})
				if err != nil {
					return nil, err
				}
				if !IsDiskVolumeOwnerLabelEmpty(diskVolume) {
					return nil, fmt.Errorf("disk %s is used by vm %s", uiDisk.ID, diskVolume.Labels[v1alpha1.VirtualizationDiskVolumeOwner])
				}

			} else if uiDisk.Action == "unmount" {
				err := ApplyUnmountDisk(vm, ConvertModifyDiskSpecToDiskSpec(&uiDisk))
				if err != nil {
					klog.Errorf("unmount disk error: %v", err)
					return nil, err
				}
			}
		}
	}

	updated_vm, err := v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Update(context.Background(), vm, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return updated_vm, nil
}

func (v *virtualizationOperator) StartVirtualMachine(namespace string, name string) (*v1alpha1.VirtualMachine, error) {
	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})

	// get the kubevirt client, using which kubevirt resources can be managed.
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	err = virtClient.VirtualMachine(namespace).Start(name, &kvapi.StartOptions{})
	if err != nil {
		return nil, err
	}

	vm, err := v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if vm.Spec.RunStrategy != v1alpha1.VirtualMachineRunStrategyAlways {
		vm.Spec.RunStrategy = v1alpha1.VirtualMachineRunStrategyAlways
		_, err = v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Update(context.Background(), vm, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}
	}

	return vm, nil
}

func (v *virtualizationOperator) StopVirtualMachine(namespace string, name string) (*v1alpha1.VirtualMachine, error) {
	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})

	// get the kubevirt client, using which kubevirt resources can be managed.
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	err = virtClient.VirtualMachine(namespace).Stop(name, &kvapi.StopOptions{})
	if err != nil {
		return nil, err
	}

	vm, err := v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if vm.Spec.RunStrategy != v1alpha1.VirtualMachineRunStrategyHalted {
		vm.Spec.RunStrategy = v1alpha1.VirtualMachineRunStrategyHalted
		_, err = v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Update(context.Background(), vm, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}
	}

	return vm, nil
}

func (v *virtualizationOperator) GetVirtualMachine(namespace string, name string) (*v1alpha1.VirtualMachine, error) {
	vm, err := v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return vm, nil
}

func (v *virtualizationOperator) ListVirtualMachine(namespace string) (*v1alpha1.VirtualMachineList, error) {
	vmList, err := v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return vmList, nil
}

func (v *virtualizationOperator) DeleteVirtualMachine(namespace string, name string) (*v1alpha1.VirtualMachine, error) {
	vm, err := v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	err = v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return nil, err
	}

	return vm, nil
}

func (v *virtualizationOperator) GetDisk(namespace string, name string) (*v1alpha1.DiskVolume, error) {
	diskVolume, err := v.ksclient.VirtualizationV1alpha1().DiskVolumes(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return diskVolume, nil
}

func (v *virtualizationOperator) ListDisk(namespace string) (*v1alpha1.DiskVolumeList, error) {
	diskVolumelist, err := v.ksclient.VirtualizationV1alpha1().DiskVolumes(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return diskVolumelist, nil
}

func (v *virtualizationOperator) CreateDisk(namespace string, ui_disk *DiskRequest) (*v1alpha1.DiskVolume, error) {
	diskVolume := v1alpha1.DiskVolume{}
	disk_uuid := uuid.New().String()[:8]

	diskVolume.Name = diskVolumeNamePrefix + disk_uuid
	diskVolume.Annotations = map[string]string{
		v1alpha1.VirtualizationAliasName:          ui_disk.Name,
		v1alpha1.VirtualizationDescription:        ui_disk.Description,
		v1alpha1.VirtualizationDiskMinioImageName: "",
	}
	diskVolume.Labels = map[string]string{
		v1alpha1.VirtualizationDiskType: "data",
		v1alpha1.VirtualizationDiskMode: "rw",
	}

	size := strconv.FormatUint(uint64(ui_disk.Size), 10) + "Gi"
	diskVolume.Spec.PVCName = diskVolumeNewPrefix + diskVolume.Name
	diskVolume.Spec.Source.Blank = &v1alpha1.DataVolumeBlankImage{}
	res := v1.ResourceList{}
	res[v1.ResourceStorage] = resource.MustParse(size)
	diskVolume.Spec.Resources.Requests = res

	createdDisk, err := v.ksclient.VirtualizationV1alpha1().DiskVolumes(namespace).Create(context.Background(), &diskVolume, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return createdDisk, nil
}

func (v *virtualizationOperator) UpdateDisk(namespace string, name string, ui_disk *ModifyDiskRequest) (*v1alpha1.DiskVolume, error) {
	diskVolume, err := v.ksclient.VirtualizationV1alpha1().DiskVolumes(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if ui_disk.Name != "" && ui_disk.Name != diskVolume.Annotations[v1alpha1.VirtualizationAliasName] {
		diskVolume.Annotations[v1alpha1.VirtualizationAliasName] = ui_disk.Name
	}

	size := strconv.FormatUint(uint64(ui_disk.Size), 10) + "Gi"
	if ui_disk.Size != 0 && resource.MustParse(size) != diskVolume.Spec.Resources.Requests[v1.ResourceStorage] {
		diskVolume.Spec.Resources.Requests[v1.ResourceStorage] = resource.MustParse(size)
	}

	if ui_disk.Description != nil {
		diskVolume.Annotations[v1alpha1.VirtualizationDescription] = *ui_disk.Description
	}

	updatedDisk, err := v.ksclient.VirtualizationV1alpha1().DiskVolumes(namespace).Update(context.Background(), diskVolume, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return updatedDisk, nil
}

func (v *virtualizationOperator) DeleteDisk(namespace string, name string) (*v1alpha1.DiskVolume, error) {
	diskVolume, err := v.ksclient.VirtualizationV1alpha1().DiskVolumes(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	err = v.ksclient.VirtualizationV1alpha1().DiskVolumes(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return nil, err
	}

	return diskVolume, nil
}

func (v *virtualizationOperator) CreateImage(namespace string, ui_image *ImageRequest) (*v1alpha1.ImageTemplate, error) {
	imageTemplate := v1alpha1.ImageTemplate{}

	imageTemplate.Name = imageNamePrefix + uuid.New().String()[:8]
	imageTemplate.Namespace = namespace
	imageTemplate.Annotations = map[string]string{
		v1alpha1.VirtualizationAliasName:   ui_image.Name,
		v1alpha1.VirtualizationDescription: ui_image.Description,
	}

	imageTemplate.Labels = map[string]string{
		v1alpha1.VirtualizationOSFamily:           ui_image.OSFamily,
		v1alpha1.VirtualizationOSVersion:          ui_image.Version,
		v1alpha1.VirtualizationImageMemory:        strconv.FormatUint(uint64(ui_image.Memory), 10),
		v1alpha1.VirtualizationCpuCores:           strconv.FormatUint(uint64(ui_image.CpuCores), 10),
		v1alpha1.VirtualizationImageStorage:       strconv.FormatUint(uint64(ui_image.Size), 10),
		v1alpha1.VirtualizationDiskMinioImageName: ui_image.MinioImageName,
		v1alpha1.VirtualizationImageType:          ui_image.Type,
	}

	// get minio ip and port
	minioServiceName := "minio"

	serviceList, err := v.k8sclient.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Warning("Failed to get service: ", err)
		return nil, err
	}

	var minioService *v1.Service

	for _, service := range serviceList.Items {
		if service.Name == minioServiceName {
			minioService = &service
			break
		}
	}

	if minioService == nil {
		klog.Warning("Cannot find the minio service ", err)
		return nil, err
	}

	ip := minioService.Spec.ClusterIP
	port := minioService.Spec.Ports[0].Port

	// image template spec
	imageTemplate.Spec.Source = v1alpha1.ImageTemplateSource{
		HTTP: &cdiv1.DataVolumeSourceHTTP{
			URL: "http://" + ip + ":" + strconv.Itoa(int(port)) + "/" + bucketName + "/" + ui_image.MinioImageName,
		},
	}
	imageTemplate.Spec.Attributes = v1alpha1.ImageTemplateAttributes{
		Public: ui_image.Shared,
	}
	size := strconv.FormatUint(uint64(ui_image.Size), 10)
	imageTemplate.Spec.Resources.Requests = v1.ResourceList{
		v1.ResourceStorage: resource.MustParse(size + "Gi"),
	}

	createdImage, err := v.ksclient.VirtualizationV1alpha1().ImageTemplates(namespace).Create(context.Background(), &imageTemplate, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return createdImage, nil
}

func (v *virtualizationOperator) CloneImage(namespace string, ui_clone_image *CloneImageRequest) (*v1alpha1.ImageTemplate, error) {
	imageTemplate := v1alpha1.ImageTemplate{}

	// get source image
	sourceImage, err := v.ksclient.VirtualizationV1alpha1().ImageTemplates(ui_clone_image.SourceImageNamespace).Get(context.Background(), ui_clone_image.SourceImageID, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// forbid cloning image if source image is not shared
	if !sourceImage.Spec.Attributes.Public {
		return nil, fmt.Errorf("source image '%s' is not shared", sourceImage.Name)
	}

	if namespace == ui_clone_image.SourceImageNamespace {
		return nil, fmt.Errorf("cannot clone image to the same namespace")
	}

	// clone image
	imageTemplate.Name = imageNamePrefix + uuid.New().String()[:8]
	imageTemplate.Namespace = namespace
	imageTemplate.Annotations = map[string]string{
		v1alpha1.VirtualizationAliasName: ui_clone_image.NewImageName,
	}
	imageTemplate.Labels = map[string]string{
		v1alpha1.VirtualizationOSFamily:           sourceImage.Labels[v1alpha1.VirtualizationOSFamily],
		v1alpha1.VirtualizationOSVersion:          sourceImage.Labels[v1alpha1.VirtualizationOSVersion],
		v1alpha1.VirtualizationImageMemory:        sourceImage.Labels[v1alpha1.VirtualizationImageMemory],
		v1alpha1.VirtualizationCpuCores:           sourceImage.Labels[v1alpha1.VirtualizationCpuCores],
		v1alpha1.VirtualizationImageStorage:       sourceImage.Labels[v1alpha1.VirtualizationImageStorage],
		v1alpha1.VirtualizationDiskMinioImageName: sourceImage.Labels[v1alpha1.VirtualizationDiskMinioImageName],
		v1alpha1.VirtualizationImageType:          sourceImage.Labels[v1alpha1.VirtualizationImageType],
	}

	imageTemplate.Spec.Attributes = v1alpha1.ImageTemplateAttributes{
		Public: false,
	}

	imageTemplate.Spec.Resources = sourceImage.Spec.Resources

	imageTemplate.Spec.Source = v1alpha1.ImageTemplateSource{
		Clone: &cdiv1.DataVolumeSourcePVC{
			Name:      sourceImage.Name,
			Namespace: sourceImage.Namespace,
		},
	}

	createdImage, err := v.ksclient.VirtualizationV1alpha1().ImageTemplates(namespace).Create(context.Background(), &imageTemplate, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return createdImage, nil
}

func (v *virtualizationOperator) UpdateImage(namespace string, name string, ui_image *ModifyImageRequest) (*v1alpha1.ImageTemplate, error) {
	imageTemplate, err := v.ksclient.VirtualizationV1alpha1().ImageTemplates(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if ui_image.Name != "" && ui_image.Name != imageTemplate.Annotations[v1alpha1.VirtualizationAliasName] {
		imageTemplate.Annotations[v1alpha1.VirtualizationAliasName] = ui_image.Name
	}

	cpuCores, _ := strconv.ParseUint(imageTemplate.Labels[v1alpha1.VirtualizationCpuCores], 10, 32)
	if ui_image.CpuCores != 0 && ui_image.CpuCores != uint(cpuCores) {
		imageTemplate.Labels[v1alpha1.VirtualizationCpuCores] = strconv.FormatUint(uint64(ui_image.CpuCores), 10)
	}

	memory, _ := strconv.ParseUint(imageTemplate.Labels[v1alpha1.VirtualizationImageMemory], 10, 32)
	if ui_image.Memory != 0 && ui_image.Memory != uint(memory) {
		imageTemplate.Labels[v1alpha1.VirtualizationImageMemory] = strconv.FormatUint(uint64(ui_image.Memory), 10)
	}

	size, _ := strconv.ParseUint(imageTemplate.Labels[v1alpha1.VirtualizationImageStorage], 10, 32)
	if ui_image.Size != 0 && ui_image.Size != uint(size) {
		imageTemplate.Labels[v1alpha1.VirtualizationImageStorage] = strconv.FormatUint(uint64(ui_image.Size), 10)
	}

	if ui_image.Description != nil {
		imageTemplate.Annotations[v1alpha1.VirtualizationDescription] = *ui_image.Description
	}

	if ui_image.Shared != imageTemplate.Spec.Attributes.Public {
		imageTemplate.Spec.Attributes.Public = ui_image.Shared
	}

	updatedImage, err := v.ksclient.VirtualizationV1alpha1().ImageTemplates(namespace).Update(context.Background(), imageTemplate, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return updatedImage, nil
}

func (v *virtualizationOperator) GetImage(namespace string, name string) (*v1alpha1.ImageTemplate, error) {
	imageTemplate, err := v.ksclient.VirtualizationV1alpha1().ImageTemplates(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return imageTemplate, nil
}

func (v *virtualizationOperator) ListImage(namespace string) (*v1alpha1.ImageTemplateList, error) {
	imageTemplateList, err := v.ksclient.VirtualizationV1alpha1().ImageTemplates(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return imageTemplateList, nil
}

func (v *virtualizationOperator) DeleteImage(namespace string, name string) (*v1alpha1.ImageTemplate, error) {
	imageTemplate, err := v.ksclient.VirtualizationV1alpha1().ImageTemplates(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	err = v.ksclient.VirtualizationV1alpha1().ImageTemplates(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return nil, err
	}

	return imageTemplate, nil
}
