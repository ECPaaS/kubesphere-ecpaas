/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com
*/

package virtualization

import (
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful"

	ui_virtz "kubesphere.io/kubesphere/pkg/models/virtualization"
)

func isValidWithinRange(validateType reflect.Type, valueToValidate int, fieldName string, resp *restful.Response) bool {
	field, found := validateType.FieldByName(fieldName)
	if found {
		minimum, _ := strconv.Atoi(field.Tag.Get("minimum"))
		maximum, _ := strconv.Atoi(field.Tag.Get("maximum"))
		if valueToValidate > maximum || valueToValidate < minimum {
			resp.WriteHeaderAndEntity(http.StatusForbidden, BadRequestError{
				Reason: fieldName + " should be in the range of " + field.Tag.Get("minimum") + " to " + field.Tag.Get("maximum"),
			})
			return false
		}
	}
	return true

}

func isValidLength(validateType reflect.Type, valueToValidate string, fieldName string, resp *restful.Response) bool {
	field, found := validateType.FieldByName(fieldName)
	if found {
		maximum, _ := strconv.Atoi(field.Tag.Get("maximum"))
		if len(valueToValidate) > int(maximum) {
			resp.WriteHeaderAndEntity(http.StatusForbidden, BadRequestError{
				Reason: fieldName + " length should be less than " + field.Tag.Get("maximum"),
			})
			return false
		}
	}
	return true
}

func isValidString(valueToValidate string, resp *restful.Response) bool {
	validRegex := regexp.MustCompile("^[A-Za-z0-9-]+$")
	if !validRegex.MatchString(valueToValidate) {
		resp.WriteHeaderAndEntity(http.StatusForbidden, BadRequestError{
			Reason: "Valid characters: A-Z, a-z, 0-9, and -(hyphen)",
		})
		return false
	}
	return true
}

func isValidVirtualMachine(vm ui_virtz.VirtualMachineRequest, resp *restful.Response) bool {

	reflectType := reflect.TypeOf(vm)
	if !isValidLength(reflectType, vm.Name, "Name", resp) {
		return false
	}

	if !isValidString(vm.Name, resp) {
		return false
	}

	if !isValidLength(reflectType, vm.Description, "Description", resp) {
		return false
	}

	if !isValidWithinRange(reflectType, int(vm.CpuCores), "CpuCores", resp) {
		return false
	}

	if !isValidWithinRange(reflectType, int(vm.Memory), "Memory", resp) {
		return false
	}

	return true
}

func isValidModifyVirtualMachine(vm ui_virtz.ModifyVirtualMachineRequest, resp *restful.Response) bool {

	reflectType := reflect.TypeOf(vm)
	if !isValidLength(reflectType, vm.Name, "Name", resp) {
		return false
	}

	if vm.Name != "" {
		if !isValidString(vm.Name, resp) {
			return false
		}
	}

	if vm.Description != "" {
		if !isValidLength(reflectType, vm.Description, "Description", resp) {
			return false
		}
	}

	if vm.CpuCores != 0 {
		if !isValidWithinRange(reflectType, int(vm.CpuCores), "CpuCores", resp) {
			return false
		}
	}

	if vm.Memory != 0 {
		if !isValidWithinRange(reflectType, int(vm.Memory), "Memory", resp) {
			return false
		}
	}

	return true
}

func isValidDiskRequest(disk ui_virtz.DiskRequest, resp *restful.Response) bool {

	reflectType := reflect.TypeOf(disk)
	if !isValidLength(reflectType, disk.Name, "Name", resp) {
		return false
	}

	if !isValidString(disk.Name, resp) {
		return false
	}

	if !isValidLength(reflectType, disk.Description, "Description", resp) {
		return false
	}

	if !isValidWithinRange(reflectType, int(disk.Size), "Size", resp) {
		return false
	}

	return true
}

func isValidModifyDiskRequest(disk ui_virtz.ModifyDiskRequest, resp *restful.Response) bool {

	reflectType := reflect.TypeOf(disk)
	if !isValidLength(reflectType, disk.Name, "Name", resp) {
		return false
	}

	if disk.Name != "" {
		if !isValidString(disk.Name, resp) {
			return false
		}
	}

	if disk.Description != "" {
		if !isValidLength(reflectType, disk.Description, "Description", resp) {
			return false
		}
	}

	if disk.Size != 0 {
		if !isValidWithinRange(reflectType, int(disk.Size), "Size", resp) {
			return false
		}
	}

	return true
}

func isValidImageRequest(image ui_virtz.ImageRequest, resp *restful.Response) bool {

	reflectType := reflect.TypeOf(image)
	if !isValidLength(reflectType, image.Name, "Name", resp) {
		return false
	}

	if !isValidString(image.Name, resp) {
		return false
	}

	if !isValidLength(reflectType, image.Description, "Description", resp) {
		return false
	}

	if !isValidWithinRange(reflectType, int(image.CpuCores), "CpuCores", resp) {
		return false
	}

	if !isValidWithinRange(reflectType, int(image.Memory), "Memory", resp) {
		return false
	}

	if !isValidWithinRange(reflectType, int(image.Size), "Size", resp) {
		return false
	}

	return true
}

func isValidModifyImageRequest(image ui_virtz.ModifyImageRequest, resp *restful.Response) bool {

	reflectType := reflect.TypeOf(image)
	if !isValidLength(reflectType, image.Name, "Name", resp) {
		return false
	}

	if image.Name != "" {
		if !isValidString(image.Name, resp) {
			return false
		}
	}

	if image.Description != "" {
		if !isValidLength(reflectType, image.Description, "Description", resp) {
			return false
		}
	}

	if image.CpuCores != 0 {
		if !isValidWithinRange(reflectType, int(image.CpuCores), "CpuCores", resp) {
			return false
		}
	}

	if image.Memory != 0 {
		if !isValidWithinRange(reflectType, int(image.Memory), "Memory", resp) {
			return false
		}
	}

	if image.Size != 0 {
		if !isValidWithinRange(reflectType, int(image.Size), "Size", resp) {
			return false
		}
	}

	return true
}

func isValidDiskSize(h *virtzhandler, namespace string, diskName string, newDiskSize int, resp *restful.Response) bool {
	diskVolume, err := h.virtz.GetDisk(namespace, diskName)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return false
	}

	oldDiskSize, _ := strconv.ParseUint(strings.Replace(diskVolume.Spec.Resources.Requests.Storage().String(), "Gi", "", -1), 10, 32)
	if int(oldDiskSize) >= newDiskSize {
		resp.WriteHeaderAndEntity(http.StatusForbidden, BadRequestError{
			Reason: "The new disk size must be larger than the old disk size",
		})
		return false
	}
	return true
}
