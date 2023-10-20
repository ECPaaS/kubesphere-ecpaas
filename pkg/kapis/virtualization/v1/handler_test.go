/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com
*/

package virtualization

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/emicklei/go-restful"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	fakeks "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"

	virtzv1alpha1 "kubesphere.io/api/virtualization/v1alpha1"
	ui_virtz "kubesphere.io/kubesphere/pkg/models/virtualization"
)

func TestGetVirtualMachine(t *testing.T) {

	ksClient := fakeks.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset()
	handler := newHandler(ksClient, k8sClient)

	namespace := "default"

	// prepare a fake image template
	image := &FakeImageTemplate{
		Name:      "image-1234",
		Namespace: namespace,
		Size:      20,
	}

	prepareFakeImageTemplate(ksClient, image)

	// prepare a fake virtual machine
	ui_virtz_vm := ui_virtz.VirtualMachineRequest{}
	ui_virtz_vm.Name = "testvm"
	ui_virtz_vm.Description = "testvm"
	ui_virtz_vm.Image = &ui_virtz.ImageInfoResponse{
		ID:   "image-1234",
		Size: 20,
	}
	ui_virtz_vm.CpuCores = 1
	ui_virtz_vm.Memory = 1

	vm, err := createVirtualMachine(&handler, &ui_virtz_vm, namespace)
	if err != nil {
		t.Error(err)
	}

	vmID := vm.Name
	url := fmt.Sprintf("/namespaces/%s/virtualmachine/%s", namespace, vmID)

	request := httptest.NewRequest("GET", url, nil)
	restfulRequest := restful.NewRequest(request)

	pathMap := make(map[string]string)
	pathMap["namespace"] = namespace
	pathMap["id"] = vmID
	if err := reflectutils.SetUnExportedField(restfulRequest, "pathParameters", pathMap); err != nil {
		t.Fatalf("set pathParameters failed")
	}

	recorder := httptest.NewRecorder()
	restfulResponse := restful.NewResponse(recorder)
	restfulResponse.SetRequestAccepts("application/json")

	handler.GetVirtualMachine(restfulRequest, restfulResponse)
	if status := restfulResponse.StatusCode(); status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	res := restfulResponse.ResponseWriter.(*httptest.ResponseRecorder)

	var vmResponse ui_virtz.VirtualMachineResponse
	err = json.Unmarshal(res.Body.Bytes(), &vmResponse)
	if err != nil {
		t.Error(err)
	}

	if vmResponse.ID != vmID {
		t.Errorf("vm id is not correct: got %v want %v", vmResponse.ID, vmID)
	}

	if vmResponse.Name != ui_virtz_vm.Name {
		t.Errorf("vm name is not correct: got %v want %v", vmResponse.Name, ui_virtz_vm.Name)
	}

	if vmResponse.CpuCores != ui_virtz_vm.CpuCores {
		t.Errorf("vm cpu cores is not correct: got %v want %v", vmResponse.CpuCores, 2)
	}

	if vmResponse.Memory != ui_virtz_vm.Memory {
		t.Errorf("vm memory is not correct: got %v want %v", vmResponse.Memory, 2)
	}

	if vmResponse.Description != ui_virtz_vm.Description {
		t.Errorf("vm description is not correct: got %v want %v", vmResponse.Description, ui_virtz_vm.Description)
	}

}

