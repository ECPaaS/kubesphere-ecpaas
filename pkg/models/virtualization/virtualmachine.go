/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com
*/

package virtualization

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/api/virtualization/v1alpha1"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
)

type Interface interface {
	CreateVirtualMachine(namespace string, virtz_vm *VirtualMachine) (*v1alpha1.VirtualMachine, error)
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

	// create virtual machine
	vm := v1alpha1.VirtualMachine{}
	vm.Name = virtz_vm.Name
	vm.Spec.Hardware.Domain = v1alpha1.Domain{
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

	v1alpha1VM, err := v.ksclient.VirtualizationV1alpha1().VirtualMachines(namespace).Create(context.Background(), &vm, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return v1alpha1VM, nil
}

func (v *virtualizationOperator) DeleteVirtualMachine(namespace string, name string) (*v1alpha1.VirtualMachine, error) {

	// delete virtual machine
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
