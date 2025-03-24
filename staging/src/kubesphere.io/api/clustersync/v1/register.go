/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com.tw
*/

// Package v1 contains API Schema definitions for the clustersync v1 API group.
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package,register
// +k8s:defaulter-gen=TypeMeta
// +groupName=clustersync.ecpaas.io
package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects.
	SchemeGroupVersion = schema.GroupVersion{Group: "clustersync.ecpaas.io", Version: "v1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)


func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}
