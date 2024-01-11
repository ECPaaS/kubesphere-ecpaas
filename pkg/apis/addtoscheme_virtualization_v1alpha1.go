/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com.tw
*/

package apis

import (
	virtzv1alpha1 "kubesphere.io/api/virtualization/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, virtzv1alpha1.SchemeBuilder.AddToScheme)
}