func TestGetVirtualMachineWithAddDisk(t *testing.T) {

	ksClient := fakeks.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset()
	handler := newHandler(ksClient, k8sClient)

	namespace := "default"

	// prepare a fake image template
	image := &FakeImageTemplate{
		Name:      "image-1234",
		Namespace: namespace,
		Size:      20,
	}

	prepareFakeImageTemplate(ksClient, image)

	// prepare a fake virtual machine
	ui_virtz_vm := ui_virtz.VirtualMachineRequest{}
	ui_virtz_vm.Name = "testvm"
	ui_virtz_vm.Description = "testvm"
	ui_virtz_vm.Image = &ui_virtz.ImageInfoResponse{
		ID:   "image-1234",
		Size: 20,
	}
	ui_virtz_vm.Disk = []ui_virtz.DiskSpec{
		{
			Action: "add",
			Size:   20,
		},
	}
	ui_virtz_vm.CpuCores = 1
	ui_virtz_vm.Memory = 1

	vm, err := createVirtualMachine(&handler, &ui_virtz_vm, namespace)
	if err != nil {
		t.Error(err)
	}

	prepareFakeDiskVolume(ksClient, vm)

	vmName := vm.Name
	url := fmt.Sprintf("/namespaces/%s/virtualmachine/%s", namespace, vmName)

	request := httptest.NewRequest("GET", url, nil)
	restfulRequest := restful.NewRequest(request)

	pathMap := make(map[string]string)
	pathMap["namespace"] = namespace
	pathMap["id"] = vmName
	if err := reflectutils.SetUnExportedField(restfulRequest, "pathParameters", pathMap); err != nil {
		t.Fatalf("set pathParameters failed")
	}

	recorder := httptest.NewRecorder()
	restfulResponse := restful.NewResponse(recorder)
	restfulResponse.SetRequestAccepts("application/json")

	handler.GetVirtualMachine(restfulRequest, restfulResponse)
	if status := restfulResponse.StatusCode(); status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	res := restfulResponse.ResponseWriter.(*httptest.ResponseRecorder)

	var vmResponse ui_virtz.VirtualMachineResponse
	err = json.Unmarshal(res.Body.Bytes(), &vmResponse)
	if err != nil {
		t.Error(err)
	}

	// verify disk
	if len(vmResponse.Disks) != 2 {
		t.Errorf("vm disk number is not correct: got %v want %v", len(vmResponse.Disks), 2)
	}

	for _, disk := range vmResponse.Disks {
		if disk.Type == "system" {
			if disk.Size != ui_virtz_vm.Image.Size {
				t.Errorf("vm disk size is not correct: got %v want %v", disk.Size, ui_virtz_vm.Image.Size)
			}
		} else if disk.Type == "data" {
			if disk.Size != ui_virtz_vm.Disk[0].Size {
				t.Errorf("vm disk size is not correct: got %v want %v", disk.Size, ui_virtz_vm.Disk[0].Size)
			}
		}
	}

}

func TestPostVirtualMachine(t *testing.T) {

	ksClient := fakeks.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset()
	handler := newHandler(ksClient, k8sClient)

	// prepare a fake image template
	fakeImageTemlate := &FakeImageTemplate{
		Name:      "image-1234",
		Namespace: "default",
		Size:      20,
	}
	prepareFakeImageTemplate(ksClient, fakeImageTemlate)

	// verify post virtual machine request
	namespace := "default"
	url := fmt.Sprintf("/namespaces/%s/virtualmachine", namespace)

	vmRequest := ui_virtz.VirtualMachineRequest{
		Name:     "testvm",
		CpuCores: 2,
		Memory:   2,
		Image: &ui_virtz.ImageInfoResponse{
			ID:   fakeImageTemlate.Name,
			Size: 20,
		},
	}

	vmRequestBodyBytes, err := json.Marshal(vmRequest)
	if err != nil {
		t.Fatalf("Error marshaling JSON: %v", err)
	}

	request := httptest.NewRequest("POST", url, bytes.NewReader(vmRequestBodyBytes))
	request.Header.Set("Content-Type", "application/json")
	restfulRequest := restful.NewRequest(request)

	pathMap := make(map[string]string)
	pathMap["namespace"] = namespace
	if err := reflectutils.SetUnExportedField(restfulRequest, "pathParameters", pathMap); err != nil {
		t.Fatalf("set pathParameters failed")
	}

	recorder := httptest.NewRecorder()
	restfulResponse := restful.NewResponse(recorder)
	restfulResponse.SetRequestAccepts("application/json")

	handler.CreateVirtualMahcine(restfulRequest, restfulResponse)
	if status := restfulResponse.StatusCode(); status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	res := restfulResponse.ResponseWriter.(*httptest.ResponseRecorder)

	var vmIDResponse ui_virtz.VirtualMachineIDResponse
	err = json.Unmarshal(res.Body.Bytes(), &vmIDResponse)
	if err != nil {
		t.Error(err)
	}

	// get virtual machine from fake ks client
	vm, err := ksClient.VirtualizationV1alpha1().VirtualMachines(namespace).Get(context.Background(), vmIDResponse.ID, metav1.GetOptions{})
	if err != nil {
		t.Error(err)
	}

	checkVirtualMachineResult(t, vm, vmRequest)

}

