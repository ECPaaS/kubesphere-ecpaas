/*
Copyright(c) 2025-present Accton. All rights reserved. www.accton.com
*/

package autoscaler

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
)

const (
	annotationKey = "ecpaas.io/autoscaler"
	hpaAnnotaitonValue = "hpa"
	vpaAnnotaitonValue = "vpa"
	cpuCurrentUtilization = "cpuCurrentUtilization"
	cpuTargetUtilization = "cpuTargetUtilization"
	memoryCurrentValue = "memoryCurrentValue"
	memoryTargetValue = "memoryTargetValue"
	vpaPath = "/apis/autoscaling.k8s.io/v1/namespaces/%s/verticalpodautoscalers"
	vpaAPIVersion = "autoscaling.k8s.io/v1"
	vpaKind = "VerticalPodAutoscaler"
	unitFactor = 1024 * 1024 // From MiBs to Bs
)

var trueFlag bool = true

type Interface interface {
	// HPA

	CreateHPA(namespace string, name string, ui_hpa *HpaRequest) (*HpaNameResponse, error)
	UpdateHPA(namespace string, name string, ui_hpa *ModifyHpaRequest) (*HpaResponse, error)
	GetHPA(namespace string, name string) (*HpaResponse, error)
	DeleteHPA(namespace string, name string) error

	// VPA

	CreateVPA(namespace string, name string, ui_hpa *VpaRequest) (*VpaNameResponse, error)
	UpdateVPA(namespace string, name string, ui_hpa *ModifyVpaRequest) (*VpaResponse, error)
	GetVPA(namespace string, name string) (*VpaResponse, error)
	DeleteVPA(namespace string, name string) error
}

type autoscalerOperator struct {
	ksClient   kubesphere.Interface
	k8sClient  kubernetes.Interface
	restClient rest.Interface
}

func New(ksclient kubesphere.Interface, k8sclient kubernetes.Interface) Interface {
	var restclient rest.Interface

	if k8sclient != nil {
		restclient = k8sclient.AppsV1().RESTClient()
	}

	return &autoscalerOperator{
		ksClient:   ksclient,
		k8sClient:  k8sclient,
		restClient: restclient,
	}
}

// HPA

func (o *autoscalerOperator) CreateHPA(namespace string, name string, ui_hpa *HpaRequest) (*HpaNameResponse, error) {
	klog.V(2).Infof("Creating HPA for deployment: \"%s\" in \"%s\" namespace", name, namespace)
	// Get deployment, check if any autoscaler is set
	deployment, err := o.k8sClient.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else if _, ok := deployment.Annotations[annotationKey]; ok {
		// Already has autoscaler on deployment
		return nil,  fmt.Errorf("target deployment \"%s\" in \"%s\" namespace already has autoscaler, abort", name, namespace)
	}

	// Format HPA
	hpa := &autoscalingv2beta2.HorizontalPodAutoscaler{}
	hpa.Name = name
	hpa.Namespace = namespace
	hpa.Annotations = map[string]string{
		cpuCurrentUtilization: "",
		cpuTargetUtilization: fmt.Sprint(ui_hpa.TargetCpuUsage),
		memoryCurrentValue: "",
		memoryTargetValue: fmt.Sprint(ui_hpa.TargetMemoryUsage) + "Mi",
	}
	hpa.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: "apps/v1",
			Kind: "Deployment",
			Name: name,
			UID: deployment.UID,
			BlockOwnerDeletion: &trueFlag,
			Controller: &trueFlag,
		},
	}
	hpa.Spec.ScaleTargetRef = autoscalingv2beta2.CrossVersionObjectReference{
		APIVersion: "apps/v1",
		Kind: "Deployment",
		Name: name,
	}
	hpa.Spec.Metrics = createHpaMetrics(ui_hpa.TargetCpuUsage, ui_hpa.TargetMemoryUsage)
	hpa.Spec.MinReplicas = &ui_hpa.MinReplicas
	hpa.Spec.MaxReplicas = ui_hpa.MaxReplicas
	hpa.Spec.Behavior = createHpaBehaviors(ui_hpa.ScaleUp, ui_hpa.ScaleDown)

	// Create HPA
	_, err = o.k8sClient.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Create(context.Background(), hpa, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	} else {
		// PUT deploment annotations
		deployment.Annotations[annotationKey] = hpaAnnotaitonValue
		_, err := o.k8sClient.AppsV1().Deployments(namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}
	}

	return &HpaNameResponse{ResourceName: name}, nil
}

