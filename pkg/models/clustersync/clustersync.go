/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com
*/

package clustersync

import (
	"context"
	"fmt"
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
)

type Interface interface {
	// Storage
	CreateStorage(ui_storage *StorageRequest) (*StorageNameResponse, error)
	UpdateStorage(name string, ui_storage *ModifyStorageRequest) (*StorageResponse, error)
	GetStorage(name string) (*StorageResponse, error)
	ListStorage() (*ListStorageResponse, error)
	DeleteStorage(name string) error

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


// Storage

func (cs *clusterSyncOperator) CreateStorage(ui_storage *StorageRequest) (*StorageNameResponse, error) {
	klog.Infof("Creating StorageConfig: \"%s\"", ui_storage.StorageName)
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
					StorageConfigs: nil,
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

	// Check if no redundant StorageConfig add
	if getStorageConfig(config.Spec.StorageConfigs, ui_storage.StorageName) != nil {
		// Duplicated, error
		return nil, fmt.Errorf("StorageConfig \"%s\" duplicated", ui_storage.StorageName)
	} else {
		// New, create config
		newStorageConfig := clustersyncv1.StorageConfig{
			StorageName: ui_storage.StorageName,
			Provider:    ui_storage.Provider,
			Bucket:      ui_storage.Bucket,
			Prefix:      ui_storage.Prefix,
			Region:      ui_storage.Region,
			Ip:          ui_storage.Ip,
			Port:        ui_storage.Port,
			AccessKey:   ui_storage.AccessKey,
			SecretKey:   ui_storage.SecretKey,
			IsDefault:   ui_storage.IsDefault,
		}
		config.Spec.StorageConfigs = append(config.Spec.StorageConfigs, newStorageConfig)
		if createFlag {
			_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Create(context.Background(), config, metav1.CreateOptions{})
		} else {
			_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Update(context.Background(), config, metav1.UpdateOptions{})
		}

		if err != nil {
			return nil, err
		} else {
			return &StorageNameResponse{StorageName: newStorageConfig.StorageName}, nil
		}
	}
}

func (cs *clusterSyncOperator) UpdateStorage(name string, ui_storage *ModifyStorageRequest) (*StorageResponse, error) {
	klog.Infof("Updating StorageConfig: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("OperatorConfig is not created")
		}
		return nil, err
	}
	if storageConfig := getStorageConfig(config.Spec.StorageConfigs, name); storageConfig != nil {
		// Duplicated, update the StorageConfig
		if ui_storage.Provider != "" {
			storageConfig.Provider = ui_storage.Provider
		}
		if ui_storage.Bucket != "" {
			storageConfig.Bucket = ui_storage.Bucket
		}
		if ui_storage.Prefix != "" {
			storageConfig.Prefix = ui_storage.Prefix
		}
		if ui_storage.Region != "" {
			storageConfig.Region = ui_storage.Region
		}
		if ui_storage.Ip != "" {
			storageConfig.Ip = ui_storage.Ip
		}
		if ui_storage.Port != nil {
			storageConfig.Port = ui_storage.Port
		}
		if ui_storage.AccessKey != "" {
			storageConfig.AccessKey = ui_storage.AccessKey
		}
		if ui_storage.SecretKey != "" {
			storageConfig.SecretKey = ui_storage.SecretKey
		}
		if ui_storage.IsDefault != nil {
			storageConfig.IsDefault = ui_storage.IsDefault
		}

		newSlice := make([]clustersyncv1.StorageConfig, 0)
		for _, config := range config.Spec.StorageConfigs {
			if config.StorageName != name {
				newSlice = append(newSlice, config)
			}
		}
		config.Spec.StorageConfigs = newSlice
		newConfig, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Update(context.Background(), config, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("updating StorageConfig \"%s\" failed", name)
		}

		newSlice = append(newSlice, *storageConfig)
		newConfig.Spec.StorageConfigs = newSlice
		_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Update(context.Background(), newConfig, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		} else {
			return makeStorageResponse(storageConfig, name), nil
		}
	} else {
		return nil, fmt.Errorf("StorageConfig \"%s\" is not created", name)
	}
}

