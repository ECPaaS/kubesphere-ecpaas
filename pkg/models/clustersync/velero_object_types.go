/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com
*/

package clustersync

import (
	corev1api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Unmarshal objects copied from Velero v1.15.0.
// Don't want these objects to become real CRDs.
// Because these objects are only used for APIs to fetch status from K8s.

// Repository

// Unmarshal object for velero.io.BackupStorageLocations
type RepositoryObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupStorageLocationSpec   `json:"spec,omitempty"`
	Status BackupStorageLocationStatus `json:"status,omitempty"`
}

// BackupStorageLocationSpec defines the desired state of a Velero BackupStorageLocation
type BackupStorageLocationSpec struct {
	// Provider is the provider of the backup storage.
	Provider string `json:"provider"`

	// Config is for provider-specific configuration fields.
	Config map[string]string `json:"config,omitempty"`

	// Credential contains the credential information intended to be used with this location
	Credential *corev1api.SecretKeySelector `json:"credential,omitempty"`

	StorageType `json:",inline"`

	// Default indicates this location is the default backup storage location.
	Default bool `json:"default,omitempty"`

	// AccessMode defines the permissions for the backup storage location.
	AccessMode BackupStorageLocationAccessMode `json:"accessMode,omitempty"`

	// BackupSyncPeriod defines how frequently to sync backup API objects from object storage. A value of 0 disables sync.
	BackupSyncPeriod *metav1.Duration `json:"backupSyncPeriod,omitempty"`

	// ValidationFrequency defines how frequently to validate the corresponding object storage. A value of 0 disables validation.
	ValidationFrequency *metav1.Duration `json:"validationFrequency,omitempty"`
}

// StorageType represents the type of storage that a backup location uses.
// ObjectStorage must be non-nil, since it is currently the only supported StorageType.
type StorageType struct {
	ObjectStorage *ObjectStorageLocation `json:"objectStorage"`
}

// ObjectStorageLocation specifies the settings necessary to connect to a provider's object storage.
type ObjectStorageLocation struct {
	// Bucket is the bucket to use for object storage.
	Bucket string `json:"bucket"`

	// Prefix is the path inside a bucket to use for Velero storage. Optional.
	Prefix string `json:"prefix,omitempty"`

	// CACert defines a CA bundle to use when verifying TLS connections to the provider.
	CACert []byte `json:"caCert,omitempty"`
}

// BackupStorageLocationAccessMode represents the permissions for a BackupStorageLocation.
type BackupStorageLocationAccessMode string

// BackupStorageLocationStatus defines the observed state of BackupStorageLocation
type BackupStorageLocationStatus struct {
	// Phase is the current state of the BackupStorageLocation.
	Phase BackupStorageLocationPhase `json:"phase,omitempty"`

	// LastSyncedTime is the last time the contents of the location were synced into
	// the cluster.
	LastSyncedTime *metav1.Time `json:"lastSyncedTime,omitempty"`

	// LastValidationTime is the last time the backup store location was validated
	// the cluster.
	LastValidationTime *metav1.Time `json:"lastValidationTime,omitempty"`

	// Message is a message about the backup storage location's status.
	Message string `json:"message,omitempty"`

	// LastSyncedRevision is the value of the `metadata/revision` file in the backup
	// storage location the last time the BSL's contents were synced into the cluster.
	//
	// Deprecated: this field is no longer updated or used for detecting changes to
	// the location's contents and will be removed entirely in v2.0.
	LastSyncedRevision types.UID `json:"lastSyncedRevision,omitempty"`

	// AccessMode is an unused field.
	//
	// Deprecated: there is now an AccessMode field on the Spec and this field
	// will be removed entirely as of v2.0.
	AccessMode BackupStorageLocationAccessMode `json:"accessMode,omitempty"`
}

// BackupStorageLocationPhase is the lifecycle phase of a Velero BackupStorageLocation.
type BackupStorageLocationPhase string


// Backup

// Unmarshal object for velero.io.Backups
type BackupObject struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec BackupSpec `json:"spec,omitempty"`
	Status BackupStatus `json:"status,omitempty"`
}

