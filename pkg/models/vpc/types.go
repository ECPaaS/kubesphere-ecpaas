/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com
*/

package vpc

type VPCNetworkBase struct {
	CIDR string `json:"cidr" description:"vpc network private segment address space" default:"10.0.0.0/16"`
	// +kubebuilder:validation:Required
	SubnetLength int `json:"subnetLength" description:"Length of vpc subnet managed by vpc network" minimum:"0" maximum:"32" default:"24"`
	// +optional
	GatewayChassis []GatewayChassis `json:"gatewayChassises,omitempty" description:"Gateway chassis information of vpc network"`
	// +optional
	L3Gateways []L3Gateway `json:"l3gateways,omitempty" description:"L3Gateway information of vpc network"`
}

type VPCNetwork struct {
	// +kubebuilder:validation:Required
	Name string `json:"name" maximum:"253" description:"must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character. Allowed characters: lowercase letters (a-z), numbers (0-9), and hyphens (-)"`
	VPCNetworkBase
}

type L3Gateway struct {
	// +kubebuilder:validation:Required
	Network string `json:"network" description:"L3 gateway address" default:"192.168.41.75/22"`
	// +kubebuilder:validation:Optional
	Destination string `json:"destination,omitempty" description:"route DST" default:"0.0.0.0/0"`
	// +kubebuilder:validation:Required
	NextHop string `json:"nexthop" description:"Next hop address" default:"192.168.40.254"`
	// +kubebuilder:validation:Optional
	// +optional
	VLANId int32 `json:"vlanid,omitempty" description:"VLAN id for external network" minimum:"0" maximum:"4094" default:"0"`
}

type GatewayChassis struct {
	// +kubebuilder:validation:Optional
	Node string `json:"node,omitempty" description:"Name of the k8s node where the gateway is located" default:"node1"`
	// +kubebuilder:validation:Required
	IP string `json:"ip" description:"Gateway IP address" default:"192.168.41.75"`
}

type VPCSubnetBase struct {
	// +kubebuilder:validation:Required
	CIDR string `json:"cidr" description:"vpc subnet private segment address space" default:"10.0.2.0/24"`
	// +kubebuilder:validation:Optional
	Vpc string `json:"vpc,omitempty" description:"vpc network name" default:"nocsys"`
}

type VPCSubnet struct {
	// +kubebuilder:validation:Required
	Name string `json:"name" maximum:"253" description:"must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character. Allowed characters: lowercase letters (a-z), numbers (0-9), and hyphens (-)"`
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`
	VPCSubnetBase
}

type GatewayChassisNode struct {
	// +kubebuilder:validation:Required
	Node string `json:"node" description:"Name of the k8s node where the gateway is located"`
}
