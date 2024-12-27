/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com
*/

package clustersync

// Storage
type StorageRequest struct {
	StorageName string `json:"storageName" description:"Storage name. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"32"`
	Provider    string `json:"provider,omitempty" default:"aws" description:"Storage provider name. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"32"`
	Bucket      string `json:"bucket,omitempty" description:"Storage bucket name. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"32"`
	Prefix      string `json:"prefix,omitempty" description:"Storage prefix name. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"32"`
	Region      string `json:"region,omitempty" description:"Storage region. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"32"`
	Ip          string `json:"ip,omitempty" description:"Storage IP."`
	Port        *int   `json:"port,omitempty" description:"Storage port." minimum:"1" maximum:"65535"`
	AccessKey   string `json:"accessKey,omitempty" description:"Storage access key." maximum:"32"`
	SecretKey   string `json:"secretKey,omitempty" description:"Storage secret key." maximum:"32"`
	IsDefault   *bool  `json:"isDefault,omitempty" default:"false" description:"Whether to set this storage as default."`
}

type StorageNameResponse struct {
	StorageName string `json:"storageName" description:"Storage name."`
}

type ModifyStorageRequest struct {
	Provider    string `json:"provider,omitempty" description:"Storage provider name. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"32"`
	Bucket      string `json:"bucket,omitempty" description:"Storage bucket name. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"32"`
	Prefix      string `json:"prefix,omitempty" description:"Storage prefix name. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"32"`
	Region      string `json:"region,omitempty" description:"Storage region. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"32"`
	Ip          string `json:"ip,omitempty" description:"Storage IP."`
	Port        *int   `json:"port,omitempty" description:"Storage port." minimum:"1" maximum:"65535"`
	AccessKey   string `json:"accessKey,omitempty" description:"Storage access key." maximum:"32"`
	SecretKey   string `json:"secretKey,omitempty" description:"Storage secret key." maximum:"32"`
	IsDefault   *bool  `json:"isDefault,omitempty" description:"Whether to set this storage as default."`
}

type StorageResponse struct {
	StorageName string `json:"storageName" description:"Storage name(unique key)."`
	Provider    string `json:"provider" description:"Storage provider name."`
	Bucket      string `json:"bucket" description:"Storage bucket name."`
	Prefix      string `json:"prefix" description:"Storage prefix name."`
	Region      string `json:"region" description:"Storage region."`
	Ip          string `json:"ip" description:"Storage IP."`
	Port        int    `json:"port" description:"Storage port."`
	AccessKey   string `json:"accessKey" description:"Storage access key."`
	SecretKey   string `json:"secretKey" description:"Storage secret key."`
	IsDefault   bool   `json:"isDefault" description:"Whether to set this storage as default."`
}

type ListStorageResponse struct {
	TotalCount int               `json:"total_count" description:"Total number of storages."`
	Items      []StorageResponse `json:"items" description:"List of storages."`
}

// Backup
type BackupRequest struct {
	BackupName               string   `json:"backupName" description:"Backup name. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"32"`
	IncludedNamespaces       []string `json:"includedNamespaces,omitempty" description:"IncludedNamespaces is a slice of namespace names to include objects from. If empty, all namespaces are included."`
	ExcludedNamespaces       []string `json:"excludedNamespaces,omitempty" description:"ExcludedNamespaces contains a list of namespaces that are not included in the backup."`
	TTL                      string   `json:"ttl,omitempty" default:"720h" description:"TTL is a time.Duration-parseable string describing how long the Backup should be retained for. Default is 720 hours(30 Days)." maximum:"32"`
	StorageLocation          string   `json:"storageLocation,omitempty" description:"StorageLocation is a string containing the name of a storage location where the backup should be stored. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"32"`
	DefaultVolumesToFsBackup *bool    `json:"defaultVolumesToFsBackup,omitempty" default:"false" description:"DefaultVolumesToFsBackup specifies whether pod volume file system backup should be used for all volumes by default."`
	VolumeSnapshotLocations  []string `json:"volumeSnapshotLocations,omitempty" description:"VolumeSnapshotLocations is a list containing names of VolumeSnapshotLocations associated with this backup. If SnapshotMoveData is enabled, at least one location shall be provided."`
	SnapshotMoveData         *bool    `json:"snapshotMoveData,omitempty" default:"false" description:"SnapshotMoveData specifies whether snapshot data should be moved."`
	IsOneTime                *bool    `json:"isOneTime,omitempty" description:"Whether this config is for one time backup or cluster sync."`
}

type BackupNameResponse struct {
	BackupName string `json:"backupName" description:"Backup name."`
}

type ModifyBackupRequest struct {
	IncludedNamespaces       []string `json:"includedNamespaces,omitempty" description:"IncludedNamespaces is a slice of namespace names to include objects from. If empty, all namespaces are included."`
	ExcludedNamespaces       []string `json:"excludedNamespaces,omitempty" description:"ExcludedNamespaces contains a list of namespaces that are not included in the backup."`
	TTL                      string   `json:"ttl,omitempty" description:"TTL is a time.Duration-parseable string describing how long the Backup should be retained for." maximum:"32"`
	StorageLocation          string   `json:"storageLocation,omitempty" description:"StorageLocation is a string containing the name of a storage location where the backup should be stored. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"32"`
	DefaultVolumesToFsBackup *bool    `json:"defaultVolumesToFsBackup,omitempty" description:"DefaultVolumesToFsBackup specifies whether pod volume file system backup should be used for all volumes by default."`
	VolumeSnapshotLocations  []string `json:"volumeSnapshotLocations,omitempty" description:"VolumeSnapshotLocations is a list containing names of VolumeSnapshotLocations associated with this backup. If SnapshotMoveData is enabled, at least one location shall be provided."`
	SnapshotMoveData         *bool    `json:"snapshotMoveData,omitempty" description:"SnapshotMoveData specifies whether snapshot data should be moved."`
}

