/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com.tw
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	ResourceKindVpcNetwork     = "VPCNetwork"
	ResourceSingularVpcNetwork = "vpcnetwork"
	ResourcePluralVpcNetworks  = "vpcnetworks"
	VpcNetworkLabel            = "k8s.ovn.org/vpcnetwork"
)

// vpc network runtime information
type VPCNetworkStatus struct {
	// List of subnets created under the current network, separated by commas
	Subnets string `json:"subnets"`
	// Transit Switch
	TransitSwitch string `json:"transitSwitch,omitempty"`
	// Transit switch port information
	TsPort string `json:"tsPort,omitempty"`
	// Transit switch IP address
	TsNetwork string `json:"tsNetwork,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="SUBNETS",type=string,JSONPath=".status.subnets"
// +kubebuilder:printcolumn:name="TransitSwitch Port",type=string,JSONPath=".status.tsPort"
// +kubebuilder:printcolumn:name="TransitSwitch Network",type=string,JSONPath=".status.tsNetwork"
// +kubebuilder:resource:scope=Cluster, shortName=vnet
// A vpc network has a set of independent virtual k8s network topology.
// In this set of virtual k8s network, users add namespaces to the virtual k8s network by creating subnets.
// Its behavior is like adding new k8s nodes in the real k8s network is also called default vpc.
type VPCNetwork struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VPCNetworkSpec   `json:"spec"`
	Status            VPCNetworkStatus `json:"status,omitempty"`
}

// Describes the gateway information of the vpc network
type GatewayChassis struct {
	// +kubebuilder:validation:Optional
	Node string `json:"node,omitempty" description:"Name of the k8s node where the gateway is located"` /// 网关所在节点
	// +kubebuilder:validation:Required
	IP string `json:"ip" description:"Gateway IP address"` /// 网关地址
}

type GatewayChassisNode struct {
	// Name of the k8s node where the gateway is located
	Node []string `json:"node"`
}

type L3Gateway struct {
	// +kubebuilder:validation:Required
	Network string `json:"network" description:"L3 gateway address"`

	// +kubebuilder:validation:Optional
	// +optional
	Destination string `json:"destination,omitempty" description:"route DST"`

	// +kubebuilder:validation:Required
	// Next hop address
	NextHop string `json:"nexthop"`

	// +kubebuilder:validation:Optional
	// +optional
	VLANId int32 `json:"vlanid,omitempty" description:"VLAN id for external network" minimum:"1" maximum:"4094"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=default
	// +optional
	OutBoundNat string `json:"outboundNat,omitempty" default:"default" description:"outgoingnat"`
}

// Peer cluster connection information
type Peer struct {
	// +kubebuilder:validation:Required
	// Peer cluster name
	Name string `json:"name"` /// 对端K8S集群名称
	// +kubebuilder:validation:Required
	// Peer cluster address
	IP string `json:"ip"` /// 对端K8S集群连接地址
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum:=1
	// +kubebuilder:validation:Maximum:=65535
	// Peer cluster port
	Port int32 `json:"port"` /// 对端K8S集群连接端口
}

// NATRule defines the nat rule on router.
type NATRule struct {
	// Type of NAT rule, must be one of dnat, dnat_and_snat, or snat
	// +kubebuilder:validation:Pattern=^SNAT|DNAT|DNAT_AND_SNAT$
	Type string `json:"type"`
	// NAT prefix, must be a network(CIDR) or an ip address
	LogicalIP string `json:"logicalIP"`
	// external ip address for nat.
	ExternalIP string `json:"externalIP"`
	// The name of the logical port where the logical_ip resides.
	// +optional
	// +kubebuilder:validation:Optional
	Port string `json:"port,omitempty"`
}

type ClusterRouterPolicy struct {
	// +kubebuilder:validation:Required
	// logical ip cidr
	Destination string `json:"destination"`

	// +kubebuilder:validation:Required
	// target port
	TargetPort string `json:"targetPort"`
}

// Configuration information of virtual k8s network
type VPCNetworkSpec struct {
	// +kubebuilder:validation:Required
	CIDR string `json:"cidr" description:"vpc network private segment address space"` /// 网络CIDR
	// +kubebuilder:validation:Required
	SubnetLength int `json:"subnetLength" description:"Length of vpc subnet managed by vpc network" minimum:"0" maximum:"32"` /// VPC网络子网长度

	// +optional
	GatewayChassis []GatewayChassis `json:"gatewayChassises,omitempty" description:"Gateway chassis information of vpc network"` /// VPC网关配置

	// +optional
	L3Gateways []L3Gateway `json:"l3gateways,omitempty" description:"L3Gateway information of vpc network"` /// VPC外网网关配置

	// +optional
	Peers []Peer `json:"peers,omitempty" description:"Interconnected peer cluster information"` /// VPC集群互联配置

	// +optional
	ClusterRouter string `json:"clusterRouter,omitempty" description:"ClusterRouter specify which T0 router to connect with"`

	// +optional
	ClusterRouterPolicies []ClusterRouterPolicy `json:"clusterRouterPolicy,omitempty" description:"CluterRouterPolcies specify the traffic policy"`

	// +optional
	Nats []NATRule `json:"nat,omitempty" description:"Nat rules which applied to vpc t1 router"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=vpcnetwork
// VPCNetworkList is a list of VPCNetwork
type VPCNetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []VPCNetwork `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VPCNetwork{}, &VPCNetworkList{})
}