// BackupSpec defines the specification for a Velero backup.
type BackupSpec struct {
	Metadata `json:"metadata,omitempty"`
	// IncludedNamespaces is a slice of namespace names to include objects
	// from. If empty, all namespaces are included.
	IncludedNamespaces []string `json:"includedNamespaces,omitempty"`

	// ExcludedNamespaces contains a list of namespaces that are not
	// included in the backup.
	ExcludedNamespaces []string `json:"excludedNamespaces,omitempty"`

	// IncludedResources is a slice of resource names to include
	// in the backup. If empty, all resources are included.
	IncludedResources []string `json:"includedResources,omitempty"`

	// ExcludedResources is a slice of resource names that are not
	// included in the backup.
	ExcludedResources []string `json:"excludedResources,omitempty"`

	// IncludedClusterScopedResources is a slice of cluster-scoped
	// resource type names to include in the backup.
	// If set to "*", all cluster-scoped resource types are included.
	// The default value is empty, which means only related
	// cluster-scoped resources are included.
	IncludedClusterScopedResources []string `json:"includedClusterScopedResources,omitempty"`

	// ExcludedClusterScopedResources is a slice of cluster-scoped
	// resource type names to exclude from the backup.
	// If set to "*", all cluster-scoped resource types are excluded.
	// The default value is empty.
	ExcludedClusterScopedResources []string `json:"excludedClusterScopedResources,omitempty"`

	// IncludedNamespaceScopedResources is a slice of namespace-scoped
	// resource type names to include in the backup.
	// The default value is "*".
	IncludedNamespaceScopedResources []string `json:"includedNamespaceScopedResources,omitempty"`

	// ExcludedNamespaceScopedResources is a slice of namespace-scoped
	// resource type names to exclude from the backup.
	// If set to "*", all namespace-scoped resource types are excluded.
	// The default value is empty.
	ExcludedNamespaceScopedResources []string `json:"excludedNamespaceScopedResources,omitempty"`

	// LabelSelector is a metav1.LabelSelector to filter with
	// when adding individual objects to the backup. If empty
	// or nil, all objects are included. Optional.
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`

	// OrLabelSelectors is list of metav1.LabelSelector to filter with
	// when adding individual objects to the backup. If multiple provided
	// they will be joined by the OR operator. LabelSelector as well as
	// OrLabelSelectors cannot co-exist in backup request, only one of them
	// can be used.
	OrLabelSelectors []*metav1.LabelSelector `json:"orLabelSelectors,omitempty"`

	// SnapshotVolumes specifies whether to take snapshots
	// of any PV's referenced in the set of objects included
	// in the Backup.
	SnapshotVolumes *bool `json:"snapshotVolumes,omitempty"`

	// TTL is a time.Duration-parseable string describing how long
	// the Backup should be retained for.
	TTL metav1.Duration `json:"ttl,omitempty"`

	// IncludeClusterResources specifies whether cluster-scoped resources
	// should be included for consideration in the backup.
	IncludeClusterResources *bool `json:"includeClusterResources,omitempty"`

	// Hooks represent custom behaviors that should be executed at different phases of the backup.
	Hooks BackupHooks `json:"hooks,omitempty"`

	// StorageLocation is a string containing the name of a BackupStorageLocation where the backup should be stored.
	StorageLocation string `json:"storageLocation,omitempty"`

	// VolumeSnapshotLocations is a list containing names of VolumeSnapshotLocations associated with this backup.
	VolumeSnapshotLocations []string `json:"volumeSnapshotLocations,omitempty"`

	// DefaultVolumesToRestic specifies whether restic should be used to take a
	// backup of all pod volumes by default.
	//
	// Deprecated: this field is no longer used and will be removed entirely in future. Use DefaultVolumesToFsBackup instead.
	DefaultVolumesToRestic *bool `json:"defaultVolumesToRestic,omitempty"`

	// DefaultVolumesToFsBackup specifies whether pod volume file system backup should be used
	// for all volumes by default.
	DefaultVolumesToFsBackup *bool `json:"defaultVolumesToFsBackup,omitempty"`

	// OrderedResources specifies the backup order of resources of specific Kind.
	// The map key is the resource name and value is a list of object names separated by commas.
	// Each resource name has format "namespace/objectname".  For cluster resources, simply use "objectname".
	OrderedResources map[string]string `json:"orderedResources,omitempty"`

	// CSISnapshotTimeout specifies the time used to wait for CSI VolumeSnapshot status turns to
	// ReadyToUse during creation, before returning error as timeout.
	// The default value is 10 minute.
	CSISnapshotTimeout metav1.Duration `json:"csiSnapshotTimeout,omitempty"`

	// ItemOperationTimeout specifies the time used to wait for asynchronous BackupItemAction operations
	// The default value is 4 hour.
	ItemOperationTimeout metav1.Duration `json:"itemOperationTimeout,omitempty"`
	// ResourcePolicy specifies the referenced resource policies that backup should follow
	ResourcePolicy *corev1api.TypedLocalObjectReference `json:"resourcePolicy,omitempty"`

	// SnapshotMoveData specifies whether snapshot data should be moved
	SnapshotMoveData *bool `json:"snapshotMoveData,omitempty"`

	// DataMover specifies the data mover to be used by the backup.
	// If DataMover is "" or "velero", the built-in data mover will be used.
	DataMover string `json:"datamover,omitempty"`

	// UploaderConfig specifies the configuration for the uploader.
	UploaderConfig *UploaderConfigForBackup `json:"uploaderConfig,omitempty"`
}

type Metadata struct {
	Labels map[string]string `json:"labels,omitempty"`
}

// BackupHooks contains custom behaviors that should be executed at different phases of the backup.
type BackupHooks struct {
	// Resources are hooks that should be executed when backing up individual instances of a resource.
	Resources []BackupResourceHookSpec `json:"resources,omitempty"`
}

// BackupResourceHookSpec defines one or more BackupResourceHooks that should be executed based on
// the rules defined for namespaces, resources, and label selector.
type BackupResourceHookSpec struct {
	// Name is the name of this hook.
	Name string `json:"name"`

	// IncludedNamespaces specifies the namespaces to which this hook spec applies. If empty, it applies
	// to all namespaces.
	IncludedNamespaces []string `json:"includedNamespaces,omitempty"`

	// ExcludedNamespaces specifies the namespaces to which this hook spec does not apply.
	ExcludedNamespaces []string `json:"excludedNamespaces,omitempty"`

	// IncludedResources specifies the resources to which this hook spec applies. If empty, it applies
	// to all resources.
	IncludedResources []string `json:"includedResources,omitempty"`

	// ExcludedResources specifies the resources to which this hook spec does not apply.
	ExcludedResources []string `json:"excludedResources,omitempty"`

	// LabelSelector, if specified, filters the resources to which this hook spec applies.
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`

	// PreHooks is a list of BackupResourceHooks to execute prior to storing the item in the backup.
	// These are executed before any "additional items" from item actions are processed.
	PreHooks []BackupResourceHook `json:"pre,omitempty"`

	// PostHooks is a list of BackupResourceHooks to execute after storing the item in the backup.
	// These are executed after all "additional items" from item actions are processed.
	PostHooks []BackupResourceHook `json:"post,omitempty"`
}

