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

	// Repository
	webservice.Route(webservice.POST("/repository").
		To(handler.CreateRepository).
		Reads(ui_clustersync.RepositoryRequest{}).
		Doc("Create repository").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.RepositoryNameResponse{}).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncRepositoryTag}))

	webservice.Route(webservice.PUT("/repository/{name}").
		To(handler.UpdateRepository).
		Param(webservice.PathParameter("name", "repository name")).
		Reads(ui_clustersync.ModifyRepositoryRequest{}).
		Doc("Update Repository").
		Notes(putNotes).
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncRepositoryTag}))

	webservice.Route(webservice.GET("/repository/{name}").
		To(handler.GetRepository).
		Param(webservice.PathParameter("name", "repository name")).
		Doc("Get repository").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.RepositoryResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncRepositoryTag}))

	webservice.Route(webservice.GET("/repository").
		To(handler.ListRepository).
		Doc("List all repositories").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.ListRepositoryResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncRepositoryTag}))

	webservice.Route(webservice.DELETE("/repository/{name}").
		To(handler.DeleteRepository).
		Param(webservice.PathParameter("name", "repository name")).
		Doc("Delete repository").
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncRepositoryTag}))

	// Backup
	webservice.Route(webservice.POST("/backup").
		To(handler.CreateBackup).
		Reads(ui_clustersync.BackupRequest{}).
		Doc("Create backup").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.BackupNameResponse{}).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncBackupTag}))

	webservice.Route(webservice.PUT("/backup/{name}").
		To(handler.UpdateBackup).
		Param(webservice.PathParameter("name", "backup name")).
		Reads(ui_clustersync.ModifyBackupRequest{}).
		Doc("Update backup").
		Notes(putNotes).
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncBackupTag}))

	webservice.Route(webservice.GET("/backup/{name}").
		To(handler.GetBackup).
		Param(webservice.PathParameter("name", "backup name")).
		Doc("Get backup").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.BackupResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncBackupTag}))

	webservice.Route(webservice.GET("/backup").
		To(handler.ListBackup).
		Doc("List all backups").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.ListBackupResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncBackupTag}))

	webservice.Route(webservice.DELETE("/backup/{name}").
		To(handler.DeleteBackup).
		Param(webservice.PathParameter("name", "backup name")).
		Doc("Delete backup").
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncBackupTag}))

	// Restore
	webservice.Route(webservice.POST("/restore").
		To(handler.CreateRestore).
		Reads(ui_clustersync.RestoreRequest{}).
		Doc("Create restore").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.RestoreNameResponse{}).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncRestoreTag}))

	webservice.Route(webservice.PUT("/restore/{name}").
		To(handler.UpdateRestore).
		Param(webservice.PathParameter("name", "restore name")).
		Reads(ui_clustersync.ModifyRestoreRequest{}).
		Doc("Update restore").
		Notes(putNotes).
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncRestoreTag}))

	webservice.Route(webservice.GET("/restore/{name}").
		To(handler.GetRestore).
		Param(webservice.PathParameter("name", "restore name")).
		Doc("Get restore").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.RestoreResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncRestoreTag}))

	webservice.Route(webservice.GET("/restore").
		To(handler.ListRestore).
		Doc("List all restores").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.ListRestoreResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncRestoreTag}))

	webservice.Route(webservice.DELETE("/restore/{name}").
		To(handler.DeleteRestore).
		Param(webservice.PathParameter("name", "restore name")).
		Doc("Delete restore").
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncRestoreTag}))

	// Schedule
	webservice.Route(webservice.POST("/schedule").
		To(handler.CreateSchedule).
		Reads(ui_clustersync.ScheduleRequest{}).
		Doc("Create schedule").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.ScheduleNameResponse{}).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncScheduleTag}))

	webservice.Route(webservice.PUT("/schedule/{name}").
		To(handler.UpdateSchedule).
		Param(webservice.PathParameter("name", "schedule name")).
		Reads(ui_clustersync.ModifyScheduleRequest{}).
		Doc("Update schedule").
		Notes(putNotes).
		Returns(http.StatusOK, api.StatusOK, nil).
		Returns(http.StatusForbidden, "Invalid format", util.BadRequestError{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncScheduleTag}))

	webservice.Route(webservice.GET("/schedule/{name}").
		To(handler.GetSchedule).
		Param(webservice.PathParameter("name", "schedule name")).
		Doc("Get schedule").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.ScheduleResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncScheduleTag}))

	webservice.Route(webservice.GET("/schedule").
		To(handler.ListSchedule).
		Doc("List all schedules").
		Returns(http.StatusOK, api.StatusOK, ui_clustersync.ListScheduleResponse{}).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterSyncScheduleTag}))

	webservice.Route(webservice.DELETE("/schedule/{name}").
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
