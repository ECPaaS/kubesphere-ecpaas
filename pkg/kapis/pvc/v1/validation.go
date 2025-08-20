/*
Copyright(c) 2025-present Accton. All rights reserved. www.accton.com
*/

package v1

import (
	"net/http"
	"reflect"

	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/kapis/util"
	ui_pvc "kubesphere.io/kubesphere/pkg/models/pvc"
)

func isValidClonePVCRequest(request *ui_pvc.PVCCloneRequest, resp *restful.Response) bool {
	reflectType := reflect.TypeOf(*request)
	// NewName string
	if request.NewName == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "NewName must not be empty.",
		})
		return false
	} else if !util.IsValidKubernetesString(request.NewName, "NewName", resp) {
		return false
	}
	// Namespace string
	if request.Namespace == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "Namespace must not be empty.",
		})
		return false
	} else if !util.IsValidNamespaceString(request.Namespace, resp) {
		return false
	}
	// StorageClass string
	if request.StorageClass == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "StorageClass must not be empty.",
		})
		return false
	} else if !util.IsValidKubernetesString(request.StorageClass, "StorageClass", resp) {
		return false
	}
	// Size int
	if request.Size == 0 {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "Size must not be empty",
		})
		return false
	} else if !util.IsValidWithinRange(reflectType, request.Size, "Size", resp) {
		return false
	}
	// AccessModes []string
	if len(request.AccessModes) == 0 {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "AccessModes must not be empty.",
		})
		return false
	} else {
		for _, mode := range request.AccessModes {
			switch mode {
			case "ReadWriteOnce", "ReadOnlyMany", "ReadWriteMany":
				continue
			default:
				resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
					Reason: "Available mode for accessModes: ReadWriteOnce, ReadOnlyMany, ReadWriteMany",
				})
				return false
			}
		}
	}
	// NodeSelector []NodeSelector optional
	if len(request.NodeSelector) != 0 {
		validateType := reflect.TypeOf(request.NodeSelector).Elem()
		if !util.IsValidLabels(validateType, len(request.NodeSelector), ui_pvc.ConvertNodeSelectorToMap(request.NodeSelector), resp) {
			return false
		}
	}

	return true
}
