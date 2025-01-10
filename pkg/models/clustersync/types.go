/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com
*/

package clustersync

// Repository
type RepositoryRequest struct {
	RepositoryName string `json:"repositoryName" description:"Repository name. Must be unique. Valid characters: A-Z, a-z, 0-9, and -(hyphen). And must start and end with alphanumeric character." maximum:"32"`
	Provider       string `json:"provider" description:"Repository provider name. Valid characters: A-Z, a-z, 0-9, and -(hyphen). And must start and end with alphanumeric character." maximum:"32"`
	Bucket         string `json:"bucket" description:"Repository bucket name. Valid characters: A-Z, a-z, 0-9, and -(hyphen). And must start and end with alphanumeric character." maximum:"32"`
	Prefix         string `json:"prefix" description:"Repository prefix name. Valid characters: A-Z, a-z, 0-9, and -(hyphen). And must start and end with alphanumeric character." maximum:"32"`
	Region         string `json:"region" description:"Repository region. Valid characters: A-Z, a-z, 0-9, and -(hyphen). And must start and end with alphanumeric character." maximum:"32"`
	Ip             string `json:"ip" description:"Repository IP. Must be valid IPv4/v6 format without mask."`
	Port           *int   `json:"port" description:"Repository port." minimum:"1" maximum:"65535"`
	AccessKey      string `json:"accessKey" description:"Repository access key." maximum:"128"`
	SecretKey      string `json:"secretKey" description:"Repository secret key." maximum:"128"`
	IsDefault      *bool  `json:"isDefault,omitempty" default:"false" description:"Whether to set this repository as default."`
}

type RepositoryNameResponse struct {
	RepositoryName string `json:"repositoryName" description:"Repository name."`
}

type ModifyRepositoryRequest struct {
	Provider  string `json:"provider,omitempty" description:"Repository provider name. Valid characters: A-Z, a-z, 0-9, and -(hyphen). And must start and end with alphanumeric character." maximum:"32"`
	Bucket    string `json:"bucket,omitempty" description:"Repository bucket name. Valid characters: A-Z, a-z, 0-9, and -(hyphen). And must start and end with alphanumeric character." maximum:"32"`
	Prefix    string `json:"prefix,omitempty" description:"Repository prefix name. Valid characters: A-Z, a-z, 0-9, and -(hyphen). And must start and end with alphanumeric character." maximum:"32"`
	Region    string `json:"region,omitempty" description:"Repository region. Valid characters: A-Z, a-z, 0-9, and -(hyphen). And must start and end with alphanumeric character." maximum:"32"`
	Ip        string `json:"ip,omitempty" description:"Repository IP. Must be valid IPv4/v6 format without mask."`
	Port      *int   `json:"port,omitempty" description:"Repository port." minimum:"1" maximum:"65535"`
	AccessKey string `json:"accessKey,omitempty" description:"Repository access key." maximum:"128"`
	SecretKey string `json:"secretKey,omitempty" description:"Repository secret key." maximum:"128"`
	IsDefault *bool  `json:"isDefault,omitempty" description:"Whether to set this repository as default."`
}

type RepositoryResponse struct {
	RepositoryName string `json:"repositoryName" description:"Repository name(unique key)."`
	Provider       string `json:"provider" description:"Repository provider name."`
	Bucket         string `json:"bucket" description:"Repository bucket name."`
	Prefix         string `json:"prefix" description:"Repository prefix name."`
	Region         string `json:"region" description:"Repository region."`
	Ip             string `json:"ip" description:"Repository IP."`
	Port           int    `json:"port" description:"Repository port."`
	AccessKey      string `json:"accessKey" description:"Repository access key. Base64 encoded(not encrypted)."`
	SecretKey      string `json:"secretKey" description:"Repository secret key. Base64 encoded(not encrypted)."`
	IsDefault      bool   `json:"isDefault" description:"Whether to set this repository as default."`
}

