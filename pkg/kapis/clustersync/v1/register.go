/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com
*/

package v1

import (
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/kapis/util"
	ui_clustersync "kubesphere.io/kubesphere/pkg/models/clustersync"
)

const (
	GroupName = "clustersync.ecpaas.io"
)

var putNotes = `Any parameters which are not provided will not be changed.`

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

func AddToContainer(container *restful.Container, ksclient kubesphere.Interface, k8sclient kubernetes.Interface) error {
	webservice := runtime.NewWebService(GroupVersion)
	handler := newHandler(ksclient, k8sclient)

	// StorageConfigs
	webservice.Route(webservice.POST("/storageconfigs").
		To(handler.CreateStorage).
		Reads(ui_clustersync.StorageRequest{}).
		Doc("Create storage").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.StorageNameResponse{}).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncStorageTag}))

	webservice.Route(webservice.PUT("/storageconfigs/{name}").
		To(handler.UpdateStorage).
		Param(webservice.PathParameter("name", "storage name")).
		Reads(ui_clustersync.ModifyStorageRequest{}).
		Doc("Update Storage").
		Notes(putNotes).
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncStorageTag}))

	webservice.Route(webservice.GET("/storageconfigs/{name}").
		To(handler.GetStorage).
		Param(webservice.PathParameter("name", "storage name")).
		Doc("Get storage").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.StorageResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncStorageTag}))

	webservice.Route(webservice.GET("/storageconfigs").
		To(handler.ListStorage).
		Doc("List all storages").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.ListStorageResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncStorageTag}))

	webservice.Route(webservice.DELETE("/storageconfigs/{name}").
		To(handler.DeleteStorage).
		Param(webservice.PathParameter("name", "storage name")).
		Doc("Delete storage").
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncStorageTag}))

	// BackupConfigs
	webservice.Route(webservice.POST("/backupconfigs").
		To(handler.CreateBackup).
		Reads(ui_clustersync.BackupRequest{}).
		Doc("Create backup").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.BackupNameResponse{}).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncBackupTag}))

	webservice.Route(webservice.PUT("/backupconfigs/{name}").
		To(handler.UpdateBackup).
		Param(webservice.PathParameter("name", "backup name")).
		Reads(ui_clustersync.ModifyBackupRequest{}).
		Doc("Update backup").
		Notes(putNotes).
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncBackupTag}))

	webservice.Route(webservice.GET("/backupconfigs/{name}").
		To(handler.GetBackup).
		Param(webservice.PathParameter("name", "backup name")).
		Doc("Get backup").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.BackupResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncBackupTag}))

	webservice.Route(webservice.GET("/backupconfigs").
		To(handler.ListBackup).
		Doc("List all backups").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.ListBackupResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncBackupTag}))

	webservice.Route(webservice.DELETE("/backupconfigs/{name}").
		To(handler.DeleteBackup).
		Param(webservice.PathParameter("name", "backup name")).
		Doc("Delete backup").
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncBackupTag}))

	// RestoreConfigs
	webservice.Route(webservice.POST("/restoreconfigs").
		To(handler.CreateRestore).
		Reads(ui_clustersync.RestoreRequest{}).
		Doc("Create restore").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.RestoreNameResponse{}).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncRestoreTag}))

	webservice.Route(webservice.PUT("/restoreconfigs/{name}").
		To(handler.UpdateRestore).
		Param(webservice.PathParameter("name", "restore name")).
		Reads(ui_clustersync.ModifyRestoreRequest{}).
		Doc("Update restore").
		Notes(putNotes).
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncRestoreTag}))

	webservice.Route(webservice.GET("/restoreconfigs/{name}").
		To(handler.GetRestore).
		Param(webservice.PathParameter("name", "restore name")).
		Doc("Get restore").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.RestoreResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncRestoreTag}))

	webservice.Route(webservice.GET("/restoreconfigs").
		To(handler.ListRestore).
		Doc("List all restores").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.ListRestoreResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncRestoreTag}))

	webservice.Route(webservice.DELETE("/restoreconfigs/{name}").
		To(handler.DeleteRestore).
		Param(webservice.PathParameter("name", "restore name")).
		Doc("Delete restore").
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncRestoreTag}))

	// ScheduleConfigs
	webservice.Route(webservice.POST("/scheduleconfigs").
		To(handler.CreateSchedule).
		Reads(ui_clustersync.ScheduleRequest{}).
		Doc("Create schedule").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.ScheduleNameResponse{}).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncScheduleTag}))

	webservice.Route(webservice.PUT("/scheduleconfigs/{name}").
		To(handler.UpdateSchedule).
		Param(webservice.PathParameter("name", "schedule name")).
		Reads(ui_clustersync.ModifyScheduleRequest{}).
		Doc("Update schedule").
		Notes(putNotes).
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncScheduleTag}))

	webservice.Route(webservice.GET("/scheduleconfigs/{name}").
		To(handler.GetSchedule).
		Param(webservice.PathParameter("name", "schedule name")).
		Doc("Get schedule").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.ScheduleResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncScheduleTag}))

	webservice.Route(webservice.GET("/scheduleconfigs").
		To(handler.ListSchedule).
		Doc("List all schedules").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.ListScheduleResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncScheduleTag}))

	webservice.Route(webservice.DELETE("/scheduleconfigs/{name}").
		To(handler.DeleteSchedule).
		Param(webservice.PathParameter("name", "schedule name")).
		Doc("Delete schedule").
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncScheduleTag}))

	container.Add(webservice)

	return nil
}