func TestPostVirtualMachineWithAddDisk(t *testing.T) {

	ksClient := fakeks.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset()
	handler := newHandler(ksClient, k8sClient)

	// prepare a fake image template
	fakeImageTemlate := &FakeImageTemplate{
		Name:      "image-1234",
		Namespace: "default",
		Size:      20,
	}
	prepareFakeImageTemplate(ksClient, fakeImageTemlate)

	// verify post virtual machine request
	namespace := "default"
	url := fmt.Sprintf("/namespaces/%s/virtualmachine", namespace)

	vmRequest := ui_virtz.VirtualMachineRequest{
		Name:     "testvm",
		CpuCores: 2,
		Memory:   2,
		Image: &ui_virtz.ImageInfoResponse{
			ID:   fakeImageTemlate.Name,
			Size: 20,
		},
		Disk: []ui_virtz.DiskSpec{
			{
				Action: "add",
				Size:   20,
			},
		},
	}

	vmRequestBodyBytes, err := json.Marshal(vmRequest)
	if err != nil {
		t.Fatalf("Error marshaling JSON: %v", err)
	}

	request := httptest.NewRequest("POST", url, bytes.NewReader(vmRequestBodyBytes))
	request.Header.Set("Content-Type", "application/json")
	restfulRequest := restful.NewRequest(request)

	pathMap := make(map[string]string)
	pathMap["namespace"] = namespace
	if err := reflectutils.SetUnExportedField(restfulRequest, "pathParameters", pathMap); err != nil {
		t.Fatalf("set pathParameters failed")
	}

	recorder := httptest.NewRecorder()
	restfulResponse := restful.NewResponse(recorder)
	restfulResponse.SetRequestAccepts("application/json")

	handler.CreateVirtualMahcine(restfulRequest, restfulResponse)
	if status := restfulResponse.StatusCode(); status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	res := restfulResponse.ResponseWriter.(*httptest.ResponseRecorder)

	var vmIDResponse ui_virtz.VirtualMachineIDResponse
	err = json.Unmarshal(res.Body.Bytes(), &vmIDResponse)
	if err != nil {
		t.Error(err)
	}

	// get virtual machine from fake ks client
	vm, err := ksClient.VirtualizationV1alpha1().VirtualMachines(namespace).Get(context.Background(), vmIDResponse.ID, metav1.GetOptions{})
	if err != nil {
		t.Error(err)
	}

	checkVirtualMachineResult(t, vm, vmRequest)
}

func TestPostDisk(t *testing.T) {

	ksClient := fakeks.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset()
	handler := newHandler(ksClient, k8sClient)

	namespace := "default"
	url := fmt.Sprintf("/namespaces/%s/disk", namespace)

	diskRequest := ui_virtz.DiskRequest{
		Name:        "testdisk",
		Description: "testdisk-description",
		Size:        20,
	}

	diskRequestBodyBytes, err := json.Marshal(diskRequest)
	if err != nil {
		t.Fatalf("Error marshaling JSON: %v", err)
	}

	request := httptest.NewRequest("POST", url, bytes.NewReader(diskRequestBodyBytes))
	request.Header.Set("Content-Type", "application/json")
	restfulRequest := restful.NewRequest(request)

	pathMap := make(map[string]string)
	pathMap["namespace"] = namespace
	if err := reflectutils.SetUnExportedField(restfulRequest, "pathParameters", pathMap); err != nil {
		t.Fatalf("set pathParameters failed")
	}

	recorder := httptest.NewRecorder()
	restfulResponse := restful.NewResponse(recorder)
	restfulResponse.SetRequestAccepts("application/json")

	handler.CreateDisk(restfulRequest, restfulResponse)
	if status := restfulResponse.StatusCode(); status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	res := restfulResponse.ResponseWriter.(*httptest.ResponseRecorder)

	var diskIDResponse ui_virtz.DiskIDResponse
	err = json.Unmarshal(res.Body.Bytes(), &diskIDResponse)
	if err != nil {
		t.Error(err)
	}

	// get disk from fake ks client
	disk, err := ksClient.VirtualizationV1alpha1().DiskVolumes(namespace).Get(context.Background(), diskIDResponse.ID, metav1.GetOptions{})
	if err != nil {
		t.Error(err)
	}

	if disk.Annotations[virtzv1alpha1.VirtualizationAliasName] != diskRequest.Name {
		t.Errorf("disk alias name is not correct: got %v want %v", disk.Annotations[virtzv1alpha1.VirtualizationAliasName], diskRequest.Name)
	}

	if disk.Annotations[virtzv1alpha1.VirtualizationDescription] != diskRequest.Description {
		t.Errorf("disk description is not correct: got %v want %v", disk.Annotations[virtzv1alpha1.VirtualizationDescription], diskRequest.Description)
	}

	size := strconv.FormatUint(uint64(diskRequest.Size), 10) + "Gi"
	if disk.Spec.Resources.Requests.Storage().String() != size {
		t.Errorf("disk size is not correct: got %v want %v", disk.Spec.Resources.Requests.Storage().String(), size)
	}

	if disk.Labels[virtzv1alpha1.VirtualizationDiskType] != "data" {
		t.Errorf("disk type is not correct: got %v want %v", disk.Labels[virtzv1alpha1.VirtualizationDiskType], "data")
	}

}

