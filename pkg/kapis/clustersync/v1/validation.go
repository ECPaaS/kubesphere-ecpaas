/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com
*/

package v1

import (
	"net"
	"net/http"
	"reflect"
	"time"

	"github.com/adhocore/gronx"
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/kapis/util"
	ui_clustersync "kubesphere.io/kubesphere/pkg/models/clustersync"
)

// RepositoryConfig

func isValidRepositoryRequest(request *ui_clustersync.RepositoryRequest, resp *restful.Response) bool {
	reflectType := reflect.TypeOf(*request)
	// RepositoryName string
	if request.RepositoryName == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "RepositoryName must not be empty.",
		})
		return false
	} else if !isValidOptionalStringField(reflectType, request.RepositoryName, "RepositoryName", resp) {
		return false
	}
	// Provider string
	if request.Provider == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "Provider must not be empty.",
		})
		return false
	} else if !isValidOptionalStringField(reflectType, request.Provider, "Provider", resp) {
		return false
	}
	// Bucket string
	if request.Bucket == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "Bucket must not be empty.",
		})
		return false
	} else if !isValidOptionalStringField(reflectType, request.Bucket, "Bucket", resp) {
		return false
	}
	// Prefix string
	if request.Prefix == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "Prefix must not be empty.",
		})
		return false
	} else if !isValidOptionalStringField(reflectType, request.Prefix, "Prefix", resp) {
		return false
	}
	// Region string
	if request.Region == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "Region must not be empty.",
		})
		return false
	} else if !isValidOptionalStringField(reflectType, request.Region, "Region", resp) {
		return false
	}
	// Ip string
	if request.Ip == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "Ip must not be empty.",
		})
		return false
	} else if !isValidOptionalIpAddress(request.Ip, resp) {
		return false
	}
	// Port *int
	if request.Port == nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "Port must not be empty.",
		})
		return false
	} else if !isValidOptionalPortNumber(reflectType, request.Port, "Port", resp) {
		return false
	}
	// AccessKey string
	if request.AccessKey == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "AccessKey must not be empty.",
		})
		return false
	} else if !util.IsValidLength(reflectType, request.AccessKey, "AccessKey", resp) {
		return false
	}
	// SecretKey string
	if request.SecretKey == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "SecretKey must not be empty.",
		})
		return false
	} else if !util.IsValidLength(reflectType, request.SecretKey, "SecretKey", resp) {
		return false
	}

	return true
}

func isValidRepositoryModifyRequest(request *ui_clustersync.ModifyRepositoryRequest, resp *restful.Response) bool {
	reflectType := reflect.TypeOf(*request)
	// Provider string
	if !isValidOptionalStringField(reflectType, request.Provider, "Provider", resp) {
		return false
	}
	// Bucket string
	if !isValidOptionalStringField(reflectType, request.Bucket, "Bucket", resp) {
		return false
	}
	// Prefix string
	if !isValidOptionalStringField(reflectType, request.Prefix, "Prefix", resp) {
		return false
	}
	// Region string
	if !isValidOptionalStringField(reflectType, request.Region, "Region", resp) {
		return false
	}
	// Ip string
	if !isValidOptionalIpAddress(request.Ip, resp) {
		return false
	}
	// Port *int
	if !isValidOptionalPortNumber(reflectType, request.Port, "Port", resp) {
		return false
	}
	// AccessKey string
	if !util.IsValidLength(reflectType, request.AccessKey, "AccessKey", resp) {
		return false
	}
	// SecretKey string
	if !util.IsValidLength(reflectType, request.SecretKey, "SecretKey", resp) {
		return false
	}

	return true
}


// BackupConfig

