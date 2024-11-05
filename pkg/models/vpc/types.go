/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com
*/

package vpc

type VPCNetworkBase struct {
	// +kubebuilder:validation:Required
	CIDR string `json:"cidr" description:"vpc network private segment address space with cidr format, e.g., 10.0.0.0/16"`
	// +optional
	GatewayChassis []GatewayChassis `json:"gatewayChassises,omitempty" description:"Gateway chassis information of vpc network"`
	// +optional
	L3Gateways []L3Gateway `json:"l3gateways,omitempty" description:"L3Gateway information of vpc network"`
}

type VPCNetworkPatch struct {
	// +optional
	CIDR string `json:"cidr,omitempty" description:"vpc network private segment address space with cidr format, e.g., 10.0.0.0/16"`
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

type VPCNetworkResponse struct {
	// +kubebuilder:validation:Required
	Name string `json:"name" description:"VPC network name [unique key]"`
	// +kubebuilder:validation:Required
	CIDR string `json:"cidr" description:"vpc network private segment address space with cidr format, e.g., 10.0.0.0/16"`
	// +kubebuilder:validation:Required
	Workspace string `json:"workspace" description:"workspace name"`
	// +kubebuilder:validation:Required
	GatewayChassis []GatewayChassisResponse `json:"gatewayChassises" description:"Gateway chassis information of vpc network"`
	// +kubebuilder:validation:Required
	L3Gateways []L3GatewayResponse `json:"l3gateways" description:"L3Gateway information of vpc network"`
}

type ListVPCNetworkResponse struct {
	TotalCount int                  `json:"total_count" description:"Total number of VPC Network"`
	Items      []VPCNetworkResponse `json:"items" description:"List of VPC Network"`
}

type L3Gateway struct {
	// +kubebuilder:validation:Required
	Network string `json:"network" description:"L3 gateway address, e.g., 192.168.41.75/22 [unique key]"`
	// +kubebuilder:validation:Optional
	Destination string `json:"destination,omitempty" description:"route destination e.g., 0.0.0.0/0"`
	// +kubebuilder:validation:Required
	NextHop string `json:"nexthop" description:"Next hop address e.g., 192.168.40.254"`
	// +kubebuilder:validation:Optional
	// +optional
	VLANId int32 `json:"vlanid,omitempty" description:"VLAN id for external network" minimum:"0" maximum:"4094"`
}

type L3GatewayResponse struct {
	// +kubebuilder:validation:Required
	Network string `json:"network" description:"L3 gateway address, e.g., 192.168.41.75/22 [unique key]"`
	// +kubebuilder:validation:Required
	Destination string `json:"destination" description:"route destination e.g., 0.0.0.0/0"`
	// +kubebuilder:validation:Required
	NextHop string `json:"nexthop" description:"Next hop address e.g., 192.168.40.254"`
	// +kubebuilder:validation:Required
	VLANId int32 `json:"vlanid" description:"VLAN id for external network" minimum:"0" maximum:"4094"`
}

type GatewayChassis struct {
	// +kubebuilder:validation:Optional
	Node string `json:"node,omitempty" description:"Name of the k8s node where the gateway is located"`
	// +kubebuilder:validation:Required
	IP string `json:"ip" description:"Gateway IP address, e.g., 192.168.41.75 [unique key]"`
}

type GatewayChassisResponse struct {
	// +kubebuilder:validation:Required
	Node string `json:"node" description:"Name of the k8s node where the gateway is located"`
	// +kubebuilder:validation:Required
	IP string `json:"ip" description:"Gateway IP address, e.g., 192.168.41.75 [unique key]"`
}

type VPCSubnetBase struct {
	// +kubebuilder:validation:Required
	CIDR string `json:"cidr" description:"vpc subnet private segment address space with cidr format, e.g., 10.0.2.0/24"`
	// +kubebuilder:validation:Optional
	Vpc string `json:"vpc,omitempty" description:"vpc network name"`
}

type VPCSubnetPut struct {
	// +kubebuilder:validation:Optional
	CIDR string `json:"cidr,omitempty" description:"vpc subnet private segment address space with cidr format, e.g., 10.0.2.0/24"`
}

type VPCSubnetPatch struct {
	// +kubebuilder:validation:Optional
	CIDR string `json:"cidr,omitempty" description:"vpc subnet private segment address space with cidr format, e.g., 10.0.2.0/24"`
}

type VPCSubnetNameSpace struct {
	// +kubebuilder:validation:Required
	Name string `json:"name" maximum:"253" description:"must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character. Allowed characters: lowercase letters (a-z), numbers (0-9), and hyphens (-)"`
	VPCSubnetBase
}

type VPCSubnet struct {
	// +kubebuilder:validation:Required
	Name string `json:"name" maximum:"253" description:"must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character. Allowed characters: lowercase letters (a-z), numbers (0-9), and hyphens (-)"`
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`
	VPCSubnetBase
}

type VPCSubnetPutResponse struct {
	// +kubebuilder:validation:Required
	Name string `json:"name" maximum:"253" description:"must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character. Allowed characters: lowercase letters (a-z), numbers (0-9), and hyphens (-)"`
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`
	VPCSubnetPut
}

type VPCSubnetResponse struct {
	// +kubebuilder:validation:Required
	Name string `json:"name" description:"vpc subnet name [unique key]"`
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`
	// +kubebuilder:validation:Required
	CIDR string `json:"cidr" description:"vpc subnet private segment address space with cidr format, e.g., 10.0.2.0/24"`
	// +kubebuilder:validation:Required
	Vpc string `json:"vpc" description:"vpc network name"`
}

type ListVPCSubnetResponse struct {
	TotalCount int                 `json:"total_count" description:"Total number of VPC Subnet"`
	Items      []VPCSubnetResponse `json:"items" description:"List of VPC Subnet"`
}

type GatewayChassisNode struct {
	// +kubebuilder:validation:Required
	Node string `json:"node" description:"Name of the k8s node where the gateway is located"`
}

type ListGatewayChassisNodeResponse struct {
	TotalCount int                  `json:"total_count" description:"Total number of gateway chassis node"`
	Items      []GatewayChassisNode `json:"items" description:"List of gateway chassis node"`
}