type BackupResponse struct {
	BackupName               string   `json:"backupName,omitempty" description:"Backup name(unique key)."`
	IncludedNamespaces       []string `json:"includedNamespaces" description:"IncludedNamespaces is a slice of namespace names to include objects from. If empty, all namespaces are included."`
	ExcludedNamespaces       []string `json:"excludedNamespaces" description:"ExcludedNamespaces contains a list of namespaces that are not included in the backup."`
	TTL                      string   `json:"ttl" description:"TTL is a time.Duration-parseable string describing how long the Backup should be retained for."`
	StorageLocation          string   `json:"storageLocation" description:"StorageLocation is a string containing the name of a storage location where the backup should be stored."`
	DefaultVolumesToFsBackup bool     `json:"defaultVolumesToFsBackup" description:"DefaultVolumesToFsBackup specifies whether pod volume file system backup should be used for all volumes by default."`
	VolumeSnapshotLocations  []string `json:"volumeSnapshotLocations" description:"VolumeSnapshotLocations is a list containing names of VolumeSnapshotLocations associated with this backup."`
	SnapshotMoveData         bool     `json:"snapshotMoveData" description:"SnapshotMoveData specifies whether snapshot data should be moved."`
}

type ListBackupResponse struct {
	TotalCount int              `json:"total_count" description:"Total number of backups."`
	Items      []BackupResponse `json:"items" description:"List of backups."`
}

// Restore
type RestoreRequest struct {
	RestoreName              string   `json:"restoreName" description:"Restore name. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"32"`
	BackupName               string   `json:"backupName" description:"BackupName is the unique name of the Velero backup to restore from. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"32"`
	IncludedNamespaces       []string `json:"includedNamespaces,omitempty" description:"IncludedNamespaces is a slice of namespace names to include objects from. If empty, all namespaces are included."`
	ExcludedNamespaces       []string `json:"excludedNamespaces,omitempty" description:"ExcludedNamespaces contains a list of namespaces that are not included in the backup."`
	IsOneTime                *bool    `json:"isOneTime,omitempty" description:"Whether this config is for one time restore or cluster sync."`
}

type RestoreNameResponse struct {
	RestoreName string `json:"restoreName" description:"Restore name."`
}

type ModifyRestoreRequest struct {
	BackupName               string   `json:"backupName" description:"BackupName is the unique name of the Velero backup to restore from. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"32"`
	IncludedNamespaces       []string `json:"includedNamespaces,omitempty" description:"IncludedNamespaces is a slice of namespace names to include objects from. If empty, all namespaces are included."`
	ExcludedNamespaces       []string `json:"excludedNamespaces,omitempty" description:"ExcludedNamespaces contains a list of namespaces that are not included in the backup."`
}

type RestoreResponse struct {
	RestoreName              string   `json:"restoreName" description:"Restore name(unique key)."`
	BackupName               string   `json:"backupName" description:"BackupName is the unique name of the Velero backup to restore from."`
	IncludedNamespaces       []string `json:"includedNamespaces" description:"IncludedNamespaces is a slice of namespace names to include objects from. If empty, all namespaces are included."`
	ExcludedNamespaces       []string `json:"excludedNamespaces" description:"ExcludedNamespaces contains a list of namespaces that are not included in the backup."`
}

type ListRestoreResponse struct {
	TotalCount int               `json:"total_count" description:"Total number of Restores."`
	Items      []RestoreResponse `json:"items" description:"List of Restores."`
}

// Schedule
type ScheduleRequest struct {
	ScheduleName string        `json:"scheduleName" description:"Schedule name. Valid characters: A-Z, a-z, 0-9, and -(hyphen)." maximum:"32"`
	Schedule     string        `json:"schedule,omitempty" description:"Schedule is a Cron expression defining when to run."`
	Paused       *bool         `json:"paused,omitempty" default:"false" description:"Paused specifies whether the schedule is paused or not."`
	Template     BackupRequest `json:"template,omitempty" description:"Template is the definition of the Backup to be run on the provided schedule."`
}

type ScheduleNameResponse struct {
	ScheduleName string `json:"scheduleName" description:"Schedule name."`
}

type ModifyScheduleRequest struct {
	Schedule     string         `json:"schedule,omitempty" description:"Schedule is a Cron expression defining when to run."`
	Paused       *bool          `json:"paused,omitempty" description:"Paused specifies whether the schedule is paused or not."`
	Template     *BackupRequest `json:"template,omitempty" description:"Template is the definition of the Backup to be run on the provided schedule."`
}

type ScheduleResponse struct {
	ScheduleName string         `json:"scheduleName" description:"Schedule name(unique key)."`
	Schedule     string         `json:"schedule" description:"Schedule is a Cron expression defining when to run."`
	Paused       bool           `json:"paused" description:"Paused specifies whether the schedule is paused or not."`
	Template     BackupResponse `json:"template" description:"Template is the definition of the Backup to be run on the provided schedule."`
}

type ListScheduleResponse struct {
	TotalCount int                `json:"total_count" description:"Total number of schedules."`
	Items      []ScheduleResponse `json:"items" description:"List of schedules."`
}