func isValidBackupRequest(request *ui_clustersync.BackupRequest, resp *restful.Response) bool {
	reflectType := reflect.TypeOf(*request)
	// BackupName string
	if request.BackupName == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "BackupName must not be empty.",
		})
		return false
	} else if !isValidOptionalStringField(reflectType, request.BackupName, "BackupName", resp) {
		return false
	}
	// IncludedNamespaces []string
	// ExcludedNamespaces []string
	if !isValidNamespaceRange(request.IncludedNamespaces, request.ExcludedNamespaces, resp) {
		return false
	}
	// TTL string
	if !isValidTTL(request.TTL, resp) {
		return false
	}
	// BackupRepository string
	if !isValidOptionalStringField(reflectType, request.BackupRepository, "BackupRepository", resp) {
		return false
	}
	// SnapshotRepositories []string
	if !isValidSnapshotRepositories(request.SnapshotRepositories, resp) {
		return false
	}
	// IsOneTime *bool
	if request.IsOneTime == nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "IsOneTime must be set.",
		})
		return false
	}

	return true
}

func isValidBackupModifyRequest(request *ui_clustersync.ModifyBackupRequest, resp *restful.Response) bool {
	reflectType := reflect.TypeOf(*request)
	// IncludedNamespaces []string
	// ExcludedNamespaces []string
	if !isValidNamespaceRange(request.IncludedNamespaces, request.ExcludedNamespaces, resp) {
		return false
	}
	// TTL *string
	if request.TTL != nil && !isValidTTL(*request.TTL, resp) {
		return false
	}
	// BackupRepository *string
	if request.BackupRepository != nil && !isValidOptionalStringField(reflectType, *request.BackupRepository, "BackupRepository", resp) {
		return false
	}
	// SnapshotRepositories []string
	if !isValidSnapshotRepositories(request.SnapshotRepositories, resp) {
		return false
	}

	return true
}


// RestoreConfig

func isValidRestoreRequest(request *ui_clustersync.RestoreRequest, resp *restful.Response) bool {
	reflectType := reflect.TypeOf(*request)
	// RestoreName string
	if request.RestoreName == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "RestoreName must not be empty.",
		})
		return false
	} else if !isValidOptionalStringField(reflectType, request.RestoreName, "RestoreName", resp) {
		return false
	}
	// BackupName string
	if request.BackupSource == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "BackupName must not be empty.",
		})
		return false
	} else if !isValidOptionalStringField(reflectType, request.BackupSource, "BackupName", resp) {
		return false
	}
	// IncludedNamespaces []string
	// ExcludedNamespaces []string
	if !isValidNamespaceRange(request.IncludedNamespaces, request.ExcludedNamespaces, resp) {
		return false
	}
	// IsOneTime *bool
	if request.IsOneTime == nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "IsOneTime must be set.",
		})
		return false
	}

	return true
}

func isValidRestoreModifyRequest(request *ui_clustersync.ModifyRestoreRequest, resp *restful.Response) bool {
	reflectType := reflect.TypeOf(*request)
	// BackupName string
	if !isValidOptionalStringField(reflectType, request.BackupSource, "BackupName", resp) {
		return false
	}
	// IncludedNamespaces []string
	// ExcludedNamespaces []string
	if !isValidNamespaceRange(request.IncludedNamespaces, request.ExcludedNamespaces, resp) {
		return false
	}

	return true
}


// ScheduleConfig

func isValidScheduleRequest(request *ui_clustersync.ScheduleRequest, resp *restful.Response) bool {
	reflectType := reflect.TypeOf(*request)
	// ScheduleName string
	if request.ScheduleName == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "ScheduleName must not be empty.",
		})
		return false
	} else if !isValidOptionalStringField(reflectType, request.ScheduleName, "ScheduleName", resp) {
		return false
	}
	// Schedule string
	if request.Schedule == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "Schedule must not be empty.",
		})
		return false
	} else if !isValidCronString(request.Schedule, resp) {
		return false
	}
	// Template struct
	if !isValidSchedulePostTemplate(&request.Template, resp) {
		return false
	}

	return true
}

func isValidScheduleModifyRequest(request *ui_clustersync.ModifyScheduleRequest, resp *restful.Response) bool {
	// Schedule string
	if request.Schedule != "" && !isValidCronString(request.Schedule, resp) {
		return false
	}
	// Template *struct
	if request.Template != nil && !isValidSchedulePutTemplate(request.Template, resp) {
		return false
	}

	return true
}

