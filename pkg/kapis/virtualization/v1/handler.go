package virtualization

import (
	"encoding/json"
	"net/http"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"

	virtzv1alpha1 "kubesphere.io/api/virtualization/v1alpha1"
	ui_virtz "kubesphere.io/kubesphere/pkg/models/virtualization"
)

type virtzhandler struct {
	virtz ui_virtz.Interface
}

func newHandler(ksclient kubesphere.Interface, k8sclient kubernetes.Interface) virtzhandler {
	return virtzhandler{
		virtz: ui_virtz.New(ksclient, k8sclient),
	}
}

func (h *virtzhandler) CreateVirtualMahcine(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")

	var ui_vm ui_virtz.VirtualMachineRequest
	err := req.ReadEntity(&ui_vm)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	_, err = h.virtz.CreateVirtualMachine(namespace, &ui_vm)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteHeader(http.StatusOK)
}

func (h *virtzhandler) UpdateVirtualMahcine(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	vmName := req.PathParameter("id")

	var ui_vm ui_virtz.VirtualMachineRequest
	err := req.ReadEntity(&ui_vm)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	_, err = h.virtz.UpdateVirtualMachine(namespace, vmName, &ui_vm)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteHeader(http.StatusOK)
}

func (h *virtzhandler) GetVirtualMachine(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	vmName := req.PathParameter("id")

	vm, err := h.virtz.GetVirtualMachine(namespace, vmName)
	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	ui_virtz_vm_resp := h.getUIVirtualMachineResponse(vm)

	resp.WriteEntity(ui_virtz_vm_resp)
}

func (h *virtzhandler) getUIVirtualMachineResponse(vm *virtzv1alpha1.VirtualMachine) ui_virtz.VirtualMachineResponse {

	ui_vm_status := ui_virtz.VMStatus{}
	ui_vm_status.Ready = vm.Status.Ready
	ui_vm_status.State = string(vm.Status.PrintableStatus)

	ui_image_spec := h.getUIImageInfoResponse(vm)

	return ui_virtz.VirtualMachineResponse{
		Name:        vm.Annotations[virtzv1alpha1.VirtualizationAliasName],
		ID:          vm.Name,
		Namespace:   vm.Namespace,
		Description: vm.Annotations[virtzv1alpha1.VirtualizationDescription],
		CpuCores:    vm.Spec.Hardware.Domain.CPU.Cores,
		Memory:      vm.Spec.Hardware.Domain.Resources.Requests.Memory().String(),
		Image:       &ui_image_spec,
		Disks:       h.getUIDisksResponse(vm),
		Status:      ui_vm_status,
	}
}

func (h *virtzhandler) getUIImageInfoResponse(vm *virtzv1alpha1.VirtualMachine) ui_virtz.ImageInfoResponse {
	jsonImageInfo := vm.Annotations[virtzv1alpha1.VirtualizationImageInfo]

	var uiImageInfo ui_virtz.ImageInfo

	err := json.Unmarshal([]byte(jsonImageInfo), &uiImageInfo)
	if err != nil {
		klog.Error(err)
		return ui_virtz.ImageInfoResponse{}
	}

	return ui_virtz.ImageInfoResponse{
		Name: uiImageInfo.Name,
		Size: vm.Annotations[virtzv1alpha1.VirtualizationSystemDiskSize],
	}
}

func (h *virtzhandler) getUIDisksResponse(vm *virtzv1alpha1.VirtualMachine) []ui_virtz.DiskResponse {
	diskvolumeList, err := h.virtz.ListDisk("")
	if err != nil {
		klog.Error(err)
		return nil
	}

	diskvolumes := make(map[string]virtzv1alpha1.DiskVolume)
	for _, diskvolume := range diskvolumeList.Items {
		for _, vm_diskvolme := range vm.Spec.DiskVolumes {
			if diskvolume.Name == vm_diskvolme {
				diskvolumes[diskvolume.Name] = diskvolume
			}
		}
	}

	ui_virtz_disk_resp := make([]ui_virtz.DiskResponse, 0)
	for _, diskvolume := range diskvolumes {
		ui_virtz_disk_resp = append(ui_virtz_disk_resp, getUIDiskResponse(&diskvolume))
	}

	return ui_virtz_disk_resp
}

func getUIDiskResponse(diskvolume *virtzv1alpha1.DiskVolume) ui_virtz.DiskResponse {

	ui_disk_status := ui_virtz.DiskStatus{}
	ui_disk_status.Ready = diskvolume.Status.Ready
	ui_disk_status.Owner = diskvolume.Status.Owner

	return ui_virtz.DiskResponse{
		Name:        diskvolume.Annotations[virtzv1alpha1.VirtualizationAliasName],
		ID:          diskvolume.Name,
		Namespace:   diskvolume.Namespace,
		Description: diskvolume.Annotations[virtzv1alpha1.VirtualizationDescription],
		Type:        diskvolume.Labels[virtzv1alpha1.VirtualizationDiskType],
		Size:        diskvolume.Spec.Resources.Requests.Storage().String(),
		Status:      ui_disk_status,
	}
}