// BackupResourceHook defines a hook for a resource.
type BackupResourceHook struct {
	// Exec defines an exec hook.
	Exec *ExecHook `json:"exec"`
}

// ExecHook is a hook that uses the pod exec API to execute a command in a container in a pod.
type ExecHook struct {
	// Container is the container in the pod where the command should be executed. If not specified,
	// the pod's first container is used.
	Container string `json:"container,omitempty"`

	// Command is the command and arguments to execute.
	Command []string `json:"command"`

	// OnError specifies how Velero should behave if it encounters an error executing this hook.
	OnError HookErrorMode `json:"onError,omitempty"`

	// Timeout defines the maximum amount of time Velero should wait for the hook to complete before
	// considering the execution a failure.
	Timeout metav1.Duration `json:"timeout,omitempty"`
}

// HookErrorMode defines how Velero should treat an error from a hook.
type HookErrorMode string

// UploaderConfigForBackup defines the configuration for the uploader when doing backup.
type UploaderConfigForBackup struct {
	// ParallelFilesUpload is the number of files parallel uploads to perform when using the uploader.
	ParallelFilesUpload int `json:"parallelFilesUpload,omitempty"`
}

// BackupStatus captures the current status of a Velero backup.
type BackupStatus struct {
	// Version is the backup format major version.
	// Deprecated: Please see FormatVersion
	Version int `json:"version,omitempty"`

	// FormatVersion is the backup format version, including major, minor, and patch version.
	FormatVersion string `json:"formatVersion,omitempty"`

	// Expiration is when this Backup is eligible for garbage-collection.
	Expiration *metav1.Time `json:"expiration,omitempty"`

	// Phase is the current state of the Backup.
	Phase BackupPhase `json:"phase,omitempty"`

	// ValidationErrors is a slice of all validation errors (if
	// applicable).
	ValidationErrors []string `json:"validationErrors,omitempty"`

	// StartTimestamp records the time a backup was started.
	// Separate from CreationTimestamp, since that value changes
	// on restores.
	// The server's time is used for StartTimestamps
	StartTimestamp *metav1.Time `json:"startTimestamp,omitempty"`

	// CompletionTimestamp records the time a backup was completed.
	// Completion time is recorded even on failed backups.
	// Completion time is recorded before uploading the backup object.
	// The server's time is used for CompletionTimestamps
	CompletionTimestamp *metav1.Time `json:"completionTimestamp,omitempty"`

	// VolumeSnapshotsAttempted is the total number of attempted
	// volume snapshots for this backup.
	VolumeSnapshotsAttempted int `json:"volumeSnapshotsAttempted,omitempty"`

	// VolumeSnapshotsCompleted is the total number of successfully
	// completed volume snapshots for this backup.
	VolumeSnapshotsCompleted int `json:"volumeSnapshotsCompleted,omitempty"`

	// FailureReason is an error that caused the entire backup to fail.
	FailureReason string `json:"failureReason,omitempty"`

	// Warnings is a count of all warning messages that were generated during
	// execution of the backup. The actual warnings are in the backup's log
	// file in object storage.
	Warnings int `json:"warnings,omitempty"`

	// Errors is a count of all error messages that were generated during
	// execution of the backup.  The actual errors are in the backup's log
	// file in object storage.
	Errors int `json:"errors,omitempty"`

	// Progress contains information about the backup's execution progress. Note
	// that this information is best-effort only -- if Velero fails to update it
	// during a backup for any reason, it may be inaccurate/stale.
	Progress *BackupProgress `json:"progress,omitempty"`

	// CSIVolumeSnapshotsAttempted is the total number of attempted
	// CSI VolumeSnapshots for this backup.
	CSIVolumeSnapshotsAttempted int `json:"csiVolumeSnapshotsAttempted,omitempty"`

	// CSIVolumeSnapshotsCompleted is the total number of successfully
	// completed CSI VolumeSnapshots for this backup.
	CSIVolumeSnapshotsCompleted int `json:"csiVolumeSnapshotsCompleted,omitempty"`

	// BackupItemOperationsAttempted is the total number of attempted
	// async BackupItemAction operations for this backup.
	BackupItemOperationsAttempted int `json:"backupItemOperationsAttempted,omitempty"`

	// BackupItemOperationsCompleted is the total number of successfully completed
	// async BackupItemAction operations for this backup.
	BackupItemOperationsCompleted int `json:"backupItemOperationsCompleted,omitempty"`

	// BackupItemOperationsFailed is the total number of async
	// BackupItemAction operations for this backup which ended with an error.
	BackupItemOperationsFailed int `json:"backupItemOperationsFailed,omitempty"`

	// HookStatus contains information about the status of the hooks.
	HookStatus *HookStatus `json:"hookStatus,omitempty"`
}