func isValidSchedulePostTemplate(template *ui_clustersync.PostTemplate, resp *restful.Response) bool {
	reflectType := reflect.TypeOf(*template)
	// IncludedNamespaces []string
	// ExcludedNamespaces []string
	if !isValidNamespaceRange(template.IncludedNamespaces, template.ExcludedNamespaces, resp) {
		return false
	}
	// TTL *string
	if !isValidTTL(template.TTL, resp) {
		return false
	}
	// BackupRepository *string
	if !isValidOptionalStringField(reflectType, template.BackupRepository, "BackupRepository", resp) {
		return false
	}
	// SnapshotRepositories []string
	if !isValidSnapshotRepositories(template.SnapshotRepositories, resp) {
		return false
	}

	return true
}

func isValidSchedulePutTemplate(template *ui_clustersync.PutTemplate, resp *restful.Response) bool {
	reflectType := reflect.TypeOf(*template)
	// IncludedNamespaces []string
	// ExcludedNamespaces []string
	if !isValidNamespaceRange(template.IncludedNamespaces, template.ExcludedNamespaces, resp) {
		return false
	}
	// TTL *string
	if template.TTL != nil && !isValidTTL(*template.TTL, resp) {
		return false
	}
	// BackupRepository *string
	if template.BackupRepository != nil && !isValidOptionalStringField(reflectType, *template.BackupRepository, "BackupRepository", resp) {
		return false
	}
	// SnapshotRepositories []string
	if !isValidSnapshotRepositories(template.SnapshotRepositories, resp) {
		return false
	}

	return true
}

// Valid characters: A-Z, a-z, 0-9, and -(hyphen).
// Valid length: <= maximum.
func isValidOptionalStringField(validateType reflect.Type, value string, fieldName string, resp *restful.Response) bool {
	if value != "" {
		if !util.IsValidLength(validateType, value, fieldName, resp) {
			return false
		} else if !util.IsValidCaseInsensitiveString(value, resp) {
			return false
		}
	}
	return true
}

func isValidOptionalIpAddress(ip string, resp *restful.Response) bool {
	if ip != "" {
		if net.ParseIP(ip) == nil {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
				Reason: "Invalid IP address: " + ip,
			})
			return false
		}
	}
	return true
}

func isValidOptionalPortNumber(validateType reflect.Type, port *int, fieldName string, resp *restful.Response) bool {
	if port != nil {
		if !util.IsValidWithinRange(validateType, *port, fieldName, resp) {
			return false
		}
	}
	return true
}

func isValidNamespaceRange(included []string, excluded []string, resp *restful.Response) bool {
	includedMap := make(map[string]bool, 0)
	excludedMap := make(map[string]bool, 0)
	for _, ns := range included {
		if !util.IsValidString(ns, resp) {
			return false
		}
		if _, ok := includedMap[ns]; ok {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
				Reason: "Invalid IncludedNamespaces, must not contain duplicated namespaces.",
			})
			return false
		}
		includedMap[ns] = true
	}
	for _, ns := range excluded {
		if !util.IsValidString(ns, resp) {
			return false
		}
		if _, ok := includedMap[ns]; ok {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
				Reason: "Invalid ExcludedNamespaces, must not be overlapped with IncludedNamespaces.",
			})
			return false
		}
		if _, ok := excludedMap[ns]; ok {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
				Reason: "Invalid ExcludedNamespace, must not contain duplicated namespaces.",
			})
			return false
		}
		excludedMap[ns] = true
	}
	return true
}

func isValidTTL(ttl string, resp *restful.Response) bool {
	if ttl != "" {
		if _, err := time.ParseDuration(ttl); err != nil {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
				Reason: "Invalid TTL : " + err.Error(),
			})
			return false
		}
	}
	return true
}

func isValidSnapshotRepositories(locations []string, resp *restful.Response) bool {
	locationMap := make(map[string]bool, 0)
	for _, location := range locations {
		if !util.IsValidString(location, resp) {
			return false
		}
		if _, ok := locationMap[location]; ok {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
				Reason: "Invalid SnapshotRepositories, must not contain duplicated repositories.",
			})
			return false
		}
		locationMap[location] = true
	}
	return true
}

func isValidCronString(expression string, resp *restful.Response) bool {
	gron := gronx.New()
	if !gron.IsValid(expression) {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "Invalid Schedule : " + expression,
		})
		return false
	}
	return true
}
