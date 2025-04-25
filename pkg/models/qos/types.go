/*
Copyright(c) 2025-present Accton. All rights reserved. www.accton.com
*/

package qos

type Dscp struct {
	Namespace string `json:"namespace" description:"namespace [unique key]"`
	Dscp      int64  `json:"dscp" description:"DSCP value" minimum:"0" maximum:"63"`
}

type DscpWithoutNs struct {
	Dscp int64 `json:"dscp" description:"DSCP value" minimum:"0" maximum:"63"`
}

type DscpList struct {
	Items      []Dscp `json:"items"`
	TotalCount int    `json:"total_count" description:"Total number of the DSCP items"`
}
