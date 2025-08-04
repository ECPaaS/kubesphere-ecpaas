/*
Copyright(c) 2025-present Accton. All rights reserved. www.accton.com
*/

package pvc

type PVCCloneRequest struct {
	NewName      string         `json:"newName" description:"New PVC name. Must be unique. Valid characters: a-z, 0-9, and -(hyphen). And must start and end with an alphanumeric character." maximum:"253"`
	Namespace    string         `json:"namespace" description:"Namespace in which the new PVC will be created in."`
	StorageClass string         `json:"storageClass" description:"Storage class of the new PVC."`
	Size         int            `json:"size" default:"10" description:"Size of the new PVC, unit is GiBs." minimum:"1" maximum:"2048"`
	AccessModes  []string       `json:"accessModes" description:"Access mode for the new PVC, at least one mode should be picked. Available option: ReadWriteOnce, ReadOnlyMany, ReadWriteMany"`
	NodeSelector []NodeSelector `json:"node_selector,omitempty" description:"Node selector of the pod to create PVC."`
}

type NodeSelector struct {
	Key   string `json:"key" description:"NodeSelector key(unique key). Valid characters: A-Z, a-z, 0-9, -(hyphen), _(underscore), .(dot), and /(slash). Must start and end with a letter or number. The maximum length of each key is 63 characters (if the key contains a domain name, the maximum domain name length is 253 characters plus 1 for seperation(/))." minimum:"1" maximum:"317"`
	Value string `json:"value" description:"NodeSelector value. Valid characters: A-Z, a-z, 0-9, -(hyphen), _(underscore), and .(dot). Must start and end with an alphanumeric charactor. Can be empty string." maximum:"63"`
}

type PVCNameResponse struct {
	NewName string `json:"newName" description:"New PVC name."`
}

func ConvertNodeSelectorToMap(ns []NodeSelector) map[string]string {
	returnMap := make(map[string]string, 0)
	for _, selector := range ns {
		returnMap[selector.Key] = selector.Value
	}
	return returnMap
}