type ListRepositoryResponse struct {
	TotalCount int                  `json:"total_count" description:"Total number of repositories."`
	Items      []RepositoryResponse `json:"items" description:"List of repositories. Key is items[].repositoryName"`
}

// Backup
type BackupRequest struct {
	BackupName               string   `json:"backupName" description:"Backup name. Must be unique. Valid characters: A-Z, a-z, 0-9, and -(hyphen). And must start and end with alphanumeric character." maximum:"32"`
	IncludedNamespaces       []string `json:"includedNamespaces,omitempty" description:"IncludedNamespaces is a slice of namespace names to include objects from. If empty, all namespaces are included. Valid array data characters: a-z, 0-9, -(hyphen). Array data must start and end with alphanumeric character."`
	ExcludedNamespaces       []string `json:"excludedNamespaces,omitempty" description:"ExcludedNamespaces contains a list of namespaces that are not included in the backup. If empty, no namespace is excluded. Valid array data characters: a-z, 0-9, -(hyphen). Array data must start and end with alphanumeric character."`
	TTL                      string   `json:"ttl,omitempty" default:"720h" description:"TTL is a time.Duration-parseable string describing how long the Backup should be retained for. Default is 720 hours(30 Days). A time.Duration-parseable string is a signed sequence of decimal numbers with optional fraction and unit suffix." maximum:"32"`
	BackupRepository         string   `json:"backupRepository,omitempty" default:"" description:"BackupRepository is a string containing the name of a repository where the backup should be stored. Empty string means use the default repository. If no default repository exists, this field is required. Valid characters: A-Z, a-z, 0-9, and -(hyphen). And must start and end with alphanumeric character." maximum:"32"`
	DefaultVolumesToFsBackup *bool    `json:"defaultVolumesToFsBackup,omitempty" default:"false" description:"DefaultVolumesToFsBackup specifies whether pod volume file system backup should be used for all volumes by default."`
	SnapshotRepositories     []string `json:"snapshotRepositories,omitempty" description:"SnapshotRepositories is a list containing names of repositories for volume snapshots associated with this backup. Empty array means use the default repository. If no default repository exists, this field is required. Valid array data characters: A-Z, a-z, 0-9, and -(hyphen). Array data must start and end with alphanumeric character."`
	SnapshotMoveData         *bool    `json:"snapshotMoveData,omitempty" default:"false" description:"SnapshotMoveData specifies whether snapshot data should be moved."`
	IsOneTime                *bool    `json:"isOneTime" description:"Whether this config is for one time backup or cluster sync."`
}

type BackupNameResponse struct {
	BackupName string `json:"backupName" description:"Backup name."`
}

type ModifyBackupRequest struct {
	IncludedNamespaces       []string `json:"includedNamespaces,omitempty" description:"IncludedNamespaces is a slice of namespace names to include objects from. Empty array means clear this array to include all namespaces. Valid array data characters: a-z, 0-9, -(hyphen). Array data must start and end with alphanumeric character."`
	ExcludedNamespaces       []string `json:"excludedNamespaces,omitempty" description:"ExcludedNamespaces contains a list of namespaces that are not included in the backup. Empty array means clear this array to exclude no namespace. Valid array data characters: a-z, 0-9, -(hyphen). Array data must start and end with alphanumeric character."`
	TTL                      *string  `json:"ttl,omitempty" description:"TTL is a time.Duration-parseable string describing how long the Backup should be retained for. A time.Duration-parseable string is a signed sequence of decimal numbers with optional fraction and unit suffix. Empty string means clear this field to use the default value(720 hours)." maximum:"32"`
	BackupRepository         *string  `json:"backupRepository,omitempty" description:"BackupRepository is a string containing the name of a repository where the backup should be stored. Empty string means use the default repository. Empty string means clear this field to use the default repository. Valid characters: A-Z, a-z, 0-9, and -(hyphen). And must start and end with alphanumeric character." maximum:"32"`
	DefaultVolumesToFsBackup *bool    `json:"defaultVolumesToFsBackup,omitempty" description:"DefaultVolumesToFsBackup specifies whether pod volume file system backup should be used for all volumes by default."`
	SnapshotRepositories     []string `json:"snapshotRepositories,omitempty" description:"SnapshotRepositories is a list containing names of repositories for volume snapshots associated with this backup. Empty array means clear this array to use the default repository. Valid array data characters: A-Z, a-z, 0-9, and -(hyphen). Array data must start and end with alphanumeric character."`
	SnapshotMoveData         *bool    `json:"snapshotMoveData,omitempty" description:"SnapshotMoveData specifies whether snapshot data should be moved."`
}

