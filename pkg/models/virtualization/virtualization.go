/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com
*/

package virtualization

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"kubesphere.io/api/virtualization/v1alpha1"
	kvapi "kubevirt.io/api/core/v1"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
)

var bucketName = "ecpaas-images"

type Interface interface {
	// VirtualMachine
	CreateVirtualMachine(namespace string, ui_vm *VirtualMachineRequest) (*v1alpha1.VirtualMachine, error)
	GetVirtualMachine(namespace string, name string) (*v1alpha1.VirtualMachine, error)
	UpdateVirtualMachine(namespace string, name string, ui_vm *ModifyVirtualMachineRequest) (*v1alpha1.VirtualMachine, error)
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

	ApplyVMSpec(ui_vm, &vm, vm_uuid)

	if ui_vm.Image != nil {
		imagetemplate, err := v.ksclient.VirtualizationV1alpha1().ImageTemplates(namespace).Get(context.Background(), ui_vm.Image.ID, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		err = ApplyImageSpec(ui_vm, &vm, imagetemplate, namespace, vm_uuid)
		if err != nil {
			return nil, err
		}
	}

	ApplyVMDiskSpec(ui_vm, &vm)

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
	vm.Annotations[v1alpha1.VirtualizationSystemDiskSize] = ui_vm.Image.Size
	vm.Name = vmNamePrefix + vm_uuid

	vm.Spec.Hardware.Domain = v1alpha1.Domain{
		CPU: v1alpha1.CPU{
			Cores: ui_vm.CpuCores,
		},
		Devices: v1alpha1.Devices{
			Interfaces: []v1alpha1.Interface{
				{ // network interface
					Name: "default",
					InterfaceBindingMethod: v1alpha1.InterfaceBindingMethod{
						Masquerade: &v1alpha1.InterfaceMasquerade{},
					},
				},
			},
		},
		Resources: v1alpha1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceMemory: resource.MustParse(ui_vm.Memory),
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
	vm.Spec.Hardware.Hostname = ui_vm.Name
}

func ApplyImageSpec(ui_vm *VirtualMachineRequest, vm *v1alpha1.VirtualMachine, imagetemplate *v1alpha1.ImageTemplate, namespace string, vm_uuid string) error {

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

	vm.Spec.DiskVolumeTemplates = []v1alpha1.DiskVolume{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: diskVolumeNamePrefix + vm_uuid,
				Annotations: map[string]string{
					v1alpha1.VirtualizationAliasName: ui_vm.Name,
				},
				Labels: map[string]string{
					v1alpha1.VirtualizationBootOrder: "1",
					v1alpha1.VirtualizationDiskType:  "system",
				},
			},
			Spec: v1alpha1.DiskVolumeSpec{
				Source: v1alpha1.DiskVolumeSource{
					Image: &v1alpha1.DataVolumeSourceImage{
						Namespace: namespace,
						Name:      ui_vm.Image.ID,
					},
				},
				Resources: v1alpha1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: resource.MustParse(ui_vm.Image.Size),
					},
				},
			},
		},
	}
	vm.Spec.DiskVolumes = []string{
		diskVolumeNamePrefix + vm_uuid,
	}

	username := "root"
	password := "123456"

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

func ApplyVMDiskSpec(ui_vm *VirtualMachineRequest, vm *v1alpha1.VirtualMachine) {
	for _, uiDisk := range ui_vm.Disk {
		if uiDisk.Action == "add" {
			ApplyAddDisk(ui_vm, vm, &uiDisk)
		} else if uiDisk.Action == "mount" {
			ApplyMountDisk(vm, &uiDisk)
		}
	}
}

func ApplyAddDisk(ui_vm *VirtualMachineRequest, vm *v1alpha1.VirtualMachine, uiDisk *DiskSpec) {
	diskVolume := AddDiskVolume(ui_vm.Name, uiDisk)
	vm.Spec.DiskVolumeTemplates = append(vm.Spec.DiskVolumeTemplates, diskVolume)
	vm.Spec.DiskVolumes = append(vm.Spec.DiskVolumes, diskVolume.Name)
}

func AddDiskVolume(diskVolumeName string, uiDisk *DiskSpec) v1alpha1.DiskVolume {
	disk_uuid := uuid.New().String()[:8]

	diskVolume := v1alpha1.DiskVolume{}
	diskVolume.Name = diskVolumeNamePrefix + disk_uuid
	diskVolume.Annotations = map[string]string{
		v1alpha1.VirtualizationAliasName:   diskVolumeName,
		v1alpha1.VirtualizationDescription: uiDisk.Description,
	}
	diskVolume.Labels = map[string]string{
		v1alpha1.VirtualizationDiskType: uiDisk.Type,
	}

	diskVolume.Spec.Source.Blank = &v1alpha1.DataVolumeBlankImage{}
	res := v1.ResourceList{}
	res[v1.ResourceStorage] = resource.MustParse(uiDisk.Size)
	diskVolume.Spec.Resources.Requests = res

	return diskVolume
}

