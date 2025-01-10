/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com
*/

package clustersync

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	clustersyncv1 "kubesphere.io/api/clustersync/v1"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
)

const (
	OperatorConfigName = "operatorconfig"
	OperatorConfigNamespace = "default"
	DefaultTTL = "720h"
)

type Interface interface {
	// Repository
	CreateRepository(ui_repository *RepositoryRequest) (*RepositoryNameResponse, error)
	UpdateRepository(name string, ui_repository *ModifyRepositoryRequest) (*RepositoryResponse, error)
	GetRepository(name string) (*RepositoryResponse, error)
	ListRepository() (*ListRepositoryResponse, error)
	DeleteRepository(name string) error

	// Backup
	CreateBackup(ui_backup *BackupRequest) (*BackupNameResponse, error)
	UpdateBackup(name string, ui_backup *ModifyBackupRequest) (*BackupResponse, error)
	GetBackup(name string) (*BackupResponse, error)
	ListBackup() (*ListBackupResponse, error)
	DeleteBackup(name string) error

	// Restore
	CreateRestore(ui_restore *RestoreRequest) (*RestoreNameResponse, error)
	UpdateRestore(name string, ui_restore *ModifyRestoreRequest) (*RestoreResponse, error)
	GetRestore(name string) (*RestoreResponse, error)
	ListRestore() (*ListRestoreResponse, error)
	DeleteRestore(name string) error

	// Schedule
	CreateSchedule(ui_schedule *ScheduleRequest) (*ScheduleNameResponse, error)
	UpdateSchedule(name string, ui_schedule *ModifyScheduleRequest) (*ScheduleResponse, error)
	GetSchedule(name string) (*ScheduleResponse, error)
	ListSchedule() (*ListScheduleResponse, error)
	DeleteSchedule(name string) error
}

type clusterSyncOperator struct {
	ksclient  kubesphere.Interface
	k8sclient kubernetes.Interface
}

func New(ksclient kubesphere.Interface, k8sclient kubernetes.Interface) Interface {
	return &clusterSyncOperator{
		ksclient:  ksclient,
		k8sclient: k8sclient,
	}
}


// Repository

func (cs *clusterSyncOperator) CreateRepository(ui_repository *RepositoryRequest) (*RepositoryNameResponse, error) {
	klog.V(2).Infof("Creating Repository: \"%s\"", ui_repository.RepositoryName)
	// Get OperatorConfig
	createFlag := false
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			//create config
			config = &clustersyncv1.OperatorConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: OperatorConfigName,
					Namespace: OperatorConfigNamespace,
				},
				Spec: clustersyncv1.OperatorConfigSpec{
					RepositoryConfigs: nil,
					BackupConfigs: nil,
					RestoreConfigs: nil,
					ScheduleConfigs: nil,
				},
			}
			createFlag = true
		} else {
			return nil, err
		}
	}

	// Check if no redundant RepositoryConfig add
	if getRepositoryConfig(config.Spec.RepositoryConfigs, ui_repository.RepositoryName) != nil {
		// Duplicated, error
		return nil, fmt.Errorf("repository \"%s\" duplicated", ui_repository.RepositoryName)
	} else {
		// Check if try to set duplicated default repository
		if ui_repository.IsDefault != nil && *ui_repository.IsDefault {
			if anyDefaultRepository(config.Spec.RepositoryConfigs, "") {
				return nil, fmt.Errorf("default repository already exists")
			}
		}
		// New, create config
		newRepositoryConfig := clustersyncv1.RepositoryConfig{
			RepositoryName: ui_repository.RepositoryName,
			Provider:    ui_repository.Provider,
			Bucket:      ui_repository.Bucket,
			Prefix:      ui_repository.Prefix,
			Region:      ui_repository.Region,
			Ip:          ui_repository.Ip,
			Port:        ui_repository.Port,
			AccessKey:   ui_repository.AccessKey,
			SecretKey:   ui_repository.SecretKey,
			IsDefault:   ui_repository.IsDefault,
			LastModified: time.Now().String(),
		}
		config.Spec.RepositoryConfigs = append(config.Spec.RepositoryConfigs, newRepositoryConfig)
		if createFlag {
			_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Create(context.Background(), config, metav1.CreateOptions{})
		} else {
			_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Update(context.Background(), config, metav1.UpdateOptions{})
		}

		if err != nil {
			return nil, err
		} else {
			return &RepositoryNameResponse{RepositoryName: newRepositoryConfig.RepositoryName}, nil
		}
	}
}

