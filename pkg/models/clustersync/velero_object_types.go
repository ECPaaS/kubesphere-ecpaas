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


// Restore


// Schedule
