/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com.tw
*/

package apis

import (
	snapv1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, snapv1.SchemeBuilder.AddToScheme)
}
