/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com
*/

package v1

import (
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"

	ui_clustersync "kubesphere.io/kubesphere/pkg/models/clustersync"
)

type clustersyncHandler struct {
	clustersync         ui_clustersync.Interface
	k8sClient           kubernetes.Interface
}

func newHandler(ksclient kubesphere.Interface, k8sclient kubernetes.Interface) clustersyncHandler {
	return clustersyncHandler{
		clustersync:  ui_clustersync.New(ksclient, k8sclient),
		k8sClient:    k8sclient,
	}
}


// Storage

// Create new storageConfig in OperatorConfig.spec.storageConfigs
func (h *clustersyncHandler) CreateStorage(req *restful.Request, resp *restful.Response) {
	var ui_storage ui_clustersync.StorageRequest
	err := req.ReadEntity(&ui_storage)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	if !isValidStorageRequest(&ui_storage , resp) {
		return
	}

	// Create storageConfig into OperatorConfig
	ui_storageName, err := h.clustersync.CreateStorage(&ui_storage)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(ui_storageName)
}

// Update existed storageConfig in OperatorConfig.spec.storageConfigs
func (h *clustersyncHandler) UpdateStorage(req *restful.Request, resp *restful.Response) {
	var ui_storage ui_clustersync.ModifyStorageRequest
	err := req.ReadEntity(&ui_storage)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Validation of ModifyStorageRequest
	if !isValidStorageModifyRequest(&ui_storage , resp) {
		return
	}

	// Update storageConfig in OperatorConfig
	storageConfigName := req.PathParameter("name")
	_, err = h.clustersync.UpdateStorage(storageConfigName, &ui_storage)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Make response
	resp.WriteEntity(http.StatusOK)
}

// Get existed storageConfig in OperatorConfig.spec.storageConfigs
func (h *clustersyncHandler) GetStorage(req *restful.Request, resp *restful.Response) {
	storageConfigName := req.PathParameter("name")

	storageResponse, err := h.clustersync.GetStorage(storageConfigName)
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(storageResponse)
}

// List all storageConfigs in OperatorConfig.spec.storageConfigs
func (h *clustersyncHandler) ListStorage(req *restful.Request, resp *restful.Response) {
	listStorageResponse, err := h.clustersync.ListStorage()
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(listStorageResponse)
}

// Delete existed storageConfig in OperatorConfig.spec.storageConfigs
func (h *clustersyncHandler) DeleteStorage(req *restful.Request, resp *restful.Response) {
	storageConfigName := req.PathParameter("name")

	err := h.clustersync.DeleteStorage(storageConfigName)
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(http.StatusOK)
}


// Backup

// Create new backupConfig in OperatorConfig.spec.backupConfigs
func (h *clustersyncHandler) CreateBackup(req *restful.Request, resp *restful.Response) {
	var ui_backup ui_clustersync.BackupRequest
	err := req.ReadEntity(&ui_backup)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	if !isValidBackupRequest(&ui_backup , resp) {
		return
	}

	// Create backupConfig into OperatorConfig
	ui_backupName, err := h.clustersync.CreateBackup(&ui_backup)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(ui_backupName)
}

// Update existed backupConfig in OperatorConfig.spec.backupConfigs
func (h *clustersyncHandler) UpdateBackup(req *restful.Request, resp *restful.Response) {
	var ui_backup ui_clustersync.ModifyBackupRequest
	err := req.ReadEntity(&ui_backup)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Validation of ModifyBackupRequest
	if !isValidBackupModifyRequest(&ui_backup , resp) {
		return
	}

	// Update backupConfig in OperatorConfig
	backupConfigName := req.PathParameter("name")
	_, err = h.clustersync.UpdateBackup(backupConfigName, &ui_backup)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Make response
	resp.WriteEntity(http.StatusOK)
}

// Get existed backupConfig in OperatorConfig.spec.backupConfigs
func (h *clustersyncHandler) GetBackup(req *restful.Request, resp *restful.Response) {
	backupConfigName := req.PathParameter("name")

	backupResponse, err := h.clustersync.GetBackup(backupConfigName)
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(backupResponse)
}

// List all backupConfigs in OperatorConfig.spec.backupConfigs
func (h *clustersyncHandler) ListBackup(req *restful.Request, resp *restful.Response) {
	listBackupResponse, err := h.clustersync.ListBackup()
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(listBackupResponse)
}

// Delete existed backupConfig in OperatorConfig.spec.backupConfigs
func (h *clustersyncHandler) DeleteBackup(req *restful.Request, resp *restful.Response) {
	backupConfigName := req.PathParameter("name")

	err := h.clustersync.DeleteBackup(backupConfigName)
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(http.StatusOK)
}


