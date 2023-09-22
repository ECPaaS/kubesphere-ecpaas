/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com
*/

package virtualization

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/api/virtualization/v1alpha1"
	kvapi "kubevirt.io/api/core/v1"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
)

type Interface interface {
	// VirtualMachine
	CreateVirtualMachine(namespace string, virtz_vm *VirtualMachine) (*v1alpha1.VirtualMachine, error)
	GetVirtualMachine(namespace string, name string) (*v1alpha1.VirtualMachine, error)
	UpdateVirtualMachine(namespace string, name string, virtz_vm *VirtualMachine) (*v1alpha1.VirtualMachine, error)
	ListVirtualMachine(namespace string) (*v1alpha1.VirtualMachineList, error)
	DeleteVirtualMachine(namespace string, name string) (*v1alpha1.VirtualMachine, error)
	// DiskVolume
	GetDiskVolume(namespace string, name string) (*v1alpha1.DiskVolume, error)
	ListDiskVolume(namespace string) (*v1alpha1.DiskVolumeList, error)
}

type virtualizationOperator struct {
	ksclient kubesphere.Interface
}

func New(ksclient kubesphere.Interface) Interface {
	return &virtualizationOperator{
		ksclient: ksclient,
	}
}

func (v *virtualizationOperator) CreateVirtualMachine(namespace string, virtz_vm *VirtualMachine) (*v1alpha1.VirtualMachine, error) {
	vm := v1alpha1.VirtualMachine{}
	vm_uuid := uuid.New().String()[:8]

	ApplyVMSpec(virtz_vm, &vm, vm_uuid)

	if virtz_vm.Image != nil {
		imagetemplate, err := v.ksclient.VirtualizationV1alpha1().ImageTemplates(namespace).Get(context.Background(), virtz_vm.Image.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		err = ApplyImageSpec(virtz_vm, &vm, imagetemplate, namespace, vm_uuid)
		if err != nil {
			return nil, err
		}
	}

	ApplyVMDiskSpec(virtz_vm, &vm)

	v1alpha1VM, err := v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Create(context.Background(), &vm, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return v1alpha1VM, nil
}

func ApplyVMSpec(virtz_vm *VirtualMachine, vm *v1alpha1.VirtualMachine, vm_uuid string) {
	vm.Annotations = make(map[string]string)
	vm.Annotations[v1alpha1.VirtualizationAliasName] = virtz_vm.Name
	vm.Annotations[v1alpha1.VirtualizationDescription] = virtz_vm.Description
	vm.Name = "vm-" + vm_uuid

	vm.Spec.Hardware.Domain = v1alpha1.Domain{
		CPU: v1alpha1.CPU{
			Cores: virtz_vm.CpuCores,
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
				v1.ResourceMemory: resource.MustParse(virtz_vm.Memory),
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
	vm.Spec.Hardware.Hostname = virtz_vm.Name
}

func ApplyImageSpec(virtz_vm *VirtualMachine, vm *v1alpha1.VirtualMachine, imagetemplate *v1alpha1.ImageTemplate, namespace string, vm_uuid string) error {

	imageInfo := ImageInfo{}
	imageInfo.Name = imagetemplate.Name
	imageInfo.Namespace = imagetemplate.Namespace
	// annotations
	imageInfo.AliasName = imagetemplate.Annotations[v1alpha1.VirtualizationAliasName]
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
					v1alpha1.VirtualizationAliasName: virtz_vm.Name,
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
						Name:      virtz_vm.Image.Name,
					},
				},
				Resources: v1alpha1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: resource.MustParse(virtz_vm.Image.Size),
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

	if virtz_vm.Guest != nil {
		username = virtz_vm.Guest.Username
		password = virtz_vm.Guest.Password
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

func ApplyVMDiskSpec(virtz_vm *VirtualMachine, vm *v1alpha1.VirtualMachine) {
	for _, disk := range virtz_vm.Disk {
		if disk.Name == "" {
			ApplyAddDiskSpec(virtz_vm, vm, &disk)
		} else {
			ApplyMountDiskSpec(virtz_vm, vm, &disk)
		}
	}
}

func ApplyAddDiskSpec(virtz_vm *VirtualMachine, vm *v1alpha1.VirtualMachine, disk *DiskSpec) {
	disk_uuid := uuid.New().String()[:8]
	vm.Spec.DiskVolumeTemplates = append(vm.Spec.DiskVolumeTemplates, v1alpha1.DiskVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: diskVolumeNamePrefix + disk_uuid,
			Annotations: map[string]string{
				v1alpha1.VirtualizationAliasName: virtz_vm.Name,
			},
			Labels: map[string]string{
				v1alpha1.VirtualizationDiskType: "data",
			},
		},
		Spec: v1alpha1.DiskVolumeSpec{
			Source: v1alpha1.DiskVolumeSource{
				Blank: &v1alpha1.DataVolumeBlankImage{},
			},
			Resources: v1alpha1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(disk.Size),
				},
			},
		},
	})
	vm.Spec.DiskVolumes = append(vm.Spec.DiskVolumes, diskVolumeNamePrefix+disk_uuid)
}