func TestGetDisk(t *testing.T) {

	ksClient := fakeks.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset()
	handler := newHandler(ksClient, k8sClient)

	namespace := "default"
	diskName := "testdisk"

	disk := ui_virtz.DiskRequest{
		Name:        diskName,
		Description: "testdisk",
		Size:        20,
	}

	diskVolume, err := createDisk(&handler, &disk, namespace)
	if err != nil {
		t.Error(err)
	}

	diskID := diskVolume.Name
	url := fmt.Sprintf("/namespaces/%s/disk/%s", namespace, diskID)

	request := httptest.NewRequest("GET", url, nil)
	restfulRequest := restful.NewRequest(request)

	pathMap := make(map[string]string)
	pathMap["namespace"] = namespace
	pathMap["id"] = diskID
	if err := reflectutils.SetUnExportedField(restfulRequest, "pathParameters", pathMap); err != nil {
		t.Fatalf("set pathParameters failed")
	}

	recorder := httptest.NewRecorder()
	restfulResponse := restful.NewResponse(recorder)
	restfulResponse.SetRequestAccepts("application/json")

	handler.GetDisk(restfulRequest, restfulResponse)
	if status := restfulResponse.StatusCode(); status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	res := restfulResponse.ResponseWriter.(*httptest.ResponseRecorder)

	var diskResponse ui_virtz.DiskResponse
	err = json.Unmarshal(res.Body.Bytes(), &diskResponse)
	if err != nil {
		t.Error(err)
	}

	if diskResponse.ID != diskID {
		t.Errorf("disk id is not correct: got %v want %v", diskResponse.ID, diskID)
	}

}

func TestPostImage(t *testing.T) {

	ksClient := fakeks.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset()
	handler := newHandler(ksClient, k8sClient)

	prepareFakeMinioService(k8sClient)

	namespace := "default"
	url := fmt.Sprintf("/namespaces/%s/image", namespace)

	imageRequest := ui_virtz.ImageRequest{
		Name:           "testimage",
		Description:    "testimage",
		Size:           20,
		OSFamily:       "ubuntu",
		Version:        "20.04_LTS_64bit",
		MinioImageName: "testimage",
		Shared:         false,
	}

	imageRequestBodyBytes, err := json.Marshal(imageRequest)
	if err != nil {
		t.Fatalf("Error marshaling JSON: %v", err)
	}

	request := httptest.NewRequest("POST", url, bytes.NewReader(imageRequestBodyBytes))
	request.Header.Set("Content-Type", "application/json")
	restfulRequest := restful.NewRequest(request)

	pathMap := make(map[string]string)
	pathMap["namespace"] = namespace
	if err := reflectutils.SetUnExportedField(restfulRequest, "pathParameters", pathMap); err != nil {
		t.Fatalf("set pathParameters failed")
	}

	recorder := httptest.NewRecorder()
	restfulResponse := restful.NewResponse(recorder)
	restfulResponse.SetRequestAccepts("application/json")

	handler.CreateImage(restfulRequest, restfulResponse)
	if status := restfulResponse.StatusCode(); status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	res := restfulResponse.ResponseWriter.(*httptest.ResponseRecorder)

	var imageIDResponse ui_virtz.ImageIDResponse
	err = json.Unmarshal(res.Body.Bytes(), &imageIDResponse)
	if err != nil {
		t.Error(err)
	}

	// get image from fake ks client
	image, err := ksClient.VirtualizationV1alpha1().ImageTemplates(namespace).Get(context.Background(), imageIDResponse.ID, metav1.GetOptions{})
	if err != nil {
		t.Error(err)
	}

	if image.Annotations[virtzv1alpha1.VirtualizationAliasName] != imageRequest.Name {
		t.Errorf("image alias name is not correct: got %v want %v", image.Annotations[virtzv1alpha1.VirtualizationAliasName], imageRequest.Name)
	}

	if image.Annotations[virtzv1alpha1.VirtualizationDescription] != imageRequest.Description {
		t.Errorf("image description is not correct: got %v want %v", image.Annotations[virtzv1alpha1.VirtualizationDescription], imageRequest.Description)
	}

	if image.Labels[virtzv1alpha1.VirtualizationOSFamily] != imageRequest.OSFamily {
		t.Errorf("image os family is not correct: got %v want %v", image.Labels[virtzv1alpha1.VirtualizationOSFamily], imageRequest.OSFamily)
	}

	if image.Labels[virtzv1alpha1.VirtualizationOSVersion] != imageRequest.Version {
		t.Errorf("image version is not correct: got %v want %v", image.Labels[virtzv1alpha1.VirtualizationOSVersion], imageRequest.Version)
	}

	size := strconv.FormatUint(uint64(imageRequest.Size), 10) + "Gi"
	if image.Spec.Resources.Requests.Storage().String() != size {
		t.Errorf("image size is not correct: got %v want %v", image.Spec.Resources.Requests.Storage().String(), size)
	}

	expectedURL := fmt.Sprintf("http://1.2.3.4:9000/ecpaas-images/%s", imageRequest.MinioImageName)
	if image.Spec.Source.HTTP.URL != expectedURL {
		t.Errorf("image url is not correct: got %v want %v", image.Spec.Source.HTTP.URL, expectedURL)
	}

	if image.Spec.Attributes.Public != imageRequest.Shared {
		t.Errorf("image shared is not correct: got %v want %v", image.Spec.Attributes.Public, imageRequest.Shared)
	}

}

