/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com
*/

package virtualization

import (
	"context"
	"encoding/json"
	"strconv"

	fakek8s "k8s.io/client-go/kubernetes/fake"
	fakeks "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"

	vm_ctrl "kubesphere.io/kubesphere/pkg/controller/virtualization/virtualmachine"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	virtzv1alpha1 "kubesphere.io/api/virtualization/v1alpha1"
	ui_virtz "kubesphere.io/kubesphere/pkg/models/virtualization"
)

// prepare fake disk volume
func prepareFakeDiskVolume(ksClient *fakeks.Clientset, vm_instance *virtzv1alpha1.VirtualMachine) error {

	for _, diskVolumeTemplate := range vm_instance.Spec.DiskVolumeTemplates {
		disk := vm_ctrl.GenerateDiskVolume(vm_instance, &diskVolumeTemplate)

		_, err := ksClient.VirtualizationV1alpha1().DiskVolumes(vm_instance.Namespace).Create(context.Background(), disk, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

type FakeImageTemplate struct {
	Name      string
	Namespace string
	Size      uint
}

func prepareFakeImageTemplate(ksClient *fakeks.Clientset, fakeImageTemlate *FakeImageTemplate) error {
	imageName := fakeImageTemlate.Name
	imageNamespace := fakeImageTemlate.Namespace

	size := strconv.FormatUint(uint64(fakeImageTemlate.Size), 10)
	imagetemplate := &virtzv1alpha1.ImageTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      imageName,
			Namespace: imageNamespace,
		},
		Spec: virtzv1alpha1.ImageTemplateSpec{
			Source: virtzv1alpha1.ImageTemplateSource{
				HTTP: &cdiv1.DataVolumeSourceHTTP{
					URL: "http://test.com",
				},
			},
			Resources: virtzv1alpha1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(size + "Gi"),
				},
			},
		},
	}

	_, err := ksClient.VirtualizationV1alpha1().ImageTemplates(fakeImageTemlate.Namespace).Create(context.TODO(), imagetemplate, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func prepareFakeVirtualMachine(ksClient *fakeks.Clientset) (*virtzv1alpha1.VirtualMachine, error) {

	diskVolumeNamePrefix := "disk-"

	vm_uuid := "1234"
	namespace := "default"
	vm_name := "testvm"
	vm_id := "vm-" + vm_uuid

	vm := &virtzv1alpha1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vm_id,
			Namespace: namespace,
			Annotations: map[string]string{
				virtzv1alpha1.VirtualizationAliasName:      vm_name,
				virtzv1alpha1.VirtualizationDescription:    vm_name,
				virtzv1alpha1.VirtualizationSystemDiskSize: "10Gi",
			},
		},
	}

	imageInfo := ui_virtz.ImageInfo{}
	imageInfo.ID = "image-1234"
	imageInfo.Namespace = namespace
	// annotations
	imageInfo.Name = "image-test"
	// labels
	imageInfo.System = "ubuntu"
	imageInfo.Version = "20.04_LTS_64bit"
	imageInfo.ImageSize = "20Gi"
	imageInfo.Cpu = "1"
	imageInfo.Memory = "1Gi"

	jsonData, err := json.Marshal(imageInfo)
	if err != nil {
		return nil, err
	}

	vm.Annotations[virtzv1alpha1.VirtualizationImageInfo] = string(jsonData)

	vm.Spec.Hardware.Domain = virtzv1alpha1.Domain{
		CPU: virtzv1alpha1.CPU{
			Cores: 2,
		},
		Resources: virtzv1alpha1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceMemory: resource.MustParse("2Gi"),
			},
		},
	}

	vm.Spec.Hardware.Networks = []virtzv1alpha1.Network{
		{
			Name: "default",
			NetworkSource: virtzv1alpha1.NetworkSource{
				Pod: &virtzv1alpha1.PodNetwork{},
			},
		},
	}
	vm.Spec.Hardware.Hostname = vm_name

	vm.Spec.DiskVolumeTemplates = []virtzv1alpha1.DiskVolume{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: diskVolumeNamePrefix + vm_uuid,
				Annotations: map[string]string{
					virtzv1alpha1.VirtualizationAliasName: vm_name,
				},
				Labels: map[string]string{
					virtzv1alpha1.VirtualizationBootOrder: "1",
					virtzv1alpha1.VirtualizationDiskType:  "system",
				},
			},
			Spec: virtzv1alpha1.DiskVolumeSpec{
				Source: virtzv1alpha1.DiskVolumeSource{
					Image: &virtzv1alpha1.DataVolumeSourceImage{
						Namespace: namespace,
						Name:      "image-1234",
					},
				},
				Resources: virtzv1alpha1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: resource.MustParse("10Gi"),
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: diskVolumeNamePrefix + vm_uuid + "-new",
				Annotations: map[string]string{
					virtzv1alpha1.VirtualizationAliasName: vm_name,
				},
				Labels: map[string]string{
					virtzv1alpha1.VirtualizationDiskType: "data",
				},
			},
			Spec: virtzv1alpha1.DiskVolumeSpec{
				Source: virtzv1alpha1.DiskVolumeSource{
					Blank: &virtzv1alpha1.DataVolumeBlankImage{},
				},
				Resources: virtzv1alpha1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: resource.MustParse("20Gi"),
					},
				},
			},
		},
	}
	vm.Spec.DiskVolumes = []string{
		diskVolumeNamePrefix + vm_uuid,
		diskVolumeNamePrefix + vm_uuid + "-new",
	}

	_, err = ksClient.VirtualizationV1alpha1().VirtualMachines(namespace).Create(context.TODO(), vm, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return vm, nil
}

func prepareFakeMinioService(k8sClient *fakek8s.Clientset) error {
	minioService := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "minio",
			Namespace: "default",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: "minio",
					Port: 9000,
				},
			},
			ClusterIP: "1.2.3.4",
		},
	}

	_, err := k8sClient.CoreV1().Services("default").Create(context.Background(), minioService, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}