func (cs *clusterSyncOperator) GetStorage(name string) (*StorageResponse, error) {
	klog.Infof("Getting StorageConfig: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("OperatorConfig is not created")
		}
		return nil, err
	}

	if storageConfig := getStorageConfig(config.Spec.StorageConfigs, name); storageConfig != nil {
		return makeStorageResponse(storageConfig, storageConfig.StorageName), nil
	} else {
		return nil, fmt.Errorf("StorageConfig \"%s\" is not found", name)
	}
}

func (cs *clusterSyncOperator) ListStorage() (*ListStorageResponse, error) {
	klog.Infof("Listing StorageConfigs")
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("OperatorConfig is not created")
		}
		return nil, err
	}

	storageConfigs := config.Spec.StorageConfigs
	responseSlice := make([]StorageResponse, 0)
	for _, storageConfig := range storageConfigs {
		responseSlice = append(responseSlice, *makeStorageResponse(&storageConfig, storageConfig.StorageName))
	}

	return &ListStorageResponse{TotalCount: len(responseSlice), Items: responseSlice}, nil
}

func (cs *clusterSyncOperator) DeleteStorage(name string) error {
	klog.Infof("Deleting StorageConfig: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Infof("OperatorConfig is not created yet")
			return fmt.Errorf("OperatorConfig is not created")
		}
		klog.Infof("Error getting OperatorConfig")
		return err
	}

	newSlice := make([]clustersyncv1.StorageConfig, 0)
	for _, storageConfig := range config.Spec.StorageConfigs {
		if storageConfig.StorageName != name {
			newSlice = append(newSlice, storageConfig)
		}
	}
	config.Spec.StorageConfigs = newSlice

	_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Update(context.Background(), config, metav1.UpdateOptions{})
	if err != nil {
		return err
	} else {
		return nil
	}
}

func getStorageConfig(configs []clustersyncv1.StorageConfig, newConfigName string) *clustersyncv1.StorageConfig {
	for _, config := range configs {
		if config.StorageName == newConfigName {
			 return &config
		}
	}
	return nil
}

func makeStorageResponse(config *clustersyncv1.StorageConfig, name string) *StorageResponse {
	response := &StorageResponse{
		StorageName: name,
		Provider:    config.Provider,
		Bucket:      config.Bucket,
		Prefix:      config.Prefix,
		Region:      config.Region,
		Ip:          config.Ip,
		AccessKey:   config.AccessKey,
		SecretKey:   config.SecretKey,
	}
	if config.Port != nil {
		response.Port = *config.Port
	}
	if config.IsDefault != nil {
		response.IsDefault = *config.IsDefault
	}

	return response
}

// Backup

func (cs *clusterSyncOperator) CreateBackup(ui_backup *BackupRequest) (*BackupNameResponse, error) {
	klog.Infof("Creating BackupConfig: \"%s\"", ui_backup.BackupName)
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
					StorageConfigs: nil,
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
		return nil, fmt.Errorf("BackupConfig \"%s\" duplicated", ui_backup.BackupName)
	} else {
		// New, create config
		backupSpec, err := makeBackpSpec(ui_backup)
		if err !=nil {
			return nil ,err
		}
		newBackupConfig := clustersyncv1.BackupConfig{
			BackupName: ui_backup.BackupName,
			BackupSpec: *backupSpec,
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
	klog.Infof("Updating BackupConfig: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("OperatorConfig is not created")
		}
		return nil, err
	}
	if backupConfig := getBackupConfig(config.Spec.BackupConfigs, name); backupConfig != nil {
		// Duplicated, update the BackupConfig
		if ui_backup.IncludedNamespaces != nil {
			backupConfig.BackupSpec.IncludedNamespaces = ui_backup.IncludedNamespaces
		}
		if ui_backup.ExcludedNamespaces != nil {
			backupConfig.BackupSpec.ExcludedNamespaces = ui_backup.ExcludedNamespaces
		}
		if ui_backup.TTL != "" {
			if duration, err := time.ParseDuration(ui_backup.TTL); err != nil {
				return nil, err
			} else {
				backupConfig.BackupSpec.TTL = metav1.Duration{Duration: duration}
			}
		}
		if ui_backup.StorageLocation != "" {
			backupConfig.BackupSpec.StorageLocation = ui_backup.StorageLocation
		}
		if ui_backup.DefaultVolumesToFsBackup != nil {
			backupConfig.BackupSpec.DefaultVolumesToFsBackup = ui_backup.DefaultVolumesToFsBackup
		}
		if ui_backup.VolumeSnapshotLocations != nil {
			backupConfig.BackupSpec.VolumeSnapshotLocations = ui_backup.VolumeSnapshotLocations
		}
		if ui_backup.SnapshotMoveData != nil {
			backupConfig.BackupSpec.SnapshotMoveData = ui_backup.SnapshotMoveData
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
		return nil, fmt.Errorf("BackupConfig \"%s\" is not created", name)
	}
}

func (cs *clusterSyncOperator) GetBackup(name string) (*BackupResponse, error) {
	klog.Infof("Getting BackupConfig: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("OperatorConfig is not created")
		}
		return nil, err
	}

	if backupConfig := getBackupConfig(config.Spec.BackupConfigs, name); backupConfig != nil {
		return makeBackupResponse(&backupConfig.BackupSpec, backupConfig.BackupName), nil
	} else {
		return nil, fmt.Errorf("BackupConfig \"%s\" is not found", name)
	}
}