func TestGetImage(t *testing.T) {

	ksClient := fakeks.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset()
	handler := newHandler(ksClient, k8sClient)

	prepareFakeMinioService(k8sClient)

	namespace := "default"
	imageName := "testimage"

	image := ui_virtz.ImageRequest{
		Name:           imageName,
		Description:    "testimage",
		Size:           20,
		OSFamily:       "ubuntu",
		Version:        "20.04_LTS_64bit",
		MinioImageName: "testimage",
		Shared:         false,
	}

	imageTempate, err := createImage(&handler, &image, namespace)
	if err != nil {
		t.Error(err)
	}

	imageID := imageTempate.Name
	url := fmt.Sprintf("/namespaces/%s/image/%s", namespace, imageID)

	request := httptest.NewRequest("GET", url, nil)
	restfulRequest := restful.NewRequest(request)

	pathMap := make(map[string]string)
	pathMap["namespace"] = namespace
	pathMap["id"] = imageID
	if err := reflectutils.SetUnExportedField(restfulRequest, "pathParameters", pathMap); err != nil {
		t.Fatalf("set pathParameters failed")
	}

	recorder := httptest.NewRecorder()
	restfulResponse := restful.NewResponse(recorder)
	restfulResponse.SetRequestAccepts("application/json")

	handler.GetImage(restfulRequest, restfulResponse)
	if status := restfulResponse.StatusCode(); status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	res := restfulResponse.ResponseWriter.(*httptest.ResponseRecorder)

	var imageResponse ui_virtz.ImageResponse
	err = json.Unmarshal(res.Body.Bytes(), &imageResponse)
	if err != nil {
		t.Error(err)
	}

	if imageResponse.ID != imageID {
		t.Errorf("image id is not correct: got %v want %v", imageResponse.ID, imageID)
	}

	if imageResponse.Name != image.Name {
		t.Errorf("image name is not correct: got %v want %v", imageResponse.Name, image.Name)
	}

	if imageResponse.Description != image.Description {
		t.Errorf("image description is not correct: got %v want %v", imageResponse.Description, image.Description)
	}

	if imageResponse.OSFamily != image.OSFamily {
		t.Errorf("image os family is not correct: got %v want %v", imageResponse.OSFamily, image.OSFamily)
	}

	if imageResponse.Version != image.Version {
		t.Errorf("image version is not correct: got %v want %v", imageResponse.Version, image.Version)
	}

	if imageResponse.Size != image.Size {
		t.Errorf("image size is not correct: got %v want %v", imageResponse.Size, image.Size)
	}

	if imageResponse.MinioImageName != image.MinioImageName {
		t.Errorf("image minio image name is not correct: got %v want %v", imageResponse.MinioImageName, image.MinioImageName)
	}

	if imageResponse.Shared != image.Shared {
		t.Errorf("image shared is not correct: got %v want %v", imageResponse.Shared, image.Shared)
	}

}
