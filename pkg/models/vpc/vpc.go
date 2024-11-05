/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com.tw
*/

package vpc

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	tenantv1alpha1 "kubesphere.io/api/tenant/v1alpha1"
	v1alpha2 "kubesphere.io/api/tenant/v1alpha2"
	v1 "kubesphere.io/api/vpc/v1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/informers"

	resourcesv1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
)

type Interface interface {
	GetVpcNetwork(vpcnetwork string) (*VPCNetworkResponse, error)
	ListVpcNetwork(query *query.Query) (*[]VPCNetworkResponse, error)
	GetGatewayChassisNode() ([]GatewayChassisNode, error)
	CreateVpcNetwork(workspace string, vpcnetwork *VPCNetwork) (*VPCNetwork, error)
	UpdateVpcNetwork(vpcnetwork *VPCNetworkBase, name string) (*VPCNetwork, error)
	PatchVpcNetwork(vpcnetworkName string, vpcnetwork *VPCNetworkPatch) (*VPCNetworkPatch, error)
	DeleteVpcNetwork(vpcnetwork string) error
	GetVpcSubnet(namespace, vpcsubnet string) (*VPCSubnetResponse, error)
	ListVpcSubnet(query *query.Query) (*[]VPCSubnetResponse, error)
	ListVpcSubnetWithinVpcNetwork(vpcnetwork string, queryParam *query.Query) (*[]VPCSubnetResponse, error)
	ListVpcSubnetWithinNamespace(namespace string, queryParam *query.Query) (*[]VPCSubnetResponse, error)
	CreateVpcSubnet(vpcsubnet *VPCSubnet) (*VPCSubnet, error)
	UpdateVpcSubnet(vpcsubnet *VPCSubnetBase, namespace string, vpcsubnetName string) (*VPCSubnetPutResponse, error)
	PatchVpcSubnet(vpcsubnet *VPCSubnetPatch, namespace string, vpcsubnetName string) (*VPCSubnetPatch, error)
	DeleteVpcSubnet(namespace, vpcsubnet string) error
}

type vpcOperator struct {
	ksclient       kubesphere.Interface
	k8sclient      kubernetes.Interface
	resourceGetter *resourcesv1alpha3.ResourceGetter
}

func New(informers informers.InformerFactory, k8sclient kubernetes.Interface, ksclient kubesphere.Interface) Interface {
	return &vpcOperator{
		resourceGetter: resourcesv1alpha3.NewResourceGetter(informers, nil),
		k8sclient:      k8sclient,
		ksclient:       ksclient,
	}
}

func (t *vpcOperator) ListVpcNetwork(queryParam *query.Query) (*[]VPCNetworkResponse, error) {

	result, err := t.resourceGetter.List(v1.ResourcePluralVpcNetworks, "", queryParam)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	vpcnetworks := []VPCNetworkResponse{}
	for _, item := range result.Items {
		v1Vpcnetwork, ok := item.(*v1.VPCNetwork)
		if ok {
			vpcnetworks = append(vpcnetworks, convertToVPCNetwork(v1Vpcnetwork))
		}
	}

	return &vpcnetworks, nil
}

func (t *vpcOperator) GetVpcNetwork(vpcnetwork string) (*VPCNetworkResponse, error) {
	vpcResource, err := t.DescribeVpcNetwork(vpcnetwork)
	if err != nil {
		return nil, err
	}

	vpc := convertToVPCNetworkBase(vpcResource)

	return &vpc, nil
}

func (t *vpcOperator) GetGatewayChassisNode() ([]GatewayChassisNode, error) {
	query := query.New()
	// This Label means which node could be a gateway node.
	query.LabelSelector = "k8s.ovn.org/ha-chassis-assignable"

	result, err := t.resourceGetter.List("nodes", "", query)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	gatewayNodes := []GatewayChassisNode{}
	for _, item := range result.Items {

		value, ok := item.(*corev1.Node)
		if ok {
			node := GatewayChassisNode{
				Node: value.Name,
			}
			gatewayNodes = append(gatewayNodes, node)
		} else {
			return nil, fmt.Errorf("items convert to node resource failed")
		}
	}

	return gatewayNodes, nil
}