// Restore

// Create new restoreConfig in OperatorConfig.spec.restoreConfigs
func (h *clustersyncHandler) CreateRestore(req *restful.Request, resp *restful.Response) {
	var ui_restore ui_clustersync.RestoreRequest
	err := req.ReadEntity(&ui_restore)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	if !isValidRestoreRequest(&ui_restore , resp) {
		return
	}

	// Create restoreConfig into OperatorConfig
	ui_restoreName, err := h.clustersync.CreateRestore(&ui_restore)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(ui_restoreName)
}

// Update existed restoreConfig in OperatorConfig.spec.restoreConfigs
func (h *clustersyncHandler) UpdateRestore(req *restful.Request, resp *restful.Response) {
	var ui_restore ui_clustersync.ModifyRestoreRequest
	err := req.ReadEntity(&ui_restore)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Validation of ModifyRestoreRequest
	if !isValidRestoreModifyRequest(&ui_restore , resp) {
		return
	}

	// Update restoreConfig in OperatorConfig
	restoreConfigName := req.PathParameter("name")
	_, err = h.clustersync.UpdateRestore(restoreConfigName, &ui_restore)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Make response
	resp.WriteEntity(http.StatusOK)
}

// Get existed restoreConfig in OperatorConfig.spec.restoreConfigs
func (h *clustersyncHandler) GetRestore(req *restful.Request, resp *restful.Response) {
	restoreConfigName := req.PathParameter("name")

	restoreResponse, err := h.clustersync.GetRestore(restoreConfigName)
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(restoreResponse)
}

// List all restoreConfigs in OperatorConfig.spec.restoreConfigs
func (h *clustersyncHandler) ListRestore(req *restful.Request, resp *restful.Response) {
	listRestoreResponse, err := h.clustersync.ListRestore()
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(listRestoreResponse)
}

// Delete existed restoreConfig in OperatorConfig.spec.restoreConfigs
func (h *clustersyncHandler) DeleteRestore(req *restful.Request, resp *restful.Response) {
	restoreConfigName := req.PathParameter("name")

	err := h.clustersync.DeleteRestore(restoreConfigName)
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(http.StatusOK)
}


// Schedule

// Create new scheduleConfig in OperatorConfig.spec.scheduleConfigs
func (h *clustersyncHandler) CreateSchedule(req *restful.Request, resp *restful.Response) {
	var ui_schedule ui_clustersync.ScheduleRequest
	err := req.ReadEntity(&ui_schedule)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	if !isValidScheduleRequest(&ui_schedule , resp) {
		return
	}

	// Create scheduleConfig into OperatorConfig
	ui_scheduleName, err := h.clustersync.CreateSchedule(&ui_schedule)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(ui_scheduleName)
}

// Update existed scheduleConfig in OperatorConfig.spec.scheduleConfigs
func (h *clustersyncHandler) UpdateSchedule(req *restful.Request, resp *restful.Response) {
	var ui_schedule ui_clustersync.ModifyScheduleRequest
	err := req.ReadEntity(&ui_schedule)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Validation of ModifyScheduleRequest
	if !isValidScheduleModifyRequest(&ui_schedule , resp) {
		return
	}

	// Update scheduleConfig in OperatorConfig
	scheduleConfigName := req.PathParameter("name")
	_, err = h.clustersync.UpdateSchedule(scheduleConfigName, &ui_schedule)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Make response
	resp.WriteEntity(http.StatusOK)
}

// Get existed scheduleConfig in OperatorConfig.spec.scheduleConfigs
func (h *clustersyncHandler) GetSchedule(req *restful.Request, resp *restful.Response) {
	scheduleConfigName := req.PathParameter("name")

	scheduleResponse, err := h.clustersync.GetSchedule(scheduleConfigName)
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(scheduleResponse)
}

// List all scheduleConfigs in OperatorConfig.spec.scheduleConfigs
func (h *clustersyncHandler) ListSchedule(req *restful.Request, resp *restful.Response) {
	listScheduleResponse, err := h.clustersync.ListSchedule()
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(listScheduleResponse)
}

// Delete existed scheduleConfig in OperatorConfig.spec.scheduleConfigs
func (h *clustersyncHandler) DeleteSchedule(req *restful.Request, resp *restful.Response) {
	scheduleConfigName := req.PathParameter("name")

	err := h.clustersync.DeleteSchedule(scheduleConfigName)
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(http.StatusOK)
}