func (cs *clusterSyncOperator) ListBackup() (*ListBackupResponse, error) {
	klog.Infof("Listing BackupConfigs")
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("OperatorConfig is not created")
		}
		return nil, err
	}

	backupConfigs := config.Spec.BackupConfigs
	responseSlice := make([]BackupResponse, 0)
	for _, backupConfig := range backupConfigs {
		responseSlice = append(responseSlice, *makeBackupResponse(&backupConfig.BackupSpec, backupConfig.BackupName))
	}

	return &ListBackupResponse{TotalCount: len(responseSlice), Items: responseSlice}, nil
}

func (cs *clusterSyncOperator) DeleteBackup(name string) error {
	klog.Infof("Deleting BackupConfig: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Infof("OperatorConfig is not created yet")
			return fmt.Errorf("OperatorConfig is not created")
		}
		klog.Infof("Error getting OperatorConfig")
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
		if config.BackupName == newConfigName {
			 return &config
		}
	}
	return nil
}

func makeBackpSpec(request *BackupRequest) (*clustersyncv1.BackupSpec, error) {
	backupSpec := &clustersyncv1.BackupSpec{
		IncludedNamespaces: request.IncludedNamespaces,
		ExcludedNamespaces: request.ExcludedNamespaces,
		StorageLocation: request.StorageLocation,
		DefaultVolumesToFsBackup: request.DefaultVolumesToFsBackup,
		VolumeSnapshotLocations: request.VolumeSnapshotLocations,
		SnapshotMoveData: request.SnapshotMoveData,
	}
	if request.TTL != "" { // TODO test this
		if duration, err := time.ParseDuration(request.TTL); err != nil {
			return nil, err
		} else {
			backupSpec.TTL = metav1.Duration{Duration: duration}
		}
	}
	return backupSpec, nil
}

func makeBackupResponse(backupSpec *clustersyncv1.BackupSpec, name string) *BackupResponse {
	response := &BackupResponse{
		BackupName: name,
		IncludedNamespaces:      backupSpec.IncludedNamespaces,
		ExcludedNamespaces:      backupSpec.ExcludedNamespaces,
		TTL:                     backupSpec.TTL.Duration.String(),
		StorageLocation:         backupSpec.StorageLocation,
		VolumeSnapshotLocations: backupSpec.VolumeSnapshotLocations,
	}
	if response.IncludedNamespaces == nil {
		response.IncludedNamespaces = make([]string, 0)
	}
	if response.ExcludedNamespaces == nil {
		response.ExcludedNamespaces = make([]string, 0)
	}
	if response.VolumeSnapshotLocations == nil {
		response.VolumeSnapshotLocations = make([]string, 0)
	}
	if backupSpec.DefaultVolumesToFsBackup != nil {
		response.DefaultVolumesToFsBackup = *backupSpec.DefaultVolumesToFsBackup
	}
	if backupSpec.SnapshotMoveData != nil {
		response.SnapshotMoveData = *backupSpec.SnapshotMoveData
	}

	return response
}