type BackupResponse struct {
	BackupName               string   `json:"backupName" description:"Backup name(unique key)."`
	IncludedNamespaces       []string `json:"includedNamespaces" description:"IncludedNamespaces is a slice of namespace names to include objects from. If empty, all namespaces are included."`
	ExcludedNamespaces       []string `json:"excludedNamespaces" description:"ExcludedNamespaces contains a list of namespaces that are not included in the backup. If empty, no namespace is excluded."`
	TTL                      string   `json:"ttl" description:"TTL is a time.Duration-parseable string describing how long the Backup should be retained for."`
	BackupRepository         string   `json:"backupRepository" description:"BackupRepository is a string containing the name of a repository where the backup should be stored. Empty string means use the default repository."`
	DefaultVolumesToFsBackup bool     `json:"defaultVolumesToFsBackup" description:"DefaultVolumesToFsBackup specifies whether pod volume file system backup should be used for all volumes by default."`
	SnapshotRepositories     []string `json:"snapshotRepositories" description:"SnapshotRepositories is a list containing names of repositories for volume snapshots associated with this backup. Empty array means use the default repository."`
	SnapshotMoveData         bool     `json:"snapshotMoveData" description:"SnapshotMoveData specifies whether snapshot data should be moved."`
}

type ListBackupResponse struct {
	TotalCount int              `json:"total_count" description:"Total number of backups."`
	Items      []BackupResponse `json:"items" description:"List of backups. Key is items[].backupName"`
}

// Restore
type RestoreRequest struct {
	RestoreName        string   `json:"restoreName" description:"Restore name. Must be unique. Valid characters: A-Z, a-z, 0-9, and -(hyphen). And must start and end with alphanumeric character." maximum:"32"`
	BackupSource       string   `json:"backupSource" description:"BackupSource is the unique name of the backup source to restore from. Valid characters: A-Z, a-z, 0-9, and -(hyphen). And must start and end with alphanumeric character." maximum:"32"`
	IncludedNamespaces []string `json:"includedNamespaces,omitempty" description:"IncludedNamespaces is a slice of namespace names to include objects from. If empty, all namespaces are included. Valid array data characters: a-z, 0-9, -(hyphen). Array data must start and end with alphanumeric character."`
	ExcludedNamespaces []string `json:"excludedNamespaces,omitempty" description:"ExcludedNamespaces contains a list of namespaces that are not included in the backup. Valid array data characters: a-z, 0-9, -(hyphen). Array data must start and end with alphanumeric character."`
	IsOneTime          *bool    `json:"isOneTime" description:"Whether this config is for one time restore or cluster sync."`
}

type RestoreNameResponse struct {
	RestoreName string `json:"restoreName" description:"Restore name."`
}

type ModifyRestoreRequest struct {
	BackupSource       string   `json:"backupSource,omitempty" description:"BackupSource is the unique name of the backup source to restore from. Valid characters: A-Z, a-z, 0-9, and -(hyphen). And must start and end with alphanumeric character." maximum:"32"`
	IncludedNamespaces []string `json:"includedNamespaces,omitempty" description:"IncludedNamespaces is a slice of namespace names to include objects from. If empty, all namespaces are included. Empty array means clear this array to include all namespaces. Valid array data characters: a-z, 0-9, -(hyphen). Array data must start and end with alphanumeric character."`
	ExcludedNamespaces []string `json:"excludedNamespaces,omitempty" description:"ExcludedNamespaces contains a list of namespaces that are not included in the backup. Empty array means clear this array to exclude no namespace. Valid array data characters: a-z, 0-9, -(hyphen). Array data must start and end with alphanumeric character."`
}