func (cs *clusterSyncOperator) UpdateRepository(name string, ui_repository *ModifyRepositoryRequest) (*RepositoryResponse, error) {
	klog.V(2).Infof("Updating Repository: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("repository \"%s\" is not created", name)
		}
		return nil, err
	}
	if repositoryConfig := getRepositoryConfig(config.Spec.RepositoryConfigs, name); repositoryConfig != nil {
		// Found, update the Repository
		newConfig := *repositoryConfig.DeepCopy()
		if ui_repository.Provider != "" {
			newConfig.Provider = ui_repository.Provider
		}
		if ui_repository.Bucket != "" {
			newConfig.Bucket = ui_repository.Bucket
		}
		if ui_repository.Prefix != "" {
			newConfig.Prefix = ui_repository.Prefix
		}
		if ui_repository.Region != "" {
			newConfig.Region = ui_repository.Region
		}
		if ui_repository.Ip != "" {
			newConfig.Ip = ui_repository.Ip
		}
		if ui_repository.Port != nil {
			newConfig.Port = ui_repository.Port
		}
		if ui_repository.AccessKey != "" {
			newConfig.AccessKey = ui_repository.AccessKey
		}
		if ui_repository.SecretKey != "" {
			newConfig.SecretKey = ui_repository.SecretKey
		}
		if ui_repository.IsDefault != nil {
			if *ui_repository.IsDefault {
				if anyDefaultRepository(config.Spec.RepositoryConfigs, name) {
					return nil, fmt.Errorf("default repository already exists")
				}
			}
			newConfig.IsDefault = ui_repository.IsDefault
		}
		if !reflect.DeepEqual(*repositoryConfig, newConfig) {
			newConfig.LastModified = time.Now().String()
			newSlice := make([]clustersyncv1.RepositoryConfig, 0)
			for _, config := range config.Spec.RepositoryConfigs {
				if config.RepositoryName != name {
					newSlice = append(newSlice, config)
				}
			}
			newSlice = append(newSlice, newConfig)
			config.Spec.RepositoryConfigs = newSlice
			_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Update(context.Background(), config, metav1.UpdateOptions{})
			if err != nil {
				return nil, err
			} else {
				return makeRepositoryResponse(repositoryConfig, name), nil
			}
		}
		return nil, nil // No update
	} else {
		return nil, fmt.Errorf("repository \"%s\" is not created", name)
	}
}

func (cs *clusterSyncOperator) GetRepository(name string) (*RepositoryResponse, error) {
	klog.V(2).Infof("Getting Repository: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("repository \"%s\" is not found", name)
		}
		return nil, err
	}

	if repositoryConfig := getRepositoryConfig(config.Spec.RepositoryConfigs, name); repositoryConfig != nil {
		return makeRepositoryResponse(repositoryConfig, repositoryConfig.RepositoryName), nil
	} else {
		return nil, fmt.Errorf("repository \"%s\" is not found", name)
	}
}

func (cs *clusterSyncOperator) ListRepository() (*ListRepositoryResponse, error) {
	klog.V(2).Infof("Listing Repositories")
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}

	responseSlice := make([]RepositoryResponse, 0)
	if config != nil {
		repositoryConfigs := config.Spec.RepositoryConfigs
		for _, repositoryConfig := range repositoryConfigs {
			responseSlice = append(responseSlice, *makeRepositoryResponse(&repositoryConfig, repositoryConfig.RepositoryName))
		}
	}

	return &ListRepositoryResponse{TotalCount: len(responseSlice), Items: responseSlice}, nil
}