// Restore

func (cs *clusterSyncOperator) CreateRestore(ui_restore *RestoreRequest) (*RestoreNameResponse, error) {
	klog.Infof("Creating RestoreConfig: \"%s\"", ui_restore.RestoreName)
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
					StorageConfigs: nil,
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
		return nil, fmt.Errorf("RestoreConfig \"%s\" duplicated", ui_restore.RestoreName)
	} else {
		// New, create config
		newRestoreConfig := clustersyncv1.RestoreConfig{
			RestoreName: ui_restore.RestoreName,
			RestoreSpec: clustersyncv1.RestoreSpec{
				BackupName: ui_restore.BackupName,
				IncludedNamespaces: ui_restore.IncludedNamespaces,
				ExcludedNamespaces: ui_restore.ExcludedNamespaces,
			},
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
	klog.Infof("Updating RestoreConfig: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("OperatorConfig is not created")
		}
		return nil, err
	}
	if restoreConfig := getRestoreConfig(config.Spec.RestoreConfigs, name); restoreConfig != nil {
		// Duplicated, update the RestoreConfig
		if ui_restore.BackupName != "" {
			restoreConfig.RestoreSpec.BackupName = ui_restore.BackupName
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
		return nil, fmt.Errorf("RestoreConfig \"%s\" is not created", name)
	}
}

func (cs *clusterSyncOperator) GetRestore(name string) (*RestoreResponse, error) {
	klog.Infof("Getting RestoreConfig: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("OperatorConfig is not created")
		}
		return nil, err
	}

	if restoreConfig := getRestoreConfig(config.Spec.RestoreConfigs, name); restoreConfig != nil {
		return makeRestoreResponse(&restoreConfig.RestoreSpec, restoreConfig.RestoreName), nil
	} else {
		return nil, fmt.Errorf("RestoreConfig \"%s\" is not found", name)
	}
}

func (cs *clusterSyncOperator) ListRestore() (*ListRestoreResponse, error) {
	klog.Infof("Listing RestoreConfigs")
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("OperatorConfig is not created")
		}
		return nil, err
	}

	restoreConfigs := config.Spec.RestoreConfigs
	responseSlice := make([]RestoreResponse, 0)
	for _, restoreConfig := range restoreConfigs {
		responseSlice = append(responseSlice, *makeRestoreResponse(&restoreConfig.RestoreSpec, restoreConfig.RestoreName))
	}

	return &ListRestoreResponse{TotalCount: len(responseSlice), Items: responseSlice}, nil
}

func (cs *clusterSyncOperator) DeleteRestore(name string) error {
	klog.Infof("Deleting RestoreConfig: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Infof("OperatorConfig is not created yet")
			return fmt.Errorf("OperatorConfig is not created")
		}
		klog.Infof("Error getting OperatorConfig")
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
		if config.RestoreName == newConfigName {
			 return &config
		}
	}
	return nil
}