type RestoreResponse struct {
	RestoreName        string   `json:"restoreName" description:"Restore name(unique key)."`
	BackupSource       string   `json:"backupSource" description:"BackupSource is the unique name of the backup source to restore from."`
	IncludedNamespaces []string `json:"includedNamespaces" description:"IncludedNamespaces is a slice of namespace names to include objects from. If empty, all namespaces are included."`
	ExcludedNamespaces []string `json:"excludedNamespaces" description:"ExcludedNamespaces contains a list of namespaces that are not included in the backup. If empty, no namespace is excluded."`
}

type ListRestoreResponse struct {
	TotalCount int               `json:"total_count" description:"Total number of restores."`
	Items      []RestoreResponse `json:"items" description:"List of restores. Key is items[].restoreName"`
}

// Schedule
type ScheduleRequest struct {
	ScheduleName string       `json:"scheduleName" description:"Schedule name. Must be unique. Valid characters: A-Z, a-z, 0-9, and -(hyphen). And must start and end with alphanumeric character." maximum:"32"`
	Schedule     string       `json:"schedule" description:"Schedule is a Cron expression defining when to run. Valid characters: 0-9, /(slash), *(asterisk), space, and -(hyphen)."`
	Paused       *bool        `json:"paused,omitempty" default:"false" description:"Paused specifies whether the schedule is paused or not."`
	Template     PostTemplate `json:"template,omitempty" description:"Template is the definition of the Backup to be run on the provided schedule."`
}

type PostTemplate struct {
	IncludedNamespaces       []string `json:"includedNamespaces,omitempty" description:"IncludedNamespaces is a slice of namespace names to include objects from. If empty, all namespaces are included. Valid array data characters: a-z, 0-9, -(hyphen). Array data must start and end with alphanumeric character."`
	ExcludedNamespaces       []string `json:"excludedNamespaces,omitempty" description:"ExcludedNamespaces contains a list of namespaces that are not included in the backup. If empty, no namespace is excluded. Valid array data characters: a-z, 0-9, -(hyphen). Array data must start and end with alphanumeric character."`
	TTL                      string   `json:"ttl,omitempty" default:"720h" description:"TTL is a time.Duration-parseable string describing how long the Backup should be retained for. Default is 720 hours(30 Days). A time.Duration-parseable string is a signed sequence of decimal numbers with optional fraction and unit suffix." maximum:"32"`
	BackupRepository         string   `json:"backupRepository,omitempty" default:"" description:"BackupRepository is a string containing the name of a repository where the backup should be stored. Empty string means use the default repository. If no default repository exists, this field is required. Valid characters: A-Z, a-z, 0-9, and -(hyphen). And must start and end with alphanumeric character." maximum:"32"`
	DefaultVolumesToFsBackup *bool    `json:"defaultVolumesToFsBackup,omitempty" default:"false" description:"DefaultVolumesToFsBackup specifies whether pod volume file system backup should be used for all volumes by default."`
	SnapshotRepositories     []string `json:"snapshotRepositories,omitempty" description:"SnapshotRepositories is a list containing names of repositories for volume snapshots associated with this backup. Empty array means use the default repository. If no default repository exists, this field is required. Valid array data characters: A-Z, a-z, 0-9, and -(hyphen). Array data must start and end with alphanumeric character."`
	SnapshotMoveData         *bool    `json:"snapshotMoveData,omitempty" default:"false" description:"SnapshotMoveData specifies whether snapshot data should be moved."`
}

type ScheduleNameResponse struct {
	ScheduleName string `json:"scheduleName" description:"Schedule name."`
}

type ModifyScheduleRequest struct {
	Schedule string       `json:"schedule,omitempty" description:"Schedule is a Cron expression defining when to run. Valid characters: 0-9, /(slash), *(asterisk), space, and -(hyphen)."`
	Paused   *bool        `json:"paused,omitempty" description:"Paused specifies whether the schedule is paused or not."`
	Template *PutTemplate `json:"template,omitempty" description:"Template is the definition of the Backup to be run on the provided schedule."`
}