func (h *virtzhandler) ListVirtualMachine(req *restful.Request, resp *restful.Response) {
	ui_list_vm_resp, err := h.listVirtualMachine("", resp)
	if err != nil {
		return
	}

	resp.WriteEntity(ui_list_vm_resp)
}

func (h *virtzhandler) ListVirtualMachineWithNamespace(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")

	ui_list_vm_resp, err := h.listVirtualMachine(namespace, resp)
	if err != nil {
		return
	}

	resp.WriteEntity(ui_list_vm_resp)
}

func (h *virtzhandler) listVirtualMachine(namespace string, resp *restful.Response) (*ui_virtz.ListVirtualMachineResponse, error) {
	vms, err := h.virtz.ListVirtualMachine(namespace)
	if err != nil {
		klog.Error(err)
		resp.WriteError(http.StatusInternalServerError, err)
		return nil, err
	}

	ui_virtz_vm_resp := make([]ui_virtz.VirtualMachineResponse, 0)
	for _, vm := range vms.Items {
		vm_resp := h.getUIVirtualMachineResponse(&vm)
		ui_virtz_vm_resp = append(ui_virtz_vm_resp, vm_resp)
	}

	ui_list_virtz_vm_resp := ui_virtz.ListVirtualMachineResponse{
		TotalCount: len(ui_virtz_vm_resp),
		Items:      ui_virtz_vm_resp,
	}

	return &ui_list_virtz_vm_resp, nil
}

func (h *virtzhandler) DeleteVirtualMachine(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	vmName := req.PathParameter("id")

	_, err := h.virtz.DeleteVirtualMachine(namespace, vmName)
	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteHeader(http.StatusOK)
}

func (h *virtzhandler) CreateDisk(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")

	var ui_disk ui_virtz.DiskRequest
	err := req.ReadEntity(&ui_disk)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	_, err = h.virtz.CreateDisk(namespace, &ui_disk)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteHeader(http.StatusOK)
}

func (h *virtzhandler) UpdateDisk(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	diskName := req.PathParameter("id")

	var ui_disk ui_virtz.DiskRequest
	err := req.ReadEntity(&ui_disk)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	_, err = h.virtz.UpdateDisk(namespace, diskName, &ui_disk)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteHeader(http.StatusOK)
}

func (h *virtzhandler) GetDisk(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	diskName := req.PathParameter("id")

	disk, err := h.virtz.GetDisk(namespace, diskName)
	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(getUIDiskResponse(disk))
}

func (h *virtzhandler) ListDiskWithNamespace(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")

	ui_list_disk_resp, err := h.listDisk(namespace, resp)
	if err != nil {
		return
	}

	resp.WriteEntity(ui_list_disk_resp)
}

func (h *virtzhandler) ListDisk(req *restful.Request, resp *restful.Response) {
	ui_list_disk_resp, err := h.listDisk("", resp)
	if err != nil {
		return
	}

	resp.WriteEntity(ui_list_disk_resp)
}

func (h *virtzhandler) listDisk(namespace string, resp *restful.Response) (*ui_virtz.ListDiskResponse, error) {
	disks, err := h.virtz.ListDisk(namespace)
	if err != nil {
		klog.Error(err)
		resp.WriteError(http.StatusInternalServerError, err)
		return nil, err
	}

	ui_disk_resp := make([]ui_virtz.DiskResponse, 0)
	for _, disk := range disks.Items {
		ui_disk_resp = append(ui_disk_resp, getUIDiskResponse(&disk))
	}

	ui_list_disk_resp := ui_virtz.ListDiskResponse{
		TotalCount: len(ui_disk_resp),
		Items:      ui_disk_resp,
	}

	return &ui_list_disk_resp, nil
}

func (h *virtzhandler) DeleteDisk(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	diskName := req.PathParameter("id")

	_, err := h.virtz.DeleteDisk(namespace, diskName)
	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteHeader(http.StatusOK)
}

func (h *virtzhandler) CreateImage(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")

	var ui_image ui_virtz.ImageRequest
	err := req.ReadEntity(&ui_image)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	_, err = h.virtz.CreateImage(namespace, &ui_image)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteHeader(http.StatusOK)
}

func (h *virtzhandler) GetImage(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	imageName := req.PathParameter("id")

	image, err := h.virtz.GetImage(namespace, imageName)
	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(getUIImageResponse(image))
}

func getUIImageResponse(image *virtzv1alpha1.ImageTemplate) ui_virtz.ImageResponse {
	return ui_virtz.ImageResponse{
		Name:        image.Name,
		Namespace:   image.Namespace,
		OSFamily:    image.Annotations[virtzv1alpha1.VirtualizationOSFamily],
		Version:     image.Annotations[virtzv1alpha1.VirtualizationOSVersion],
		CpuCores:    image.Annotations[virtzv1alpha1.VirtualizationCpuCores],
		Memory:      image.Annotations[virtzv1alpha1.VirtualizationImageMemory],
		Size:        image.Annotations[virtzv1alpha1.VirtualizationSystemDiskSize],
		Description: image.Annotations[virtzv1alpha1.VirtualizationDescription],
	}
}
