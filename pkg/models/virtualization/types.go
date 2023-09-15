/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com
*/

package virtualization

type VirtualMachine struct {
	Name     string `json:"name"`
	CpuCores int32  `json:"cpu_cores"`
	Memory   string `json:"memory"`
	Disk     int32  `json:"disk,omitempty"`
	Image    string `json:"image,omitempty"`
}