func (cs *clusterSyncOperator) DeleteRepository(name string) error {
	klog.V(2).Infof("Deleting Repository: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	newSlice := make([]clustersyncv1.RepositoryConfig, 0)
	for _, repositoryConfig := range config.Spec.RepositoryConfigs {
		if repositoryConfig.RepositoryName != name {
			newSlice = append(newSlice, repositoryConfig)
		}
	}
	config.Spec.RepositoryConfigs = newSlice

	_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Update(context.Background(), config, metav1.UpdateOptions{})
	if err != nil {
		return err
	} else {
		return nil
	}
}

func getRepositoryConfig(configs []clustersyncv1.RepositoryConfig, newConfigName string) *clustersyncv1.RepositoryConfig {
	for _, config := range configs {
		if config.RepositoryName == newConfigName {
			 return &config
		}
	}
	return nil
}

func makeRepositoryResponse(config *clustersyncv1.RepositoryConfig, name string) *RepositoryResponse {
	response := &RepositoryResponse{
		RepositoryName: name,
		Provider:       config.Provider,
		Bucket:         config.Bucket,
		Prefix:         config.Prefix,
		Region:         config.Region,
		Ip:             config.Ip,
		AccessKey:      base64.StdEncoding.EncodeToString([]byte(config.AccessKey)),
		SecretKey:      base64.StdEncoding.EncodeToString([]byte(config.SecretKey)),
	}
	if config.Port != nil {
		response.Port = *config.Port
	}
	if config.IsDefault != nil {
		response.IsDefault = *config.IsDefault
	}

	return response
}

func anyDefaultRepository(configs []clustersyncv1.RepositoryConfig, except string) bool {
	for _, config := range configs {
		if config.IsDefault != nil && *config.IsDefault && config.RepositoryName != except {
			// except is to tolerate set default repository default again
			return true
		}
	}
	return false
}

// Backup

func (cs *clusterSyncOperator) CreateBackup(ui_backup *BackupRequest) (*BackupNameResponse, error) {
	klog.V(2).Infof("Creating Backup: \"%s\"", ui_backup.BackupName)
	// Get OperatorConfig
	createFlag := false
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			//create config
			config = &clustersyncv1.OperatorConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: OperatorConfigName,
					Namespace: OperatorConfigNamespace,
				},
				Spec: clustersyncv1.OperatorConfigSpec{
					RepositoryConfigs: nil,
					BackupConfigs: nil,
					RestoreConfigs: nil,
					ScheduleConfigs: nil,
				},
			}
			createFlag = true
		} else {
			return nil, err
		}
	}

	// Check if no redundant BackupConfig add
	if getBackupConfig(config.Spec.BackupConfigs, ui_backup.BackupName) != nil {
		// Duplicated, error
		return nil, fmt.Errorf("backup \"%s\" duplicated", ui_backup.BackupName)
	} else {
		// New, create config
		backupSpec, err := makeBackupSpec(ui_backup)
		if err !=nil {
			return nil ,err
		}
		if !anyDefaultRepository(config.Spec.RepositoryConfigs, "") {
			if ui_backup.BackupRepository == "" {
				return nil ,fmt.Errorf("no default repository existed for BackupRepository")
			}
			if len(ui_backup.SnapshotRepositories) == 0 {
				return nil ,fmt.Errorf("no default repository existed for SnapshotRepositories")
			}
		}
		newBackupConfig := clustersyncv1.BackupConfig{
			BackupName: ui_backup.BackupName,
			BackupSpec: *backupSpec,
			IsOneTime: ui_backup.IsOneTime,
		}

		config.Spec.BackupConfigs = append(config.Spec.BackupConfigs, newBackupConfig)
		if createFlag {
			_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Create(context.Background(), config, metav1.CreateOptions{})
		} else {
			_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Update(context.Background(), config, metav1.UpdateOptions{})
		}

		if err != nil {
			return nil, err
		} else {
			return &BackupNameResponse{BackupName: newBackupConfig.BackupName}, nil
		}
	}
}