func ApplyMountDisk(vm *v1alpha1.VirtualMachine, uiDisk *DiskSpec) {
	vm.Spec.DiskVolumes = append(vm.Spec.DiskVolumes, uiDisk.ID)
}

func (v *virtualizationOperator) UpdateVirtualMachine(namespace string, name string, ui_vm *ModifyVirtualMachineRequest) (*v1alpha1.VirtualMachine, error) {
	vm, err := v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if ui_vm.Name != "" && ui_vm.Name != vm.Annotations[v1alpha1.VirtualizationAliasName] {
		vm.Annotations[v1alpha1.VirtualizationAliasName] = ui_vm.Name
	}

	if ui_vm.Description != "" && ui_vm.Description != vm.Annotations[v1alpha1.VirtualizationDescription] {
		vm.Annotations[v1alpha1.VirtualizationDescription] = ui_vm.Description
	}

	if ui_vm.CpuCores != 0 && ui_vm.CpuCores != vm.Spec.Hardware.Domain.CPU.Cores {
		vm.Spec.Hardware.Domain.CPU.Cores = ui_vm.CpuCores
	}

	if ui_vm.Memory != "" && ui_vm.Memory != vm.Spec.Hardware.Domain.Resources.Requests.Memory().String() {
		vm.Spec.Hardware.Domain.Resources.Requests[v1.ResourceMemory] = resource.MustParse(ui_vm.Memory)
	}

	// TODO: update image size
	// if ui_vm.Image.Size != "" && ui_vm.Image.Size != vm.Annotations[v1alpha1.VirtualizationSystemDiskSize] {
	// 	vm.Annotations[v1alpha1.VirtualizationSystemDiskSize] = ui_vm.Image.Size
	// }

	updated_vm, err := v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Update(context.Background(), vm, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return updated_vm, nil
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
		v1alpha1.VirtualizationAliasName:   ui_disk.Name,
		v1alpha1.VirtualizationDescription: ui_disk.Description,
	}
	diskVolume.Labels = map[string]string{
		v1alpha1.VirtualizationDiskType: "data",
	}

	diskVolume.Spec.PVCName = diskVolumeNewPrefix + diskVolume.Name
	diskVolume.Spec.Source.Blank = &v1alpha1.DataVolumeBlankImage{}
	res := v1.ResourceList{}
	res[v1.ResourceStorage] = resource.MustParse(ui_disk.Size)
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

	if ui_disk.Size != "" && ui_disk.Size != diskVolume.Spec.Resources.Requests.Storage().String() {
		diskVolume.Spec.Resources.Requests[v1.ResourceStorage] = resource.MustParse(ui_disk.Size)
	}

	if ui_disk.Description != "" && ui_disk.Description != diskVolume.Annotations[v1alpha1.VirtualizationDescription] {
		diskVolume.Annotations[v1alpha1.VirtualizationDescription] = ui_disk.Description
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
		v1alpha1.VirtualizationOSFamily:       ui_image.OSFamily,
		v1alpha1.VirtualizationOSVersion:      ui_image.Version,
		v1alpha1.VirtualizationImageMemory:    ui_image.Memory,
		v1alpha1.VirtualizationCpuCores:       ui_image.CpuCores,
		v1alpha1.VirtualizationImageStorage:   ui_image.Size,
		v1alpha1.VirtualizationUploadFileName: ui_image.UploadFileName,
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
			URL: "http://" + ip + ":" + strconv.Itoa(int(port)) + "/" + bucketName + "/" + ui_image.UploadFileName,
		},
	}
	imageTemplate.Spec.Attributes = v1alpha1.ImageTemplateAttributes{
		Public: ui_image.Shared,
	}
	imageTemplate.Spec.Resources.Requests = v1.ResourceList{
		v1.ResourceStorage: resource.MustParse(ui_image.Size),
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

	if ui_image.CpuCores != "" && ui_image.CpuCores != imageTemplate.Labels[v1alpha1.VirtualizationCpuCores] {
		imageTemplate.Labels[v1alpha1.VirtualizationCpuCores] = ui_image.CpuCores
	}

	if ui_image.Memory != "" && ui_image.Memory != imageTemplate.Labels[v1alpha1.VirtualizationImageMemory] {
		imageTemplate.Labels[v1alpha1.VirtualizationImageMemory] = ui_image.Memory
	}

	if ui_image.Size != "" && ui_image.Size != imageTemplate.Labels[v1alpha1.VirtualizationImageStorage] {
		imageTemplate.Labels[v1alpha1.VirtualizationImageStorage] = ui_image.Size
	}

	if ui_image.Description != "" && ui_image.Description != imageTemplate.Annotations[v1alpha1.VirtualizationDescription] {
		imageTemplate.Annotations[v1alpha1.VirtualizationDescription] = ui_image.Description
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