func (o *autoscalerOperator) UpdateHPA(namespace string, name string, ui_hpa *ModifyHpaRequest) (*HpaResponse, error) {
	klog.V(2).Infof("Updating HPA for deployment: \"%s\" in \"%s\" namespace", name, namespace)
	// Get deployment, check if HPA is set
	deployment, err := o.k8sClient.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else if value, ok := deployment.Annotations[annotationKey]; !ok {
		// No HPA on target deployment
		return nil, fmt.Errorf("target deployment \"%s\" in \"%s\" namespace has no HPA, abort", name, namespace)
	} else if value != hpaAnnotaitonValue {
		// Other autoscaler on target deployment
		return nil, fmt.Errorf("target deployment \"%s\" in \"%s\" namespace has other autoscaler, abort", name, namespace)
	}

	// Get HPA
	oldHPA, err := o.k8sClient.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	newHPA := oldHPA.DeepCopy()

	// Format HPA
	metrics := make([]autoscalingv2beta2.MetricSpec, 0)
	if ui_hpa.TargetCpuUsage != nil {
		metrics = append(metrics, createHpaMetrics(*ui_hpa.TargetCpuUsage, 0.0)...)
		newHPA.Annotations[cpuTargetUtilization] = fmt.Sprint(*ui_hpa.TargetCpuUsage)
	} else if ui_hpa.TargetMemoryUsage != nil {
		newHPA.Annotations[cpuTargetUtilization] = ""
	}
	if ui_hpa.TargetMemoryUsage != nil {
		metrics = append(metrics, createHpaMetrics(0, *ui_hpa.TargetMemoryUsage)...)
		newHPA.Annotations[memoryTargetValue] = fmt.Sprint(*ui_hpa.TargetMemoryUsage) + "Mi"
	} else if ui_hpa.TargetCpuUsage != nil {
		newHPA.Annotations[memoryTargetValue] = ""
	}
	if len(metrics) != 0 {
		newHPA.Spec.Metrics = metrics
	}
	if ui_hpa.MinReplicas != nil {
		newHPA.Spec.MinReplicas = ui_hpa.MinReplicas
	}
	if ui_hpa.MaxReplicas != nil {
		newHPA.Spec.MaxReplicas = *ui_hpa.MaxReplicas
	}
	if ui_hpa.ScaleUp != nil {
		newHPA.Spec.Behavior.ScaleUp = createHpaScalingRules(ui_hpa.ScaleUp)
	}
	if ui_hpa.ScaleDown != nil {
		newHPA.Spec.Behavior.ScaleDown = createHpaScalingRules(ui_hpa.ScaleDown)
	}
	if reflect.DeepEqual(newHPA.Spec, oldHPA.Spec) {
		// No update, abort
		return makeHpaResponse(newHPA), nil
	}

	// Update HPA
	updatedHPA, err := o.k8sClient.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Update(context.Background(), newHPA, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	
	return makeHpaResponse(updatedHPA), nil
}

func (o *autoscalerOperator) GetHPA(namespace string, name string) (*HpaResponse, error) {
	klog.V(2).Infof("Getting HPA for deployment: \"%s\" in \"%s\" namespace", name, namespace)

	// Get HPA
	hpa, err := o.k8sClient.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return makeHpaResponse(hpa), nil
}

func (o *autoscalerOperator) DeleteHPA(namespace string, name string) error {
	klog.V(2).Infof("Deleting HPA for deployment: \"%s\" in \"%s\" namespace", name, namespace)

	// Delete HPA
	err := o.k8sClient.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	// Get deployment, remove the annotation
	deployment, err := o.k8sClient.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return err
	} else {
		delete(deployment.Annotations, annotationKey)
		_, err := o.k8sClient.AppsV1().Deployments(namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func createHpaMetrics(cpu int32, memory float32) []autoscalingv2beta2.MetricSpec {
	returnArray := make([]autoscalingv2beta2.MetricSpec, 0)
	if cpu != 0 {
		returnArray = append(returnArray, autoscalingv2beta2.MetricSpec{
			Type: "Resource",
			Resource: &autoscalingv2beta2.ResourceMetricSource{
				Name: "cpu",
				Target: autoscalingv2beta2.MetricTarget{
					Type: "Utilization",
					AverageUtilization: &cpu,
				},
			},
		})
	}

	if memory != 0.0 {
		quantity := resource.MustParse(fmt.Sprint(memory) + "Mi")
		returnArray = append(returnArray, autoscalingv2beta2.MetricSpec{
			Type: "Resource",
			Resource: &autoscalingv2beta2.ResourceMetricSource{
				Name: "memory",
				Target: autoscalingv2beta2.MetricTarget{
					Type: "AverageValue",
					AverageValue: &quantity,
				},
			},
		})
	}

	return returnArray
}

func createHpaBehaviors(up *ScalingRules, down *ScalingRules) *autoscalingv2beta2.HorizontalPodAutoscalerBehavior {
	returnBehaviors := &autoscalingv2beta2.HorizontalPodAutoscalerBehavior{}
	if up != nil {
		returnBehaviors.ScaleUp = createHpaScalingRules(up)
	}

	if down != nil {
		returnBehaviors.ScaleDown = createHpaScalingRules(down)
	}

	return returnBehaviors
}

func createHpaScalingRules(rules *ScalingRules) *autoscalingv2beta2.HPAScalingRules {
	return &autoscalingv2beta2.HPAScalingRules{
		StabilizationWindowSeconds: rules.StabilizationWindowSeconds,
		SelectPolicy: (*autoscalingv2beta2.ScalingPolicySelect)(&rules.SelectPolicy),
		Policies: createHpaScalingPolicies(rules.Policies),
	}
}

func createHpaScalingPolicies(policies []ScalingPolicy) []autoscalingv2beta2.HPAScalingPolicy {
	returnPolicies := make([]autoscalingv2beta2.HPAScalingPolicy, 0)
	for _, policy := range policies {
		returnPolicies = append(returnPolicies, autoscalingv2beta2.HPAScalingPolicy{
			Type: autoscalingv2beta2.HPAScalingPolicyType(policy.Type),
			Value: policy.Value,
			PeriodSeconds: policy.PeriodSeconds,
		})
	}
	return returnPolicies
}

func makeHpaResponse(hpa *autoscalingv2beta2.HorizontalPodAutoscaler) *HpaResponse {
	cpuUsage, _ := strconv.ParseInt(hpa.Annotations[cpuTargetUtilization], 10, 32)
	memUsage, _ := strconv.ParseFloat(strings.TrimSuffix(hpa.Annotations[memoryTargetValue], "Mi"), 32)
	return &HpaResponse{
		ResourceName: hpa.Name,
		TargetCpuUsage: int32(cpuUsage),
		TargetMemoryUsage: float32(memUsage),
		MinReplicas: *hpa.Spec.MinReplicas,
		MaxReplicas: hpa.Spec.MaxReplicas,
		ScaleUp: ScalingRules{
			StabilizationWindowSeconds: hpa.Spec.Behavior.ScaleUp.StabilizationWindowSeconds,
			SelectPolicy: string(*hpa.Spec.Behavior.ScaleUp.SelectPolicy),
			Policies: makeHpaPoliciesResponse(hpa.Spec.Behavior.ScaleUp.Policies),
		},
		ScaleDown: ScalingRules{
			StabilizationWindowSeconds: hpa.Spec.Behavior.ScaleDown.StabilizationWindowSeconds,
			SelectPolicy: string(*hpa.Spec.Behavior.ScaleDown.SelectPolicy),
			Policies: makeHpaPoliciesResponse(hpa.Spec.Behavior.ScaleDown.Policies),
		},
	}
}

func makeHpaPoliciesResponse(policies []autoscalingv2beta2.HPAScalingPolicy) []ScalingPolicy {
	returnPolicies := make([]ScalingPolicy, 0)
	for _, policy := range policies {
		returnPolicies = append(returnPolicies, ScalingPolicy{
			Type: string(policy.Type),
			Value: policy.Value,
			PeriodSeconds: policy.PeriodSeconds,
		})
	}
	return returnPolicies
}


// VPA

func (o *autoscalerOperator) CreateVPA(namespace string, name string, ui_vpa *VpaRequest) (*VpaNameResponse, error) {
	klog.V(2).Infof("Creating VPA for deployment: \"%s\" in \"%s\" namespace", name, namespace)
	// Get deployment, check if any autoscaler is set
	deployment, err := o.k8sClient.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else if _, ok := deployment.Annotations[annotationKey]; ok {
		// Already has autoscaler on deployment
		return nil,  fmt.Errorf("target deployment \"%s\" in \"%s\" namespace already has autoscaler, abort", name, namespace)
	}

	// Format VPA
	vpa := &vpav1.VerticalPodAutoscaler{}
	vpa.APIVersion = vpaAPIVersion
	vpa.Kind = vpaKind
	vpa.Name = name
	vpa.Namespace = namespace
	vpa.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: "apps/v1",
			Kind: "Deployment",
			Name: name,
			UID: deployment.UID,
			BlockOwnerDeletion: &trueFlag,
			Controller: &trueFlag,
		},
	}
	vpa.Spec.TargetRef = &autoscalingv1.CrossVersionObjectReference{
		APIVersion: "apps/v1",
		Kind: "Deployment",
		Name: name,
	}
	vpa.Spec.UpdatePolicy = &vpav1.PodUpdatePolicy{
		UpdateMode: (*vpav1.UpdateMode)(&ui_vpa.UpdateMode),
		MinReplicas: &ui_vpa.MinReplicas,
	}
	vpa.Spec.ResourcePolicy = &vpav1.PodResourcePolicy{
		ContainerPolicies: createVpaContainerPolicies(ui_vpa.CPolicies),
	}

	// Create VPA
	bytes, err := json.Marshal(vpa)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf(vpaPath, namespace)
	_, err = o.restClient.Post().AbsPath(path).Body(bytes).DoRaw(context.Background())
	if err != nil {
		return nil, err
	} else {
		// PUT deployment annotations
		deployment.Annotations[annotationKey] = vpaAnnotaitonValue
		_, err := o.k8sClient.AppsV1().Deployments(namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}
	}

	return &VpaNameResponse{ResourceName: name}, nil
}

func (o *autoscalerOperator) UpdateVPA(namespace string, name string, ui_vpa *ModifyVpaRequest) (*VpaResponse, error) {
	klog.V(2).Infof("Updating VPA for deployment: \"%s\" in \"%s\" namespace", name, namespace)
	// Get deployment, check if VPA is set
	deployment, err := o.k8sClient.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else if value, ok := deployment.Annotations[annotationKey]; !ok {
		// No VPA on target deployment
		return nil, fmt.Errorf("target deployment \"%s\" in \"%s\" namespace has no VPA, abort", name, namespace)
	} else if value != vpaAnnotaitonValue {
		// Other autoscaler on target deployment
		return nil, fmt.Errorf("target deployment \"%s\" in \"%s\" namespace has other autoscaler, abort", name, namespace)
	}

	// Get VPA
	path := fmt.Sprintf(vpaPath, namespace) + "/" + name
	bytes, err := o.restClient.Get().AbsPath(path).DoRaw(context.Background())
	if err != nil {
		return nil, err
	}
	oldVPA := &vpav1.VerticalPodAutoscaler{}
	err = json.Unmarshal(bytes, oldVPA)
	if err != nil {
		return nil, err
	}
	newVPA := oldVPA.DeepCopy()

	// Format VPA
	if ui_vpa.UpdateMode != nil {
		newVPA.Spec.UpdatePolicy.UpdateMode = (*vpav1.UpdateMode)(ui_vpa.UpdateMode)
	}
	if ui_vpa.MinReplicas != nil {
		newVPA.Spec.UpdatePolicy.MinReplicas = ui_vpa.MinReplicas
	}
	if ui_vpa.CPolicies != nil {
		newVPA.Spec.ResourcePolicy.ContainerPolicies = createVpaContainerPolicies(ui_vpa.CPolicies)
	}
	if reflect.DeepEqual(oldVPA.Spec, newVPA.Spec) {
		// No update, abort
		return makeVpaResponse(newVPA), nil
	}

	// Update VPA
	bytes, err = json.Marshal(newVPA)
	if err != nil {
		return nil, err
	}
	_, err = o.restClient.Put().AbsPath(path).Body(bytes).DoRaw(context.Background())
	if err != nil {
		return nil, err
	}
	
	return makeVpaResponse(newVPA), nil
}

func (o *autoscalerOperator) GetVPA(namespace string, name string) (*VpaResponse, error) {
	klog.V(2).Infof("Getting VPA for deployment: \"%s\" in \"%s\" namespace", name, namespace)

	// Get VPA
	path := fmt.Sprintf(vpaPath, namespace) + "/" + name
	bytes, err := o.restClient.Get().AbsPath(path).DoRaw(context.Background())
	if err != nil {
		return nil, err
	}
	vpa := &vpav1.VerticalPodAutoscaler{}
	err = json.Unmarshal(bytes, vpa)
	if err != nil {
		return nil, err
	}

	return makeVpaResponse(vpa), nil
}

func (o *autoscalerOperator) DeleteVPA(namespace string, name string) error {
	klog.V(2).Infof("Deleting VPA for deployment: \"%s\" in \"%s\" namespace", name, namespace)

	// Delete VPA
	path := fmt.Sprintf(vpaPath, namespace) + "/" + name
	_, err := o.restClient.Delete().AbsPath(path).DoRaw(context.Background())
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	// Get deployment, remove the annotation
	deployment, err := o.k8sClient.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return err
	} else {
		delete(deployment.Annotations, annotationKey)
		_, err := o.k8sClient.AppsV1().Deployments(namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func createVpaContainerPolicies(policies []CPolicy) []vpav1.ContainerResourcePolicy {
	returnArray := make([]vpav1.ContainerResourcePolicy, 0)
	for _, policy := range policies {
		returnArray = append(returnArray, vpav1.ContainerResourcePolicy{
			ContainerName: policy.ContainerName,
			Mode: (*vpav1.ContainerScalingMode)(&policy.Mode),
			MinAllowed: createVpaResourceList(policy.MinAllowed),
			MaxAllowed: createVpaResourceList(policy.MaxAllowed),
			ControlledResources: createVpaCtrledResources(policy.CtrledResources),
			ControlledValues: (*vpav1.ContainerControlledValues)(&policy.CtrledValues),
		})
	}
	return returnArray
}

func createVpaResourceList(resources *Resources) corev1.ResourceList {
	returnMap := make(map[corev1.ResourceName]resource.Quantity, 0)
	if resources.Cpu != 0.0 {
		returnMap[corev1.ResourceCPU] = resource.MustParse(fmt.Sprint(resources.Cpu))
	}
	if resources.Memory != 0.0 {
		memory := math.Round(float64(resources.Memory) * unitFactor) // from MiBs to Bs, and round to whole number
		returnMap[corev1.ResourceMemory] = resource.MustParse(fmt.Sprintf("%.0f", memory)) // Bs, make sure no .xx
	}
	return returnMap
}

func createVpaCtrledResources(ctrledResources []string) *[]corev1.ResourceName {
	returnArray := make([]corev1.ResourceName, 0)
	for _, resource := range ctrledResources {
		if resource == "cpu" {
			returnArray = append(returnArray, corev1.ResourceCPU)
		} else if resource == "memory" {
			returnArray = append(returnArray, corev1.ResourceMemory)
		}
	}
	return &returnArray
}

func makeVpaResponse(vpa *vpav1.VerticalPodAutoscaler) *VpaResponse {
	return &VpaResponse{
		ResourceName: vpa.Name,
		UpdateMode: string(*vpa.Spec.UpdatePolicy.UpdateMode),
		MinReplicas: *vpa.Spec.UpdatePolicy.MinReplicas,
		CPolicies: makeVpaContainerPoliciesResponse(vpa.Spec.ResourcePolicy.ContainerPolicies),
	}
}

func makeVpaContainerPoliciesResponse(policies []vpav1.ContainerResourcePolicy) []CPolicy {
	returnArray := make([]CPolicy, 0)
	for _, policy := range policies {
		returnArray = append(returnArray, CPolicy{
			ContainerName: policy.ContainerName,
			Mode: string(*policy.Mode),
			MinAllowed: makeVpaResourcesResponse(&policy.MinAllowed),
			MaxAllowed: makeVpaResourcesResponse(&policy.MaxAllowed),
			CtrledResources: makeVpaCtrledResourcesResponse(policy.ControlledResources),
			CtrledValues: string(*policy.ControlledValues),
		})
	}
	return returnArray
}

func makeVpaResourcesResponse(list *corev1.ResourceList) *Resources {
	return &Resources{
		Cpu: float32(list.Cpu().AsApproximateFloat64()), // Unit: cores, round to .xx
		Memory: float32(math.Round(float64(list.Memory().Value()) / unitFactor * 100 ) / 100), // Unit: MiBs, round to .xx
	}
}

func makeVpaCtrledResourcesResponse(ctrledResources *[]corev1.ResourceName) []string {
	returnArray := make([]string, 0)
	for _, resource := range *ctrledResources {
		returnArray = append(returnArray, resource.String())
	}
	return returnArray
}