func (cs *clusterSyncOperator) UpdateBackup(name string, ui_backup *ModifyBackupRequest) (*BackupResponse, error) {
	klog.V(2).Infof("Updating Backup: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("backup \"%s\" is not created", name)
		}
		return nil, err
	}
	if backupConfig := getBackupConfig(config.Spec.BackupConfigs, name); backupConfig != nil {
		// Found, update the Backup
		if ui_backup.IncludedNamespaces != nil {
			backupConfig.BackupSpec.IncludedNamespaces = ui_backup.IncludedNamespaces
		}
		if ui_backup.ExcludedNamespaces != nil {
			backupConfig.BackupSpec.ExcludedNamespaces = ui_backup.ExcludedNamespaces
		}
		if ui_backup.TTL != nil {
			if duration, err := parseDurationOrDefault(*ui_backup.TTL); err != nil {
				return nil, err
			} else {
				backupConfig.BackupSpec.TTL = metav1.Duration{Duration: duration}
			}
		}
		if ui_backup.BackupRepository != nil {
			backupConfig.BackupSpec.StorageLocation = *ui_backup.BackupRepository
		}
		if ui_backup.DefaultVolumesToFsBackup != nil {
			backupConfig.BackupSpec.DefaultVolumesToFsBackup = ui_backup.DefaultVolumesToFsBackup
		}
		if ui_backup.SnapshotRepositories != nil {
			backupConfig.BackupSpec.VolumeSnapshotLocations = ui_backup.SnapshotRepositories
		}
		if ui_backup.SnapshotMoveData != nil {
			backupConfig.BackupSpec.SnapshotMoveData = ui_backup.SnapshotMoveData
		}
		if !anyDefaultRepository(config.Spec.RepositoryConfigs, "") {
			if ui_backup.BackupRepository != nil && *ui_backup.BackupRepository == "" {
				// try to clear BackupRepository when no default repository
				return nil ,fmt.Errorf("no default repository existed for BackupRepository")
			}
			if len(ui_backup.SnapshotRepositories) == 0 {
				// try to clear SnapshotRepositories when no default repository
				return nil ,fmt.Errorf("no default repository existed for SnapshotRepositories")
			}
		}

		newSlice := make([]clustersyncv1.BackupConfig, 0)
		for _, config := range config.Spec.BackupConfigs {
			if config.BackupName == name {
				newSlice = append(newSlice, *backupConfig)
			} else {
				newSlice = append(newSlice, config)
			}
		}
		config.Spec.BackupConfigs = newSlice
		_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Update(context.Background(), config, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		} else {
			return makeBackupResponse(&backupConfig.BackupSpec, name), nil
		}
	} else {
		return nil, fmt.Errorf("backup \"%s\" is not created", name)
	}
}

func (cs *clusterSyncOperator) GetBackup(name string) (*BackupResponse, error) {
	klog.V(2).Infof("Getting Backup: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("backup \"%s\" is not found", name)
		}
		return nil, err
	}

	if backupConfig := getBackupConfig(config.Spec.BackupConfigs, name); backupConfig != nil {
		return makeBackupResponse(&backupConfig.BackupSpec, backupConfig.BackupName), nil
	} else {
		return nil, fmt.Errorf("backup \"%s\" is not found", name)
	}
}

func (cs *clusterSyncOperator) ListBackup() (*ListBackupResponse, error) {
	klog.V(2).Infof("Listing Backups")
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}

	responseSlice := make([]BackupResponse, 0)
	if config != nil {
		backupConfigs := config.Spec.BackupConfigs
		for _, backupConfig := range backupConfigs {
			if backupConfig.IsOneTime != nil && !*backupConfig.IsOneTime {
				responseSlice = append(responseSlice, *makeBackupResponse(&backupConfig.BackupSpec, backupConfig.BackupName))
			}
		}
	}

	return &ListBackupResponse{TotalCount: len(responseSlice), Items: responseSlice}, nil
}