func (t *vpcOperator) CreateVpcNetwork(workspaceName string, vpcnetwork *VPCNetwork) (*VPCNetwork, error) {

	// update vpc network name into workspace meatadata labels
	_, err := addVpcNetworkNameIntoWorkspace(t, workspaceName, vpcnetwork.Name)
	if err != nil {
		return nil, err
	}

	rawVPCNetwork := convertToRawVPCNetwork(&vpcnetwork.VPCNetworkBase, workspaceName, "", vpcnetwork.Name)

	_, err = t.ksclient.K8sV1().VPCNetworks().Create(context.Background(), &rawVPCNetwork, metav1.CreateOptions{})
	return vpcnetwork, err
}

func (t *vpcOperator) UpdateVpcNetwork(vpcnetwork *VPCNetworkBase, name string) (*VPCNetwork, error) {

	vpc, err := t.DescribeVpcNetwork(name)
	if err != nil {
		return nil, err
	}

	rawVPCNetwork := convertToRawVPCNetwork(vpcnetwork, vpc.Labels[tenantv1alpha1.WorkspaceLabel], vpc.ResourceVersion, name)

	_, err = t.ksclient.K8sV1().VPCNetworks().Update(context.Background(), &rawVPCNetwork, metav1.UpdateOptions{})

	vpcnetworkResponse := VPCNetwork{
		VPCNetworkBase: *vpcnetwork,
		Name:           name,
	}
	return &vpcnetworkResponse, err
}