type PutTemplate struct {
	IncludedNamespaces       []string `json:"includedNamespaces,omitempty" description:"IncludedNamespaces is a slice of namespace names to include objects from. Empty array means clear this array to include all namespaces. Valid array data characters: a-z, 0-9, -(hyphen). Array data must start and end with alphanumeric character."`
	ExcludedNamespaces       []string `json:"excludedNamespaces,omitempty" description:"ExcludedNamespaces contains a list of namespaces that are not included in the backup. Empty array means clear this array to exclude no namespace. Valid array data characters: a-z, 0-9, -(hyphen). Array data must start and end with alphanumeric character."`
	TTL                      *string  `json:"ttl,omitempty" description:"TTL is a time.Duration-parseable string describing how long the Backup should be retained for. A time.Duration-parseable string is a signed sequence of decimal numbers with optional fraction and unit suffix. Empty string means clear this field to use the default value(720 hours)." maximum:"32"`
	BackupRepository         *string  `json:"backupRepository,omitempty" description:"BackupRepository is a string containing the name of a repository where the backup should be stored. Empty string means use the default repository. Empty string means clear this field to use the default repository. Valid characters: A-Z, a-z, 0-9, and -(hyphen). And must start and end with alphanumeric character." maximum:"32"`
	DefaultVolumesToFsBackup *bool    `json:"defaultVolumesToFsBackup,omitempty" description:"DefaultVolumesToFsBackup specifies whether pod volume file system backup should be used for all volumes by default."`
	SnapshotRepositories     []string `json:"snapshotRepositories,omitempty" description:"SnapshotRepositories is a list containing names of repositories for volume snapshots associated with this backup. Empty array means clear this array to use the default repository. Valid array data characters: A-Z, a-z, 0-9, and -(hyphen). Array data must start and end with alphanumeric character."`
	SnapshotMoveData         *bool    `json:"snapshotMoveData,omitempty" description:"SnapshotMoveData specifies whether snapshot data should be moved."`
}

type ScheduleResponse struct {
	ScheduleName string           `json:"scheduleName" description:"Schedule name(unique key)."`
	Schedule     string           `json:"schedule" description:"Schedule is a Cron expression defining when to run."`
	Paused       bool             `json:"paused" description:"Paused specifies whether the schedule is paused or not."`
	Template     TemplateResponse `json:"template" description:"Template is the definition of the Backup to be run on the provided schedule."`
}

type TemplateResponse struct {
	IncludedNamespaces       []string `json:"includedNamespaces" description:"IncludedNamespaces is a slice of namespace names to include objects from. If empty, all namespaces are included."`
	ExcludedNamespaces       []string `json:"excludedNamespaces" description:"ExcludedNamespaces contains a list of namespaces that are not included in the backup. If empty, no namespace is excluded."`
	TTL                      string   `json:"ttl" description:"TTL is a time.Duration-parseable string describing how long the Backup should be retained for."`
	BackupRepository         string   `json:"backupRepository" description:"BackupRepository is a string containing the name of a repository where the backup should be stored. Empty string means use the default repository."`
	DefaultVolumesToFsBackup bool     `json:"defaultVolumesToFsBackup" description:"DefaultVolumesToFsBackup specifies whether pod volume file system backup should be used for all volumes by default."`
	SnapshotRepositories     []string `json:"snapshotRepositories" description:"SnapshotRepositories is a list containing names of repositories for volume snapshots associated with this backup. Empty array means use the default repository."`
	SnapshotMoveData         bool     `json:"snapshotMoveData" description:"SnapshotMoveData specifies whether snapshot data should be moved."`
}

type ListScheduleResponse struct {
	TotalCount int                `json:"total_count" description:"Total number of schedules."`
	Items      []ScheduleResponse `json:"items" description:"List of schedules. Key is items[].scheduleName"`
}