func (cs *clusterSyncOperator) DeleteBackup(name string) error {
	klog.V(2).Infof("Deleting Backup: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	newSlice := make([]clustersyncv1.BackupConfig, 0)
	for _, backupConfig := range config.Spec.BackupConfigs {
		if backupConfig.BackupName != name {
			newSlice = append(newSlice, backupConfig)
		}
	}
	config.Spec.BackupConfigs = newSlice

	_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Update(context.Background(), config, metav1.UpdateOptions{})
	if err != nil {
		return err
	} else {
		return nil
	}
}

func getBackupConfig(configs []clustersyncv1.BackupConfig, newConfigName string) *clustersyncv1.BackupConfig {
	for _, config := range configs {
		if config.BackupName == newConfigName && !*config.IsOneTime {
			 return &config
		}
	}
	return nil
}

func makeBackupSpec(request *BackupRequest) (*clustersyncv1.BackupSpec, error) {
	backupSpec := &clustersyncv1.BackupSpec{
		IncludedNamespaces: request.IncludedNamespaces,
		ExcludedNamespaces: request.ExcludedNamespaces,
		StorageLocation: request.BackupRepository,
		DefaultVolumesToFsBackup: request.DefaultVolumesToFsBackup,
		VolumeSnapshotLocations: request.SnapshotRepositories,
		SnapshotMoveData: request.SnapshotMoveData,
	}
	if duration, err := parseDurationOrDefault(request.TTL); err != nil {
		return nil, err
	} else {
		backupSpec.TTL = metav1.Duration{Duration: duration}
	}
	return backupSpec, nil
}

func makeBackupResponse(backupSpec *clustersyncv1.BackupSpec, name string) *BackupResponse {
	response := &BackupResponse{
		BackupName: name,
		IncludedNamespaces:      backupSpec.IncludedNamespaces,
		ExcludedNamespaces:      backupSpec.ExcludedNamespaces,
		TTL:                     backupSpec.TTL.Duration.String(),
		BackupRepository:         backupSpec.StorageLocation,
		SnapshotRepositories: backupSpec.VolumeSnapshotLocations,
	}
	if response.IncludedNamespaces == nil {
		response.IncludedNamespaces = make([]string, 0)
	}
	if response.ExcludedNamespaces == nil {
		response.ExcludedNamespaces = make([]string, 0)
	}
	if response.SnapshotRepositories == nil {
		response.SnapshotRepositories = make([]string, 0)
	}
	if backupSpec.DefaultVolumesToFsBackup != nil {
		response.DefaultVolumesToFsBackup = *backupSpec.DefaultVolumesToFsBackup
	}
	if backupSpec.SnapshotMoveData != nil {
		response.SnapshotMoveData = *backupSpec.SnapshotMoveData
	}

	return response
}

func parseDurationOrDefault(durationStr string) (time.Duration, error) {
	if durationStr == "" {
		durationStr = DefaultTTL
	}
	if duration, err := time.ParseDuration(durationStr); err != nil {
		return 0, err
	} else {
		return duration, nil
	}
}


// Restore

func (cs *clusterSyncOperator) CreateRestore(ui_restore *RestoreRequest) (*RestoreNameResponse, error) {
	klog.V(2).Infof("Creating Restore: \"%s\"", ui_restore.RestoreName)
	// Get OperatorConfig
	createFlag := false
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			//create config
			config = &clustersyncv1.OperatorConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: OperatorConfigName,
					Namespace: OperatorConfigNamespace,
				},
				Spec: clustersyncv1.OperatorConfigSpec{
					RepositoryConfigs: nil,
					BackupConfigs: nil,
					RestoreConfigs: nil,
					ScheduleConfigs: nil,
				},
			}
			createFlag = true
		} else {
			return nil, err
		}
	}

	// Check if no redundant RestoreConfig add
	if getRestoreConfig(config.Spec.RestoreConfigs, ui_restore.RestoreName) != nil {
		// Duplicated, error
		return nil, fmt.Errorf("restore \"%s\" duplicated", ui_restore.RestoreName)
	} else {
		// New, create config
		newRestoreConfig := clustersyncv1.RestoreConfig{
			RestoreName: ui_restore.RestoreName,
			RestoreSpec: clustersyncv1.RestoreSpec{
				BackupName: ui_restore.BackupSource,
				IncludedNamespaces: ui_restore.IncludedNamespaces,
				ExcludedNamespaces: ui_restore.ExcludedNamespaces,
			},
			IsOneTime: ui_restore.IsOneTime,
		}
		config.Spec.RestoreConfigs = append(config.Spec.RestoreConfigs, newRestoreConfig)
		if createFlag {
			_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Create(context.Background(), config, metav1.CreateOptions{})
		} else {
			_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Update(context.Background(), config, metav1.UpdateOptions{})
		}

		if err != nil {
			return nil, err
		} else {
			return &RestoreNameResponse{RestoreName: newRestoreConfig.RestoreName}, nil
		}
	}
}

func (cs *clusterSyncOperator) UpdateRestore(name string, ui_restore *ModifyRestoreRequest) (*RestoreResponse, error) {
	klog.V(2).Infof("Updating RestoreConfig: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("restore \"%s\" is not created", name)
		}
		return nil, err
	}
	if restoreConfig := getRestoreConfig(config.Spec.RestoreConfigs, name); restoreConfig != nil {
		// Found, update the Restore
		if ui_restore.BackupSource != "" {
			restoreConfig.RestoreSpec.BackupName = ui_restore.BackupSource
		}
		if ui_restore.IncludedNamespaces != nil {
			restoreConfig.RestoreSpec.IncludedNamespaces = ui_restore.IncludedNamespaces
		}
		if ui_restore.ExcludedNamespaces != nil {
			restoreConfig.RestoreSpec.ExcludedNamespaces = ui_restore.ExcludedNamespaces
		}

		newSlice := make([]clustersyncv1.RestoreConfig, 0)
		for _, config := range config.Spec.RestoreConfigs {
			if config.RestoreName == name {
				newSlice = append(newSlice, *restoreConfig)
			} else {
				newSlice = append(newSlice, config)
			}
		}
		config.Spec.RestoreConfigs = newSlice
		_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Update(context.Background(), config, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		} else {
			return makeRestoreResponse(&restoreConfig.RestoreSpec, name), nil
		}
	} else {
		return nil, fmt.Errorf("restore \"%s\" is not created", name)
	}
}