func makeRestoreResponse(restoreSpec *clustersyncv1.RestoreSpec, name string) *RestoreResponse {
	response := &RestoreResponse{
		RestoreName:        name,
		BackupName:         restoreSpec.BackupName,
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
	klog.Infof("Creating ScheduleConfig: \"%s\"", ui_schedule.ScheduleName)
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
					StorageConfigs: nil,
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
		return nil, fmt.Errorf("ScheduleConfig \"%s\" duplicated", ui_schedule.ScheduleName)
	} else {
		// New, create config
		backupSpec, err := makeBackpSpec(&ui_schedule.Template)
		if err != nil {
			return nil, err
		}
		newScheduleConfig := clustersyncv1.ScheduleConfig{
			ScheduleName: ui_schedule.ScheduleName,
			ScheduleSpec: clustersyncv1.ScheduleSpec{
				Schedule: ui_schedule.Schedule,
				Template: *backupSpec,
			},
		}
		if ui_schedule.Paused != nil {
			newScheduleConfig.ScheduleSpec.Paused = *ui_schedule.Paused
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
	klog.Infof("Updating ScheduleConfig: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("OperatorConfig is not created")
		}
		return nil, err
	}
	if scheduleConfig := getScheduleConfig(config.Spec.ScheduleConfigs, name); scheduleConfig != nil {
		// Duplicated, update the RestoreConfig
		if ui_schedule.Schedule != "" {
			scheduleConfig.ScheduleSpec.Schedule = ui_schedule.Schedule
		}
		if ui_schedule.Paused != nil {
			scheduleConfig.ScheduleSpec.Paused = *ui_schedule.Paused
		}
		if ui_schedule.Template != nil {
			backupSpec, err := makeBackpSpec(ui_schedule.Template)
			if err != nil {
				return nil, err
			}
			scheduleConfig.ScheduleSpec.Template = *backupSpec
		}

		newSlice := make([]clustersyncv1.ScheduleConfig, 0)
		for _, config := range config.Spec.ScheduleConfigs {
			if config.ScheduleName != name {
				newSlice = append(newSlice, config)
			}
		}
		config.Spec.ScheduleConfigs = newSlice
		newConfig, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Update(context.Background(), config, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("updating ScheduleConfig \"%s\" failed", name)
		}

		newSlice = append(newSlice, *scheduleConfig)
		newConfig.Spec.ScheduleConfigs = newSlice
		_, err = cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Update(context.Background(), newConfig, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		} else {
			return makeScheduleResponse(scheduleConfig, name), nil
		}
	} else {
		return nil, fmt.Errorf("ScheduleConfig \"%s\" is not created", name)
	}
}

func (cs *clusterSyncOperator) GetSchedule(name string) (*ScheduleResponse, error) {
	klog.Infof("Getting ScheduleConfig: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("OperatorConfig is not created")
		}
		return nil, err
	}

	if scheduleConfig := getScheduleConfig(config.Spec.ScheduleConfigs, name); scheduleConfig != nil {
		return makeScheduleResponse(scheduleConfig, scheduleConfig.ScheduleName), nil
	} else {
		return nil, fmt.Errorf("ScheduleConfig \"%s\" is not found", name)
	}
}

func (cs *clusterSyncOperator) ListSchedule() (*ListScheduleResponse, error) {
	klog.Infof("Listing ScheduleConfigs")
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("OperatorConfig is not created")
		}
		return nil, err
	}

	scheduleConfigs := config.Spec.ScheduleConfigs
	responseSlice := make([]ScheduleResponse, 0)
	for _, scheduleConfig := range scheduleConfigs {
		responseSlice = append(responseSlice, *makeScheduleResponse(&scheduleConfig, scheduleConfig.ScheduleName))
	}

	return &ListScheduleResponse{TotalCount: len(responseSlice), Items: responseSlice}, nil
}

func (cs *clusterSyncOperator) DeleteSchedule(name string) error {
	klog.Infof("Deleting ScheduleConfig: \"%s\"", name)
	config, err := cs.ksclient.ClustersyncV1().OperatorConfigs(OperatorConfigNamespace).Get(context.Background(), OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Infof("OperatorConfig is not created yet")
			return fmt.Errorf("OperatorConfig is not created")
		}
		klog.Infof("Error getting OperatorConfig")
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
		Template:     *makeBackupResponse(&config.ScheduleSpec.Template, ""),
		//{
		//	BackupName: "", // make this field be omitted
		//	IncludedNamespaces: config.ScheduleSpec.Template.IncludedNamespaces,
		//	ExcludedNamespaces: config.ScheduleSpec.Template.ExcludedNamespaces,
		//	TTL: config.ScheduleSpec.Template.TTL.String(),
		//	StorageLocation: config.ScheduleSpec.Template.StorageLocation,
		//	DefaultVolumesToFsBackup: *config.ScheduleSpec.Template.DefaultVolumesToFsBackup,
		//	VolumeSnapshotLocations: config.ScheduleSpec.Template.VolumeSnapshotLocations,
		//	SnapshotMoveData: *config.ScheduleSpec.Template.SnapshotMoveData,
		//},
	}

	return response
}