func addVpcNetworkNameIntoWorkspace(t *vpcOperator, workspaceName string, vpcnetworkName string) (*v1.VPCNetwork, error) {
	_, err := t.resourceGetter.Get(v1alpha2.ResourcePluralWorkspaceTemplate, "", workspaceName)
	if err != nil {
		return nil, err
	}

	var workspaceTemplate = &v1alpha2.WorkspaceTemplate{}
	workspaceTemplate = labelWorkspaceWithVpcNetworkName(vpcnetworkName, workspaceTemplate)

	data, err := json.Marshal(workspaceTemplate)
	if err != nil {
		return nil, err
	}

	_, err = t.ksclient.TenantV1alpha2().WorkspaceTemplates().Patch(context.Background(), workspaceName, types.MergePatchType, data, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func deleteVpcNetworkNameFromWorkspace(t *vpcOperator, workspaceName string) error {
	obj, err := t.resourceGetter.Get(v1alpha2.ResourcePluralWorkspaceTemplate, "", workspaceName)
	if err != nil {
		klog.Error(err)
		return err
	}
	workspaceTemplate := obj.(*v1alpha2.WorkspaceTemplate)
	if workspaceTemplate.Labels != nil {
		delete(workspaceTemplate.Labels, v1.VpcNetworkLabel)
	}

	_, err = t.ksclient.TenantV1alpha2().WorkspaceTemplates().Update(context.Background(), workspaceTemplate, metav1.UpdateOptions{})
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

// labelWorkspaceWithVpcNetworkName adds a k8s.ovn.org/vpcnetwork=[vpcnetworkName] label to workspace which
// indicates vpcnetwork is under the workspace
func labelWorkspaceWithVpcNetworkName(vpcnetworkName string, workspace *v1alpha2.WorkspaceTemplate) *v1alpha2.WorkspaceTemplate {
	if workspace.Labels == nil {
		workspace.Labels = make(map[string]string, 0)
	}

	workspace.Labels[v1.VpcNetworkLabel] = vpcnetworkName

	return workspace
}

// labelNamespaceWithVpcSubnetName adds a k8s.ovn.org/vpcsubnet=[vpcsubnetName] label to namespace which
// indicates vpcsubnet is under the namespace
func labelNamespaceWithVpcSubnetName(vpcsubnetName string, namespace *corev1.Namespace) *corev1.Namespace {
	if namespace.Labels == nil {
		namespace.Labels = make(map[string]string, 0)
	}

	namespace.Labels[v1.VpcSubnetLabel] = vpcsubnetName

	return namespace
}

func (t *vpcOperator) PatchVpcNetwork(vpcnetworkName string, vpcnetwork *VPCNetworkPatch) (*VPCNetworkPatch, error) {
	_, err := t.DescribeVpcNetwork(vpcnetworkName)
	if err != nil {
		return nil, err
	}

	rawVPCNetwork := convertToRawVPCNetworkPatch(vpcnetwork)

	rawVPCNetworkSpec := map[string]interface{}{
		"spec": rawVPCNetwork,
	}
	data, err := json.Marshal(rawVPCNetworkSpec)

	if err != nil {
		return nil, err
	}

	_, err = t.ksclient.K8sV1().VPCNetworks().Patch(context.Background(), vpcnetworkName, types.MergePatchType, data, metav1.PatchOptions{})

	return vpcnetwork, err
}

func (t *vpcOperator) DeleteVpcNetwork(vpcnetwork string) error {

	vpc, err := t.DescribeVpcNetwork(vpcnetwork)
	if err != nil {
		return err
	}

	workspaceName := vpc.Labels[tenantv1alpha1.WorkspaceLabel]

	deleteVpcNetworkNameFromWorkspace(t, workspaceName)

	return t.ksclient.K8sV1().VPCNetworks().Delete(context.Background(), vpcnetwork, metav1.DeleteOptions{})
}

func (t *vpcOperator) DescribeVpcNetwork(vpcnetworkName string) (*v1.VPCNetwork, error) {
	obj, err := t.resourceGetter.Get(v1.ResourcePluralVpcNetworks, "", vpcnetworkName)
	if err != nil {
		return nil, err
	}
	vpcnetwork := obj.(*v1.VPCNetwork)
	return vpcnetwork, nil
}

func (t *vpcOperator) DescribeWorkspaceTemplate(workspaceName string) (*v1alpha2.WorkspaceTemplate, error) {
	obj, err := t.resourceGetter.Get(tenantv1alpha1.ResourcePluralWorkspace, "", workspaceName)
	if err != nil {
		return nil, err
	}

	workspace := obj.(*v1alpha2.WorkspaceTemplate)
	return workspace, nil
}

func (t *vpcOperator) ListVpcSubnet(queryParam *query.Query) (*[]VPCSubnetResponse, error) {

	result, err := t.resourceGetter.List(v1.ResourcePluralVpcSubnets, "", queryParam)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	vpcsubnets := []VPCSubnetResponse{}
	for _, item := range result.Items {
		rawVpcsubnet, ok := item.(*v1.VPCSubnet)
		if ok {
			vpcsubnets = append(vpcsubnets, convertToVPCSubnetResponse(rawVpcsubnet))
		}
	}

	return &vpcsubnets, nil
}

func (t *vpcOperator) ListVpcSubnetWithinVpcNetwork(vpcnetwork string, queryParam *query.Query) (*[]VPCSubnetResponse, error) {

	result, err := t.ksclient.K8sV1().VPCSubnets("").List(context.Background(), metav1.ListOptions{})

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	vpcsubnets := []VPCSubnetResponse{}
	for _, rawVpcsubnet := range result.Items {
		if rawVpcsubnet.Spec.Vpc == vpcnetwork {
			vpcsubnet := convertToVPCSubnetResponse(&rawVpcsubnet)
			vpcsubnets = append(vpcsubnets, vpcsubnet)
		}
	}

	return &vpcsubnets, nil
}

func (t *vpcOperator) ListVpcSubnetWithinNamespace(namespace string, queryParam *query.Query) (*[]VPCSubnetResponse, error) {

	result, err := t.ksclient.K8sV1().VPCSubnets(namespace).List(context.Background(), metav1.ListOptions{})

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	vpcsubnets := []VPCSubnetResponse{}
	for _, rawVpcsubnet := range result.Items {
		vpcsubnet := convertToVPCSubnetResponse(&rawVpcsubnet)
		vpcsubnets = append(vpcsubnets, vpcsubnet)
	}

	return &vpcsubnets, nil
}

func (t *vpcOperator) GetVpcSubnet(namespace, vpcsubnet string) (*VPCSubnetResponse, error) {
	obj, err := t.resourceGetter.Get(v1.ResourcePluralVpcSubnets, namespace, vpcsubnet)
	if err != nil {
		return nil, err
	}
	rawVpcSubnet := obj.(*v1.VPCSubnet)
	vpcSubnet := convertToVPCSubnetResponse(rawVpcSubnet)

	return &vpcSubnet, nil
}

func (t *vpcOperator) CreateVpcSubnet(vpcsubnet *VPCSubnet) (*VPCSubnet, error) {
	// update vpc subnet name into namespace meatadata labels
	_, err := addVpcSubnetNameIntoNamespace(t, vpcsubnet)
	if err != nil {
		return nil, err
	}

	// Assign VPC network into VPC Subnet SPEC vpc element.
	err = t.assignVPCNetworkIntoVPCSubnet(&vpcsubnet.VPCSubnetBase, vpcsubnet.Namespace)
	if err != nil {
		return nil, err
	}

	rawVpcSubnet := convertToRawVPCSubnet(&vpcsubnet.VPCSubnetBase, vpcsubnet.Namespace, vpcsubnet.Name, "")
	_, err = t.ksclient.K8sV1().VPCSubnets(vpcsubnet.Namespace).Create(context.Background(), rawVpcSubnet, metav1.CreateOptions{})

	return vpcsubnet, err
}

func addVpcSubnetNameIntoNamespace(t *vpcOperator, vpcsubnet *VPCSubnet) (*VPCSubnet, error) {
	_, err := t.resourceGetter.Get("namespaces", "", vpcsubnet.Namespace)
	if err != nil {
		return nil, err
	}

	var namespace = &corev1.Namespace{}
	namespace = labelNamespaceWithVpcSubnetName(vpcsubnet.Name, namespace)

	data, err := json.Marshal(namespace)
	if err != nil {
		return nil, err
	}

	_, err = t.k8sclient.CoreV1().Namespaces().Patch(context.Background(), vpcsubnet.Namespace, types.MergePatchType, data, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (t *vpcOperator) UpdateVpcSubnet(vpcsubnet *VPCSubnetBase, namespace string, vpcsubnetName string) (*VPCSubnetPutResponse, error) {

	obj, err := t.resourceGetter.Get(v1.ResourcePluralVpcSubnets, namespace, vpcsubnetName)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	vpc := obj.(*v1.VPCSubnet)

	// Assign VPC network into VPC Subnet SPEC vpc element.
	err = t.assignVPCNetworkIntoVPCSubnet(vpcsubnet, namespace)
	if err != nil {
		return nil, err
	}

	rawVPCSubnet := convertToRawVPCSubnet(vpcsubnet, namespace, vpcsubnetName, vpc.ResourceVersion)
	_, err = t.ksclient.K8sV1().VPCSubnets(namespace).Update(context.Background(), rawVPCSubnet, metav1.UpdateOptions{})

	vpcSubnetResponse := VPCSubnetPutResponse{
		VPCSubnetPut: VPCSubnetPut{
			CIDR: vpcsubnet.CIDR,
		},
		Name:      vpcsubnetName,
		Namespace: namespace,
	}
	return &vpcSubnetResponse, err
}

func (t *vpcOperator) PatchVpcSubnet(vpcsubnet *VPCSubnetPatch, namespace string, vpcsubnetName string) (*VPCSubnetPatch, error) {

	rawVPCSubnet := convertToRawVPCSubnetPatch(vpcsubnet)

	rawVPCSubnetSpec := map[string]interface{}{
		"spec": rawVPCSubnet,
	}
	data, err := json.Marshal(rawVPCSubnetSpec)

	if err != nil {
		return nil, err
	}

	_, err = t.ksclient.K8sV1().VPCSubnets(namespace).Patch(context.Background(), vpcsubnetName, types.MergePatchType, data, metav1.PatchOptions{})

	return vpcsubnet, err
}

func (t *vpcOperator) DeleteVpcSubnet(namespace, vpcsubnet string) error {
	vpc, err := t.DescribeVpcSubnet(namespace, vpcsubnet)
	if err != nil {
		return err
	}

	deleteVpcSubnetNameFromNamespace(t, namespace)
	return t.ksclient.K8sV1().VPCSubnets(vpc.Namespace).Delete(context.Background(), vpcsubnet, metav1.DeleteOptions{})
}

func (t *vpcOperator) DescribeVpcSubnet(namespace, vpcsubnetName string) (*v1.VPCSubnet, error) {
	obj, err := t.resourceGetter.Get(v1.ResourcePluralVpcSubnets, namespace, vpcsubnetName)
	if err != nil {
		return nil, err
	}
	vpcsbunet := obj.(*v1.VPCSubnet)
	return vpcsbunet, nil
}

func deleteVpcSubnetNameFromNamespace(t *vpcOperator, namespaceName string) error {
	obj, err := t.resourceGetter.Get("namespaces", "", namespaceName)
	if err != nil {
		klog.Error(err)
		return err
	}
	namespace := obj.(*corev1.Namespace)
	if namespace.Labels != nil {
		delete(namespace.Labels, v1.VpcSubnetLabel)
	}

	_, err = t.k8sclient.CoreV1().Namespaces().Update(context.Background(), namespace, metav1.UpdateOptions{})
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func (t *vpcOperator) assignVPCNetworkIntoVPCSubnet(vpcsubnet *VPCSubnetBase, namespace string) error {
	if vpcsubnet.Vpc == "" {
		ns, err := t.k8sclient.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
		if err != nil {
			return err
		}
		workspace := ns.Labels["kubesphere.io/workspace"]
		if workspace != "" {
			ws, err := t.ksclient.TenantV1alpha2().WorkspaceTemplates().Get(context.Background(), workspace, metav1.GetOptions{})
			if err != nil {
				return err
			}
			vpc := ws.Labels["k8s.ovn.org/vpcnetwork"]
			if vpc != "" {
				vpcsubnet.Vpc = vpc
			}
		}
	}
	return nil
}

func (t *vpcOperator) updateVPCNetworkIntoVPCSubnet(newVpcsubnet *VPCSubnetBase, namespace string) error {
	if newVpcsubnet.Vpc == "" {
		ns, err := t.k8sclient.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
		if err != nil {
			return err
		}
		workspace := ns.Labels["kubesphere.io/workspace"]
		if workspace != "" {
			ws, err := t.ksclient.TenantV1alpha2().WorkspaceTemplates().Get(context.Background(), workspace, metav1.GetOptions{})
			if err != nil {
				return err
			}
			vpc := ws.Labels["k8s.ovn.org/vpcnetwork"]
			if vpc != "" {
				newVpcsubnet.Vpc = vpc
			}
		}
	}
	return nil
}

func convertToVPCNetworkBase(vpcResource *v1.VPCNetwork) VPCNetworkResponse {
	vpc := VPCNetworkResponse{}
	vpc.Name = vpcResource.Name
	vpc.CIDR = vpcResource.Spec.CIDR
	vpc.Workspace = vpcResource.Labels[tenantv1alpha1.WorkspaceLabel]

	// Gateway Chassis
	gatewayChassises := []GatewayChassisResponse{}
	for _, gateway := range vpcResource.Spec.GatewayChassis {
		gatewayChassis := GatewayChassisResponse{
			Node: gateway.Node,
			IP:   gateway.IP,
		}
		gatewayChassises = append(gatewayChassises, gatewayChassis)
	}
	vpc.GatewayChassis = gatewayChassises

	// L3 Gateway
	gatewayes := []L3GatewayResponse{}
	for _, gateway := range vpcResource.Spec.L3Gateways {

		l3Gateway := L3GatewayResponse{
			Network:     gateway.Network,
			Destination: gateway.Destination,
			NextHop:     gateway.NextHop,
			VLANId:      gateway.VLANId,
		}
		gatewayes = append(gatewayes, l3Gateway)
	}
	vpc.L3Gateways = gatewayes
	return vpc
}

func convertToVPCNetwork(vpcResource *v1.VPCNetwork) VPCNetworkResponse {
	vpc := VPCNetworkResponse{}
	vpc.Name = vpcResource.Name
	vpc.CIDR = vpcResource.Spec.CIDR
	vpc.Workspace = vpcResource.Labels[tenantv1alpha1.WorkspaceLabel]

	// Gateway Chassis
	gatewayChassises := []GatewayChassisResponse{}
	for _, gateway := range vpcResource.Spec.GatewayChassis {
		gatewayChassis := GatewayChassisResponse{
			Node: gateway.Node,
			IP:   gateway.IP,
		}
		gatewayChassises = append(gatewayChassises, gatewayChassis)
	}
	vpc.GatewayChassis = gatewayChassises

	// L3 Gateway
	gatewayes := []L3GatewayResponse{}
	for _, gateway := range vpcResource.Spec.L3Gateways {

		l3Gateway := L3GatewayResponse{
			Network:     gateway.Network,
			Destination: gateway.Destination,
			NextHop:     gateway.NextHop,
			VLANId:      gateway.VLANId,
		}
		gatewayes = append(gatewayes, l3Gateway)
	}
	vpc.L3Gateways = gatewayes
	return vpc
}

func convertToRawVPCNetwork(vpcnetwork *VPCNetworkBase, workspaceName string, resourceVersion string, vpcnetworkName string) v1.VPCNetwork {
	// Gateway Chassis
	gatewayChassises := []v1.GatewayChassis{}
	for _, gateway := range vpcnetwork.GatewayChassis {
		gatewayChassis := v1.GatewayChassis{
			Node: gateway.Node,
			IP:   gateway.IP,
		}
		gatewayChassises = append(gatewayChassises, gatewayChassis)
	}

	// L3 gateway
	gatewayes := []v1.L3Gateway{}
	for _, gateway := range vpcnetwork.L3Gateways {
		l3Gateway := v1.L3Gateway{
			Network:     gateway.Network,
			Destination: gateway.Destination,
			NextHop:     gateway.NextHop,
			VLANId:      gateway.VLANId,
		}
		gatewayes = append(gatewayes, l3Gateway)
	}
	// labelNamespaceWithWorkspaceName adds a kubesphere.io/workspace=[workspaceName] label to namespace which
	// indicates namespace is under the workspace
	workspaceNameLabel := make(map[string]string, 0)
	workspaceNameLabel[tenantv1alpha1.WorkspaceLabel] = workspaceName
	length := getCIDRPrefix(vpcnetwork)

	rawVPCNetwork := v1.VPCNetwork{
		ObjectMeta: metav1.ObjectMeta{
			Name:            vpcnetworkName,
			Labels:          workspaceNameLabel,
			ResourceVersion: resourceVersion,
		},
		Spec: v1.VPCNetworkSpec{
			CIDR:           vpcnetwork.CIDR,
			SubnetLength:   length,
			GatewayChassis: gatewayChassises,
			L3Gateways:     gatewayes,
		},
	}
	return rawVPCNetwork
}

func getCIDRPrefix(vpcnetwork *VPCNetworkBase) int {
	prefixLength := strings.Split(vpcnetwork.CIDR, "/")
	length, _ := strconv.Atoi(prefixLength[1])
	return length
}

func convertToRawVPCNetworkPatch(vpcnetwork *VPCNetworkPatch) map[string]interface{} {
	// Gateway Chassis
	gatewayChassises := []v1.GatewayChassis{}
	for _, gateway := range vpcnetwork.GatewayChassis {
		gatewayChassis := v1.GatewayChassis{
			Node: gateway.Node,
			IP:   gateway.IP,
		}
		gatewayChassises = append(gatewayChassises, gatewayChassis)
	}

	// L3 gateway
	gatewayes := []v1.L3Gateway{}
	for _, gateway := range vpcnetwork.L3Gateways {
		l3Gateway := v1.L3Gateway{
			Network:     gateway.Network,
			Destination: gateway.Destination,
			NextHop:     gateway.NextHop,
			VLANId:      gateway.VLANId,
		}
		gatewayes = append(gatewayes, l3Gateway)
	}

	patchData := make(map[string]interface{})

	if vpcnetwork.CIDR != "" {
		patchData["cidr"] = vpcnetwork.CIDR
	}

	if len(gatewayChassises) > 0 {
		patchData["gatewayChassises"] = gatewayChassises
	}

	if len(gatewayes) > 0 {
		patchData["l3gateways"] = gatewayes
	}

	return patchData
}

func convertToRawVPCSubnetPatch(vpcsubnet *VPCSubnetPatch) map[string]interface{} {

	patchData := make(map[string]interface{})

	if vpcsubnet.CIDR != "" {
		patchData["cidr"] = vpcsubnet.CIDR
	}

	return patchData
}

func convertToVPCSubnet(vpcResource *v1.VPCSubnet) VPCSubnet {
	vpcSubnet := VPCSubnet{}

	vpcSubnet.Name = vpcResource.Name
	vpcSubnet.Namespace = vpcResource.Namespace
	vpcSubnet.CIDR = vpcResource.Spec.CIDR
	vpcSubnet.Vpc = vpcResource.Spec.Vpc

	return vpcSubnet
}

func convertToVPCSubnetResponse(vpcResource *v1.VPCSubnet) VPCSubnetResponse {
	vpcSubnet := VPCSubnetResponse{}

	vpcSubnet.Name = vpcResource.Name
	vpcSubnet.Namespace = vpcResource.Namespace
	vpcSubnet.CIDR = vpcResource.Spec.CIDR
	vpcSubnet.Vpc = vpcResource.Spec.Vpc

	return vpcSubnet
}

func convertToRawVPCSubnet(vpcResource *VPCSubnetBase, namespace string, vpcSubnetName string, resourceVersion string) *v1.VPCSubnet {
	rawVpcSubnet := v1.VPCSubnet{}

	rawVpcSubnet.Name = vpcSubnetName
	rawVpcSubnet.Namespace = namespace
	if vpcResource.CIDR != "" {
		rawVpcSubnet.Spec.CIDR = vpcResource.CIDR
	}
	if vpcResource.Vpc != "" {
		rawVpcSubnet.Spec.Vpc = vpcResource.Vpc
	}
	rawVpcSubnet.ResourceVersion = resourceVersion

	return &rawVpcSubnet
}