func (cs *clusterSyncOperator) GetRestore(name string) (*RestoreResponse, error) {
	klog.V(2).Infof("Getting Restore: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("restore \"%s\" is not found", name)
		}
		return nil, err
	}

	if restoreConfig := getRestoreConfig(config.Spec.RestoreConfigs, name); restoreConfig != nil {
		return makeRestoreResponse(&restoreConfig.RestoreSpec, restoreConfig.RestoreName), nil
	} else {
		return nil, fmt.Errorf("restore \"%s\" is not found", name)
	}
}

func (cs *clusterSyncOperator) ListRestore() (*ListRestoreResponse, error) {
	klog.V(2).Infof("Listing Restores")
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}

	responseSlice := make([]RestoreResponse, 0)
	if config != nil {
		restoreConfigs := config.Spec.RestoreConfigs
		for _, restoreConfig := range restoreConfigs {
			if restoreConfig.IsOneTime != nil && !*restoreConfig.IsOneTime {
				responseSlice = append(responseSlice, *makeRestoreResponse(&restoreConfig.RestoreSpec, restoreConfig.RestoreName))
			}
		}
	}

	return &ListRestoreResponse{TotalCount: len(responseSlice), Items: responseSlice}, nil
}

func (cs *clusterSyncOperator) DeleteRestore(name string) error {
	klog.V(2).Infof("Deleting Restore: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	newSlice := make([]clustersyncv1.RestoreConfig, 0)
	for _, restoreConfig := range config.Spec.RestoreConfigs {
		if restoreConfig.RestoreName != name {
			newSlice = append(newSlice, restoreConfig)
		}
	}
	config.Spec.RestoreConfigs = newSlice

	_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Update(context.Background(), config, metav1.UpdateOptions{})
	if err != nil {
		return err
	} else {
		return nil
	}
}

func getRestoreConfig(configs []clustersyncv1.RestoreConfig, newConfigName string) *clustersyncv1.RestoreConfig {
	for _, config := range configs {
		if config.RestoreName == newConfigName && !*config.IsOneTime {
			 return &config
		}
	}
	return nil
}

func makeRestoreResponse(restoreSpec *clustersyncv1.RestoreSpec, name string) *RestoreResponse {
	response := &RestoreResponse{
		RestoreName:        name,
		BackupSource:       restoreSpec.BackupName,
		IncludedNamespaces: restoreSpec.IncludedNamespaces,
		ExcludedNamespaces: restoreSpec.ExcludedNamespaces,
	}
	if response.IncludedNamespaces == nil {
		response.IncludedNamespaces = make([]string, 0)
	}
	if response.ExcludedNamespaces == nil {
		response.ExcludedNamespaces = make([]string, 0)
	}

	return response
}


// Schedule

func (cs *clusterSyncOperator) CreateSchedule(ui_schedule *ScheduleRequest) (*ScheduleNameResponse, error) {
	klog.V(2).Infof("Creating Schedule: \"%s\"", ui_schedule.ScheduleName)
	// Get OperatorConfig
	createFlag := false
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			//create config
			config = &clustersyncv1.OperatorConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: OperatorConfigName,
					Namespace: OperatorConfigNamespace,
				},
				Spec: clustersyncv1.OperatorConfigSpec{
					RepositoryConfigs: nil,
					BackupConfigs: nil,
					RestoreConfigs: nil,
					ScheduleConfigs: nil,
				},
			}
			createFlag = true
		} else {
			return nil, err
		}
	}

	// Check if no redundant ScheduleConfig add
	if getScheduleConfig(config.Spec.ScheduleConfigs, ui_schedule.ScheduleName) != nil {
		// Duplicated, error
		return nil, fmt.Errorf("schedule \"%s\" duplicated", ui_schedule.ScheduleName)
	} else {
		// New, create config
		newScheduleConfig := clustersyncv1.ScheduleConfig{
			ScheduleName: ui_schedule.ScheduleName,
			ScheduleSpec: clustersyncv1.ScheduleSpec{
				Schedule: ui_schedule.Schedule,
				Template: clustersyncv1.BackupSpec{
					IncludedNamespaces: ui_schedule.Template.IncludedNamespaces,
					ExcludedNamespaces: ui_schedule.Template.ExcludedNamespaces,
					DefaultVolumesToFsBackup: ui_schedule.Template.DefaultVolumesToFsBackup,
					StorageLocation: ui_schedule.Template.BackupRepository,
					VolumeSnapshotLocations: ui_schedule.Template.SnapshotRepositories,
					SnapshotMoveData: ui_schedule.Template.SnapshotMoveData,
				},
			},
			LastModified: time.Now().String(),
		}
		if ui_schedule.Paused != nil {
			newScheduleConfig.ScheduleSpec.Paused = *ui_schedule.Paused
		}
		if duration, err := parseDurationOrDefault(ui_schedule.Template.TTL); err != nil {
			return nil, err
		} else {
			newScheduleConfig.ScheduleSpec.Template.TTL = metav1.Duration{Duration: duration}
		}
		if !anyDefaultRepository(config.Spec.RepositoryConfigs, "") {
			if ui_schedule.Template.BackupRepository == "" {
				return nil, fmt.Errorf("no default repository existed for BackupRepository")
			}
			if len(ui_schedule.Template.SnapshotRepositories) == 0 {
				return nil, fmt.Errorf("no default repository existed for SnapshotRepositories")
			}
		}

		config.Spec.ScheduleConfigs = append(config.Spec.ScheduleConfigs, newScheduleConfig)
		if createFlag {
			_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Create(context.Background(), config, metav1.CreateOptions{})
		} else {
			_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Update(context.Background(), config, metav1.UpdateOptions{})
		}

		if err != nil {
			return nil, err
		} else {
			return &ScheduleNameResponse{ScheduleName: newScheduleConfig.ScheduleName}, nil
		}
	}
}

