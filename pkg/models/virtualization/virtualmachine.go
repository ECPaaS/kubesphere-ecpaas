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
	CreateVirtualMachine(namespace string, virtz_vm *VirtualMachine) (*v1alpha1.VirtualMachine, error)
	GetVirtualMachine(namespace string, name string) (*v1alpha1.VirtualMachine, error)
	UpdateVirtualMachine(namespace string, name string, virtz_vm *VirtualMachine) (*v1alpha1.VirtualMachine, error)
	ListVirtualMachine(namespace string) (*v1alpha1.VirtualMachineList, error)
	DeleteVirtualMachine(namespace string, name string) (*v1alpha1.VirtualMachine, error)
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

	if virtz_vm.AddDisk != nil {
		ApplyAddDiskSpec(virtz_vm, &vm)
	}

	if virtz_vm.MountDisk != nil {
		ApplyMountDiskSpec(virtz_vm, &vm)
	}

	v1alpha1VM, err := v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Create(context.Background(), &vm, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return v1alpha1VM, nil
}

func ApplyVMSpec(virtz_vm *VirtualMachine, vm *v1alpha1.VirtualMachine, vm_uuid string) {
	vm.Annotations = make(map[string]string)
	vm.Annotations[v1alpha1.VirtualizationAliasName] = virtz_vm.AliasName
	vm.Annotations[v1alpha1.VirtualizationSystemDiskSize] = virtz_vm.Image.Size
	vm.Annotations[v1alpha1.VirtualizationDescription] = virtz_vm.Description
	vm.Name = "vm-" + vm_uuid

	vm.Spec.Hardware.Domain = v1alpha1.Domain{
		Cpu: v1alpha1.Cpu{
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
	vm.Spec.Hardware.Hostname = virtz_vm.AliasName
}

type ImageInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	System    string `json:"system"`
	Version   string `json:"version"`
	AliasName string `json:"aliasName"`
	ImageSize string `json:"imageSize"`
	Cpu       string `json:"cpu"`
	Memory    string `json:"memory"`
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
				Name: "diskvol-" + vm_uuid,
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
		"diskvol-" + vm_uuid,
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

func ApplyAddDiskSpec(virtz_vm *VirtualMachine, vm *v1alpha1.VirtualMachine) {
	for _, disk := range virtz_vm.AddDisk {
		disk_uuid := uuid.New().String()[:8]
		vm.Spec.DiskVolumeTemplates = append(vm.Spec.DiskVolumeTemplates, v1alpha1.DiskVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name: "diskvol-" + disk_uuid,
				Annotations: map[string]string{
					v1alpha1.VirtualizationAliasName: virtz_vm.AliasName,
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
		vm.Spec.DiskVolumes = append(vm.Spec.DiskVolumes, "diskvol-"+disk_uuid)
	}
}

func ApplyMountDiskSpec(virtz_vm *VirtualMachine, vm *v1alpha1.VirtualMachine) {
	for _, disk := range virtz_vm.MountDisk {
		vm.Spec.DiskVolumes = append(vm.Spec.DiskVolumes, disk.Name)
	}
}

func (v *virtualizationOperator) UpdateVirtualMachine(namespace string, name string, virtz_vm *VirtualMachine) (*v1alpha1.VirtualMachine, error) {
	vm, err := v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if virtz_vm.AliasName != vm.Annotations[v1alpha1.VirtualizationAliasName] {
		vm.Annotations[v1alpha1.VirtualizationAliasName] = virtz_vm.AliasName
	}

	if virtz_vm.Description != vm.Annotations[v1alpha1.VirtualizationDescription] {
		vm.Annotations[v1alpha1.VirtualizationDescription] = virtz_vm.Description
	}

	if virtz_vm.Image != nil {
		return nil, errors.NewBadRequest("Image cannot be updated")
	}

	if virtz_vm.AddDisk != nil {
		return nil, errors.NewBadRequest("AddDisk cannot be updated")
	}

	if virtz_vm.MountDisk != nil {
		return nil, errors.NewBadRequest("MountDisk cannot be updated")
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