func ApplyMountDiskSpec(virtz_vm *VirtualMachine, vm *v1alpha1.VirtualMachine, disk *DiskSpec) {
	vm.Spec.DiskVolumes = append(vm.Spec.DiskVolumes, disk.Name)
}

func (v *virtualizationOperator) UpdateVirtualMachine(namespace string, name string, virtz_vm *VirtualMachine) (*v1alpha1.VirtualMachine, error) {
	vm, err := v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if virtz_vm.Name != vm.Annotations[v1alpha1.VirtualizationAliasName] {
		vm.Annotations[v1alpha1.VirtualizationAliasName] = virtz_vm.Name
	}

	if virtz_vm.Description != vm.Annotations[v1alpha1.VirtualizationDescription] {
		vm.Annotations[v1alpha1.VirtualizationDescription] = virtz_vm.Description
	}

	if virtz_vm.CpuCores != vm.Spec.Hardware.Domain.CPU.Cores {
		vm.Spec.Hardware.Domain.CPU.Cores = virtz_vm.CpuCores
	}

	if virtz_vm.Memory != vm.Spec.Hardware.Domain.Resources.Requests.Memory().String() {
		vm.Spec.Hardware.Domain.Resources.Requests[v1.ResourceMemory] = resource.MustParse(virtz_vm.Memory)
	}

	if virtz_vm.Image != nil {
		return nil, errors.NewBadRequest("Image cannot be updated")
	}

	v1alpha1VM, err := v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Update(context.Background(), vm, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return v1alpha1VM, nil
}

func (v *virtualizationOperator) GetVirtualMachine(namespace string, name string) (*v1alpha1.VirtualMachine, error) {
	v1alpha1VM, err := v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return v1alpha1VM, nil
}

func (v *virtualizationOperator) ListVirtualMachine(namespace string) (*v1alpha1.VirtualMachineList, error) {
	v1alpha1VMlist, err := v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return v1alpha1VMlist, nil
}

func (v *virtualizationOperator) DeleteVirtualMachine(namespace string, name string) (*v1alpha1.VirtualMachine, error) {
	v1alpha1VM, err := v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	err = v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return nil, err
	}

	return v1alpha1VM, nil
}

func (v *virtualizationOperator) GetDiskVolume(namespace string, name string) (*v1alpha1.DiskVolume, error) {
	v1alpha1DiskVolume, err := v.ksclient.VirtualizationV1alpha1().DiskVolumes(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return v1alpha1DiskVolume, nil
}

func (v *virtualizationOperator) ListDiskVolume(namespace string) (*v1alpha1.DiskVolumeList, error) {
	v1alpha1DiskVolumelist, err := v.ksclient.VirtualizationV1alpha1().DiskVolumes(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return v1alpha1DiskVolumelist, nil
}
