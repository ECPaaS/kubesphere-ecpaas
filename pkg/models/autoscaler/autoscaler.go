/*
Copyright(c) 2025-present Accton. All rights reserved. www.accton.com
*/

package autoscaler

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	autoscaling "k8s.io/api/autoscaling/v2beta2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	//repositoiesPath = "/apis/velero.io/v1/namespaces/velero/backupstoragelocations"
	//backupsPath = "/apis/velero.io/v1/namespaces/velero/backups"
	//deleteBackupFilePath = "/apis/velero.io/v1/namespaces/velero/deletebackuprequests"
	//restoresPath = "/apis/velero.io/v1/namespaces/velero/restores"
)

var trueFlag bool = true
//var falseFlag bool = false


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
	ksclient  kubesphere.Interface
	k8sclient kubernetes.Interface
	restclient rest.Interface
}

func New(ksclient kubesphere.Interface, k8sclient kubernetes.Interface) Interface {
	var restclient rest.Interface

	if k8sclient != nil {
		restclient = k8sclient.AppsV1().RESTClient()
	}

	return &autoscalerOperator{
		ksclient:   ksclient,
		k8sclient:  k8sclient,
		restclient: restclient,
	}
}

// HPA

func (o *autoscalerOperator) CreateHPA(namespace string, name string, ui_hpa *HpaRequest) (*HpaNameResponse, error) {
	klog.V(2).Infof("Creating HPA for deployment: \"%s\" in \"%s\" namespace", name, namespace)
	// Get deployment, check if any autoscaler is set
	deployment, err := o.k8sclient.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else if _, ok := deployment.Annotations[annotationKey]; ok {
		// Already has autoscaler on deployment
		return nil,  fmt.Errorf("target deployment \"%s\" in \"%s\" namespace already has autoscaler, abort", name, namespace)
	}

	// Format HPA
	hpa := &autoscaling.HorizontalPodAutoscaler{}
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
	hpa.Spec.ScaleTargetRef = autoscaling.CrossVersionObjectReference{
		APIVersion: "apps/v1",
		Kind: "Deployment",
		Name: name,
	}
	hpa.Spec.Metrics = createHpaMetrics(ui_hpa.TargetCpuUsage, ui_hpa.TargetMemoryUsage)
	hpa.Spec.MinReplicas = &ui_hpa.MinReplicas
	hpa.Spec.MaxReplicas = ui_hpa.MaxReplicas
	hpa.Spec.Behavior = createHpaBehaviors(ui_hpa.ScaleUp, ui_hpa.ScaleDown)

	// Create HPA
	_, err = o.k8sclient.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Create(context.Background(), hpa, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	} else {
		// PUT deploment annotations
		deployment.Annotations[annotationKey] = hpaAnnotaitonValue
		_, err := o.k8sclient.AppsV1().Deployments(namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}
	}

	return &HpaNameResponse{ResourceName: name}, nil
}