func (cs *clusterSyncOperator) UpdateSchedule(name string, ui_schedule *ModifyScheduleRequest) (*ScheduleResponse, error) {
	klog.V(2).Infof("Updating Schedule: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("schedule \"%s\" is not created", name)
		}
		return nil, err
	}
	if scheduleConfig := getScheduleConfig(config.Spec.ScheduleConfigs, name); scheduleConfig != nil {
		// Found, update the Schedule
		newConfig := *scheduleConfig.DeepCopy()
		if ui_schedule.Schedule != "" {
			newConfig.ScheduleSpec.Schedule = ui_schedule.Schedule
		}
		if ui_schedule.Paused != nil {
			newConfig.ScheduleSpec.Paused = *ui_schedule.Paused
		}
		if ui_schedule.Template != nil {
			newConfig.ScheduleSpec.Template = clustersyncv1.BackupSpec{
				IncludedNamespaces: ui_schedule.Template.IncludedNamespaces,
				ExcludedNamespaces: ui_schedule.Template.ExcludedNamespaces,
				DefaultVolumesToFsBackup: ui_schedule.Template.DefaultVolumesToFsBackup,
				VolumeSnapshotLocations: ui_schedule.Template.SnapshotRepositories,
				SnapshotMoveData: ui_schedule.Template.SnapshotMoveData,
			}
			if ui_schedule.Template.BackupRepository != nil {
				newConfig.ScheduleSpec.Template.StorageLocation = *ui_schedule.Template.BackupRepository
			}
			if ui_schedule.Template.TTL != nil {
				if duration, err := time.ParseDuration(*ui_schedule.Template.TTL); err != nil {
					return nil, err
				} else {
					newConfig.ScheduleSpec.Template.TTL = metav1.Duration{Duration: duration}
				}
			}
			if !anyDefaultRepository(config.Spec.RepositoryConfigs, "") {
				if ui_schedule.Template.BackupRepository != nil && *ui_schedule.Template.BackupRepository == "" {
					return nil, fmt.Errorf("no default repository existed for BackupRepository")
				}
				if len(ui_schedule.Template.SnapshotRepositories) == 0 {
					return nil, fmt.Errorf("no default repository existed for SnapshotRepositories")
				}
			}
		}
		if !reflect.DeepEqual(scheduleConfig.ScheduleSpec, newConfig.ScheduleSpec) {
			newConfig.LastModified = time.Now().String()

			newSlice := make([]clustersyncv1.ScheduleConfig, 0)
			for _, config := range config.Spec.ScheduleConfigs {
				if config.ScheduleName != name {
					newSlice = append(newSlice, config)
				}
			}
			newSlice = append(newSlice, newConfig)
			config.Spec.ScheduleConfigs = newSlice
			_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Update(context.Background(), config, metav1.UpdateOptions{})
			if err != nil {
				return nil, err
			} else {
				return makeScheduleResponse(scheduleConfig, name), nil
			}
		}
		return nil, nil // No update
	} else {
		return nil, fmt.Errorf("schedule \"%s\" is not created", name)
	}
}

