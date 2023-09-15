package virtualization

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"

	virtzv1alpha1 "kubesphere.io/api/virtualization/v1alpha1"
	ui_virtz "kubesphere.io/kubesphere/pkg/models/virtualization"
)

type virtzhandler struct {
	virtz ui_virtz.Interface
}

func newHandler(ksclient kubesphere.Interface) virtzhandler {
	return virtzhandler{
		virtz: ui_virtz.New(ksclient),
	}
}

func (h *virtzhandler) CreateVirtualMahcine(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")

	var virtz_vm ui_virtz.VirtualMachine
	err := req.ReadEntity(&virtz_vm)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	_, err = h.virtz.CreateVirtualMachine(namespace, &virtz_vm)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteHeader(http.StatusOK)
}

func (h *virtzhandler) UpdateVirtualMahcine(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	vmName := req.PathParameter("id")

	var virtz_vm ui_virtz.VirtualMachine
	err := req.ReadEntity(&virtz_vm)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	_, err = h.virtz.UpdateVirtualMachine(namespace, vmName, &virtz_vm)
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

	ui_virtz_vm_resp := h.getVirtualMachineResponse(vm)

	resp.WriteEntity(ui_virtz_vm_resp)
}

func (h *virtzhandler) getVirtualMachineResponse(vm *virtzv1alpha1.VirtualMachine) ui_virtz.VirtualMachineResponse {

	status := ui_virtz.VMStatus{}
	status.Ready = vm.Status.Ready
	status.State = string(vm.Status.PrintableStatus)

	return ui_virtz.VirtualMachineResponse{
		Name:        vm.Annotations[virtzv1alpha1.VirtualizationAliasName],
		ID:          vm.Name,
		Namespace:   vm.Namespace,
		Description: vm.Annotations[virtzv1alpha1.VirtualizationDescription],
		CpuCores:    vm.Spec.Hardware.Domain.Cpu.Cores,
		Memory:      vm.Spec.Hardware.Domain.Resources.Requests.Memory().String(),
		Disks:       h.getDisksResponse(vm),
		Status:      status,
	}
}

func (h *virtzhandler) getDisksResponse(vm *virtzv1alpha1.VirtualMachine) []ui_virtz.DiskSpec {
	diskvolumeList, err := h.virtz.ListDiskVolume("")
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

	ui_virtz_diskvolume_resp := make([]ui_virtz.DiskSpec, 0)
	for _, diskvolume := range diskvolumes {
		diskvolume_resp := getDiskVolumeResponse(&diskvolume)
		ui_virtz_diskvolume_resp = append(ui_virtz_diskvolume_resp, diskvolume_resp)
	}

	return ui_virtz_diskvolume_resp
}

func getDiskVolumeResponse(diskvolume *virtzv1alpha1.DiskVolume) ui_virtz.DiskSpec {
	return ui_virtz.DiskSpec{
		Name:      diskvolume.Annotations[virtzv1alpha1.VirtualizationAliasName],
		ID:        diskvolume.Name,
		Namespace: diskvolume.Namespace,
		Type:      diskvolume.Labels[virtzv1alpha1.VirtualizationDiskType],
		Size:      diskvolume.Spec.Resources.Requests.Storage().String(),
	}
}

func (h *virtzhandler) ListVirtualMachine(req *restful.Request, resp *restful.Response) {
	vms, err := h.virtz.ListVirtualMachine("")
	if err != nil {
		klog.Error(err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	ui_virtz_vm_resp := make([]ui_virtz.VirtualMachineResponse, 0)
	for _, vm := range vms.Items {
		vm_resp := h.getVirtualMachineResponse(&vm)
		ui_virtz_vm_resp = append(ui_virtz_vm_resp, vm_resp)
	}

	ui_list_virtz_vm_resp := ui_virtz.ListVirtualMachineResponse{
		TotalCount: len(ui_virtz_vm_resp),
		Items:      ui_virtz_vm_resp,
	}

	resp.WriteEntity(ui_list_virtz_vm_resp)
}

func (h *virtzhandler) ListVirtualMachineWithNamespace(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")

	vms, err := h.virtz.ListVirtualMachine(namespace)
	if err != nil {
		klog.Error(err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	ui_virtz_vm_resp := make([]ui_virtz.VirtualMachineResponse, 0)
	for _, vm := range vms.Items {
		vm_resp := h.getVirtualMachineResponse(&vm)
		ui_virtz_vm_resp = append(ui_virtz_vm_resp, vm_resp)
	}

	ui_list_virtz_vm_resp := ui_virtz.ListVirtualMachineResponse{
		TotalCount: len(ui_virtz_vm_resp),
		Items:      ui_virtz_vm_resp,
	}

	resp.WriteEntity(ui_list_virtz_vm_resp)
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
