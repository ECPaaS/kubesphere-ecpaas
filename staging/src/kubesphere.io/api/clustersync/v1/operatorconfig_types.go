/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com.tw
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OperatorConfigSpec defines the desired state of OperatorConfig.
type OperatorConfigSpec struct {
	RepositoryConfigs []RepositoryConfig `json:"repositoryConfigs,omitempty"`
	BackupConfigs     []BackupConfig     `json:"backupConfigs,omitempty"`
	RestoreConfigs    []RestoreConfig    `json:"restoreConfigs,omitempty"`
	ScheduleConfigs   []ScheduleConfig   `json:"scheduleConfigs,omitempty"`
}

// RepositoryConfig defines the repository configurations.
type RepositoryConfig struct {
	RepositoryName string `json:"repositoryName,omitempty"`
	Provider       string `json:"provider,omitempty"`
	Bucket         string `json:"bucket,omitempty"`
	Prefix         string `json:"prefix,omitempty"`
	Region         string `json:"region,omitempty"`
	Ip             string `json:"ip,omitempty"`
	Port           *int   `json:"port,omitempty"`
	AccessKey      string `json:"accessKey,omitempty"`
	SecretKey      string `json:"secretKey,omitempty"`
	IsDefault      *bool  `json:"isDefault,omitempty"`
	LastModified   string `json:"lastModified,omitempty"`
}

// BackupConfig defines the backup configurations.
// Try to make BackupSpec as close as velerov1.Backup to simplify transition.
type BackupConfig struct {
	BackupName string     `json:"backupName,omitempty"`
	BackupSpec BackupSpec `json:"backupSpec,omitempty"`
	IsOneTime  *bool      `json:"isOneTime,omitempty"`
}

// Keep fields in BackupSpec identical to velerov1.BackupSpec.
// Add fields if more velero function is needed.
type BackupSpec struct {
	IncludedNamespaces       []string        `json:"includedNamespaces,omitempty"`
	ExcludedNamespaces       []string        `json:"excludedNamespaces,omitempty"`
	TTL                      metav1.Duration `json:"ttl,omitempty"`
	StorageLocation          string          `json:"storageLocation,omitempty"`
	DefaultVolumesToFsBackup *bool           `json:"defaultVolumesToFsBackup,omitempty"`
	VolumeSnapshotLocations  []string        `json:"volumeSnapshotLocations,omitempty"`
	SnapshotMoveData         *bool           `json:"snapshotMoveData,omitempty"`
}

// RestoreConfig defines the restore configurations.
// Try to make RestoreSpec as close as velerov1.Restore to simplify transition.
type RestoreConfig struct {
	RestoreName string      `json:"restoreName,omitempty"`
	RestoreSpec RestoreSpec `json:"restoreSpec,omitempty"`
	IsOneTime   *bool       `json:"isOneTime,omitempty"`
}

// Keep fields in RestoreSpec identical to velerov1.RestoreSpec.
// Add fields if more velero function is needed.
type RestoreSpec struct {
	BackupName         string   `json:"backupName,omitempty"`
	IncludedNamespaces []string `json:"includedNamespaces,omitempty"`
	ExcludedNamespaces []string `json:"excludedNamespaces,omitempty"`
}

// ScheduleConfig defines the schedule configurations.
// Try to make ScheduleSpec as close as velerov1.Schedule to simplify transition.
type ScheduleConfig struct {
	ScheduleName string       `json:"scheduleName,omitempty"`
	ScheduleSpec ScheduleSpec `json:"scheduleSpec,omitempty"`
	LastModified string       `json:"lastModified,omitempty"`
}

// Keep fields in ScheduleSpec identical to velerov1.ScheduleSpec.
// Add fields if more velero function is needed.
type ScheduleSpec struct {
	Schedule string     `json:"schedule"`
	Paused   bool       `json:"paused,omitempty"`
	Template BackupSpec `json:"template"`
}

// OperatorConfigStatus defines the observed state of OperatorConfig.
type OperatorConfigStatus struct {
	ConfigInitialized    bool               `json:"configInitialized,omitempty"`
	RepositoryConfigured []ModificationInfo `json:"repositoryConfigured,omitempty"`
	ScheduleConfigured   []ModificationInfo `json:"scheduleConfigured,omitempty"`
}

type ModificationInfo struct {
	Name         string `json:"name,omitempty"`
	LastModified string `json:"lastModified,omitempty"`
}

// +kubebuilder:subresource:status
// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OperatorConfig is the Schema for the operatorconfigs API.
type OperatorConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OperatorConfigSpec   `json:"spec,omitempty"`
	Status OperatorConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OperatorConfigList contains a list of OperatorConfig.
type OperatorConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OperatorConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OperatorConfig{}, &OperatorConfigList{})
}