func (cs *clusterSyncOperator) GetSchedule(name string) (*ScheduleResponse, error) {
	klog.V(2).Infof("Getting Schedule: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("schedule \"%s\" is not found", name)
		}
		return nil, err
	}

	if scheduleConfig := getScheduleConfig(config.Spec.ScheduleConfigs, name); scheduleConfig != nil {
		return makeScheduleResponse(scheduleConfig, scheduleConfig.ScheduleName), nil
	} else {
		return nil, fmt.Errorf("schedule \"%s\" is not found", name)
	}
}

func (cs *clusterSyncOperator) ListSchedule() (*ListScheduleResponse, error) {
	klog.V(2).Infof("Listing Schedules")
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}

	responseSlice := make([]ScheduleResponse, 0)
	if config != nil {
		scheduleConfigs := config.Spec.ScheduleConfigs
		for _, scheduleConfig := range scheduleConfigs {
			responseSlice = append(responseSlice, *makeScheduleResponse(&scheduleConfig, scheduleConfig.ScheduleName))
		}
	}

	return &ListScheduleResponse{TotalCount: len(responseSlice), Items: responseSlice}, nil
}

func (cs *clusterSyncOperator) DeleteSchedule(name string) error {
	klog.V(2).Infof("Deleting Schedule: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	newSlice := make([]clustersyncv1.ScheduleConfig, 0)
	for _, scheduleConfig := range config.Spec.ScheduleConfigs {
		if scheduleConfig.ScheduleName != name {
			newSlice = append(newSlice, scheduleConfig)
		}
	}
	config.Spec.ScheduleConfigs = newSlice

	_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Update(context.Background(), config, metav1.UpdateOptions{})
	if err != nil {
		return err
	} else {
		return nil
	}
}

func getScheduleConfig(configs []clustersyncv1.ScheduleConfig, newConfigName string) *clustersyncv1.ScheduleConfig {
	for _, config := range configs {
		if config.ScheduleName == newConfigName {
			 return &config
		}
	}
	return nil
}

func makeScheduleResponse(config *clustersyncv1.ScheduleConfig, name string) *ScheduleResponse {
	response := &ScheduleResponse{
		ScheduleName: name,
		Schedule:     config.ScheduleSpec.Schedule,
		Paused:       config.ScheduleSpec.Paused,
		Template:     *makeTemplateResponse(&config.ScheduleSpec.Template),
	}

	return response
}

func makeTemplateResponse(backupSpec *clustersyncv1.BackupSpec) *TemplateResponse {
	response := &TemplateResponse{
		IncludedNamespaces:      backupSpec.IncludedNamespaces,
		ExcludedNamespaces:      backupSpec.ExcludedNamespaces,
		TTL:                     backupSpec.TTL.Duration.String(),
		BackupRepository:         backupSpec.StorageLocation,
		SnapshotRepositories: backupSpec.VolumeSnapshotLocations,
	}
	if response.IncludedNamespaces == nil {
		response.IncludedNamespaces = make([]string, 0)
	}
	if response.ExcludedNamespaces == nil {
		response.ExcludedNamespaces = make([]string, 0)
	}
	if response.SnapshotRepositories == nil {
		response.SnapshotRepositories = make([]string, 0)
	}
	if backupSpec.DefaultVolumesToFsBackup != nil {
		response.DefaultVolumesToFsBackup = *backupSpec.DefaultVolumesToFsBackup
	}
	if backupSpec.SnapshotMoveData != nil {
		response.SnapshotMoveData = *backupSpec.SnapshotMoveData
	}

	return response
}