// BackupPhase is a string representation of the lifecycle phase
// of a Velero backup.
type BackupPhase string

// BackupProgress stores information about the progress of a Backup's execution.
type BackupProgress struct {
	// TotalItems is the total number of items to be backed up. This number may change
	// throughout the execution of the backup due to plugins that return additional related
	// items to back up, the velero.io/exclude-from-backup label, and various other
	// filters that happen as items are processed.
	TotalItems int `json:"totalItems,omitempty"`

	// ItemsBackedUp is the number of items that have actually been written to the
	// backup tarball so far.
	ItemsBackedUp int `json:"itemsBackedUp,omitempty"`
}

// HookStatus stores information about the status of the hooks.
type HookStatus struct {
	// HooksAttempted is the total number of attempted hooks
	// Specifically, HooksAttempted represents the number of hooks that failed to execute
	// and the number of hooks that executed successfully.
	HooksAttempted int `json:"hooksAttempted,omitempty"`

	// HooksFailed is the total number of hooks which ended with an error
	HooksFailed int `json:"hooksFailed,omitempty"`
}

// BackupList is a list of Backups.
type BackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items []BackupObject `json:"items"`
}

// DeleteBackupRequest is a request to delete one or more backups.
type DeleteBackupRequestObject struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec DeleteBackupRequestSpec `json:"spec,omitempty"`
	Status DeleteBackupRequestStatus `json:"status,omitempty"`
}

// DeleteBackupRequestSpec is the specification for which backups to delete.
type DeleteBackupRequestSpec struct {
	BackupName string `json:"backupName"`
}

// DeleteBackupRequestStatus is the current status of a DeleteBackupRequest.
type DeleteBackupRequestStatus struct {
	// Phase is the current state of the DeleteBackupRequest.
	Phase DeleteBackupRequestPhase `json:"phase,omitempty"`

	// Errors contains any errors that were encountered during the deletion process.
	Errors []string `json:"errors,omitempty"`
}

// DeleteBackupRequestPhase represents the lifecycle phase of a DeleteBackupRequest.
// +kubebuilder:validation:Enum=New;InProgress;Processed
type DeleteBackupRequestPhase string


// Restore


// Schedule