func (o *autoscalerOperator) UpdateHPA(namespace string, name string, ui_hpa *ModifyHpaRequest) (*HpaResponse, error) {
	klog.V(2).Infof("Updating HPA for deployment: \"%s\" in \"%s\" namespace", name, namespace)
	// Get deployment, check if HPA is set
	deployment, err := o.k8sclient.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
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
	oldHPA, err := o.k8sclient.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	newHPA := oldHPA.DeepCopy()

	// Format HPA
	metrics := make([]autoscaling.MetricSpec, 0)
	if ui_hpa.TargetCpuUsage != nil {
		metrics = append(metrics, createHpaMetrics(*ui_hpa.TargetCpuUsage, 0.0)...)
		newHPA.Annotations[cpuTargetUtilization] = fmt.Sprint(*ui_hpa.TargetCpuUsage)
	}
	if ui_hpa.TargetMemoryUsage != nil {
		metrics = append(metrics, createHpaMetrics(0, *ui_hpa.TargetMemoryUsage)...)
		newHPA.Annotations[memoryTargetValue] = fmt.Sprint(*ui_hpa.TargetMemoryUsage) + "Mi"
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
		// No change, abort
		return makeHpaResponse(newHPA), nil // No Update
	}

	// Update HPA
	updatedHPA, err := o.k8sclient.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Update(context.Background(), newHPA, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	
	return makeHpaResponse(updatedHPA), nil
}

func (o *autoscalerOperator) GetHPA(namespace string, name string) (*HpaResponse, error) {
	klog.V(2).Infof("Getting HPA for deployment: \"%s\" in \"%s\" namespace", name, namespace)

	// Get HPA
	hpa, err := o.k8sclient.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return makeHpaResponse(hpa), nil
}

func (o *autoscalerOperator) DeleteHPA(namespace string, name string) error {
	klog.V(2).Infof("Deleting HPA for deployment: \"%s\" in \"%s\" namespace", name, namespace)

	// Delete HPA
	err := o.k8sclient.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}

func createHpaMetrics(cpu int32, memory float32) []autoscaling.MetricSpec {
	returnArray := make([]autoscaling.MetricSpec, 0)
	if cpu != 0 {
		returnArray = append(returnArray, autoscaling.MetricSpec{
			Type: "Resources",
			Resource: &autoscaling.ResourceMetricSource{
				Name: "cpu",
				Target: autoscaling.MetricTarget{
					Type: "Utilization",
					AverageUtilization: &cpu,
				},
			},
		})
	}

	if memory != 0.0 {
		quantity := resource.MustParse(fmt.Sprint(memory) + "Mi")
		returnArray = append(returnArray, autoscaling.MetricSpec{
			Type: "Resources",
			Resource: &autoscaling.ResourceMetricSource{
				Name: "memory",
				Target: autoscaling.MetricTarget{
					Type: "AverageValue",
					AverageValue: &quantity,
				},
			},
		})
	}

	return returnArray
}

func createHpaBehaviors(up *ScalingRules, down *ScalingRules) *autoscaling.HorizontalPodAutoscalerBehavior {
	returnBehaviors := &autoscaling.HorizontalPodAutoscalerBehavior{}
	if up != nil {
		returnBehaviors.ScaleUp = createHpaScalingRules(up)
	}

	if down != nil {
		returnBehaviors.ScaleDown = createHpaScalingRules(down)
	}

	return returnBehaviors
}

func createHpaScalingRules(rules *ScalingRules) *autoscaling.HPAScalingRules {
	return &autoscaling.HPAScalingRules{
		StabilizationWindowSeconds: rules.StableWindow,
		SelectPolicy: (*autoscaling.ScalingPolicySelect)(&rules.SelectPolicy),
		Policies: createHpaScalingPolicies(rules.Policies),
	}
}

func createHpaScalingPolicies(policies []ScalingPolicy) []autoscaling.HPAScalingPolicy {
	returnPolicies := make([]autoscaling.HPAScalingPolicy, 0)
	for _, policy := range policies {
		returnPolicies = append(returnPolicies, autoscaling.HPAScalingPolicy{
			Type: autoscaling.HPAScalingPolicyType(policy.Type),
			Value: policy.Value,
			PeriodSeconds: policy.PeriodSeconds,
		})
	}
	return returnPolicies
}

func makeHpaResponse(hpa *autoscaling.HorizontalPodAutoscaler) *HpaResponse {
	cpuUsage, _ := strconv.ParseInt(hpa.Annotations[cpuTargetUtilization], 10, 32)
	memUsage, _ := strconv.ParseFloat(hpa.Annotations[memoryTargetValue], 32)
	return &HpaResponse{
		ResourceName: hpa.Name,
		TargetCpuUsage: int32(cpuUsage),
		TargetMemoryUsage: float32(memUsage),
		MinReplicas: *hpa.Spec.MinReplicas,
		MaxReplicas: hpa.Spec.MaxReplicas,
		ScaleUp: ScalingRules{
			StableWindow: hpa.Spec.Behavior.ScaleUp.StabilizationWindowSeconds,
			SelectPolicy: string(*hpa.Spec.Behavior.ScaleUp.SelectPolicy),
			Policies: makeHpaPolicies(hpa.Spec.Behavior.ScaleUp.Policies),
		},
		ScaleDown: ScalingRules{
			StableWindow: hpa.Spec.Behavior.ScaleDown.StabilizationWindowSeconds,
			SelectPolicy: string(*hpa.Spec.Behavior.ScaleDown.SelectPolicy),
			Policies: makeHpaPolicies(hpa.Spec.Behavior.ScaleDown.Policies),
		},
	}
}

func makeHpaPolicies(policies []autoscaling.HPAScalingPolicy) []ScalingPolicy {
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
	deployment, err := o.k8sclient.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else if _, ok := deployment.Annotations[annotationKey]; ok {
		// Already has autoscaler on deployment
		return nil,  fmt.Errorf("target deployment \"%s\" in \"%s\" namespace already has autoscaler, abort", name, namespace)
	}

	// Format VPA

	// Create VPA
	_, err = o.k8sclient.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Create(context.Background(), nil, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	} else {
		// PUT deploment annotations
		deployment.Annotations[annotationKey] = vpaAnnotaitonValue
		_, err := o.k8sclient.AppsV1().Deployments(namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}
	}

	return &VpaNameResponse{ResourceName: name}, nil
}

func (o *autoscalerOperator) UpdateVPA(namespace string, name string, ui_vpa *ModifyVpaRequest) (*VpaResponse, error) {
	klog.V(2).Infof("Updating VPA for deployment: \"%s\" in \"%s\" namespace", name, namespace)
	// Get deployment, check if VPA is set
	deployment, err := o.k8sclient.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
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
	oldHPA, err := o.k8sclient.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	newHPA := oldHPA.DeepCopy()

	// Format VPA
	//if ui_vpa.UpdateMode != nil { }
	//if ui_vpa.MinReplicas != nil { }
	//if ui_vpa.CPolicies != nil { }
	if reflect.DeepEqual(nil, nil) {
		// No change, abort
		return nil, nil // makeVpaResponse(newVPA), nil // No Update
	}

	// Update VPA
	_, err = o.k8sclient.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Update(context.Background(), newHPA, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	
	return nil, nil // makeVpaResponse(updatedVPA), nil
}

func (o *autoscalerOperator) GetVPA(namespace string, name string) (*VpaResponse, error) {
	klog.V(2).Infof("Getting VPA for deployment: \"%s\" in \"%s\" namespace", name, namespace)

	// Get VPA
	_, err := o.k8sclient.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return nil, nil // makeVpaResponse(vpa), nil
}

func (o *autoscalerOperator) DeleteVPA(namespace string, name string) error {
	klog.V(2).Infof("Deleting VPA for deployment: \"%s\" in \"%s\" namespace", name, namespace)

	// Delete VPA
	err := o.k8sclient.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}


