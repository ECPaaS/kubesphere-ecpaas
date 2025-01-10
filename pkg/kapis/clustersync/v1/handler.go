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


// Repository

// Create new repositoryConfig in OperatorConfig.spec.repositoryConfigs
func (h *clustersyncHandler) CreateRepository(req *restful.Request, resp *restful.Response) {
	var ui_repository ui_clustersync.RepositoryRequest
	err := req.ReadEntity(&ui_repository)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	if !isValidRepositoryRequest(&ui_repository , resp) {
		return
	}

	// Create repositoryConfig into OperatorConfig
	ui_repositoryName, err := h.clustersync.CreateRepository(&ui_repository)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(ui_repositoryName)
}

// Update existed repositoryConfig in OperatorConfig.spec.repositoryConfigs
func (h *clustersyncHandler) UpdateRepository(req *restful.Request, resp *restful.Response) {
	var ui_repository ui_clustersync.ModifyRepositoryRequest
	err := req.ReadEntity(&ui_repository)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Validation of ModifyRepositoryRequest
	if !isValidRepositoryModifyRequest(&ui_repository , resp) {
		return
	}

	// Update repositoryConfig in OperatorConfig
	repositoryConfigName := req.PathParameter("name")
	_, err = h.clustersync.UpdateRepository(repositoryConfigName, &ui_repository)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Make response
	resp.WriteEntity(http.StatusOK)
}

// Get existed repositoryConfig in OperatorConfig.spec.repositoryConfigs
func (h *clustersyncHandler) GetRepository(req *restful.Request, resp *restful.Response) {
	repositoryConfigName := req.PathParameter("name")

	repositoryResponse, err := h.clustersync.GetRepository(repositoryConfigName)
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(repositoryResponse)
}

// List all repositoryConfigs in OperatorConfig.spec.repositoryConfigs
func (h *clustersyncHandler) ListRepository(req *restful.Request, resp *restful.Response) {
	listRepositoryResponse, err := h.clustersync.ListRepository()
	if err != nil {
		if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "is not found") {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(listRepositoryResponse)
}

// Delete existed repositoryConfig in OperatorConfig.spec.repositoryConfigs
func (h *clustersyncHandler) DeleteRepository(req *restful.Request, resp *restful.Response) {
	repositoryConfigName := req.PathParameter("name")

	err := h.clustersync.DeleteRepository(repositoryConfigName)
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
