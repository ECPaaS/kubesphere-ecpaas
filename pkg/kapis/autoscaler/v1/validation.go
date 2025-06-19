/*
Copyright(c) 2025-present Accton. All rights reserved. www.accton.com
*/

package v1

import (
	"math"
	"net/http"
	"reflect"

	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/kapis/util"
	ui_autoscaler "kubesphere.io/kubesphere/pkg/models/autoscaler"
)

const (
	resourceUpperLimit = 10000.0
	resourceLowerLimit = 0.01
	percentUpperLimit = 100
	percentLowerLimit = 1
)

// HPA

func isValidHpaRequest(request *ui_autoscaler.HpaRequest, resp *restful.Response) bool {
	reflectType := reflect.TypeOf(*request)
	// ResourcName string
	if request.ResourceName == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "resourceName must not be empty.",
		})
		return false
	} else if !util.IsValidKubernetesString(request.ResourceName, "resourceName", resp) {
		return false
	}
	// TargetCpuUsage int32
	if request.TargetCpuUsage == 0 {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "targetCpuUsage must not be empty.",
		})
		return false
	} else if request.TargetCpuUsage > percentUpperLimit || request.TargetCpuUsage < percentLowerLimit {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "targetCpuUsage must be whole number percentage between 1 ~ 100",
		})
		return false
	}
	// TargetMemoryUsage float32
	if request.TargetMemoryUsage == 0.0 {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "targetMemoryUsage must not be empty.",
		})
		return false
	} else if request.TargetMemoryUsage > resourceUpperLimit || request.TargetMemoryUsage < resourceLowerLimit {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "targetMemoryUsage must be between 0.01 ~ 10000, and will be rounded to 2 decimal places",
		})
		return false
	} else {
		request.TargetMemoryUsage = float32(math.Round(float64(request.TargetMemoryUsage) * 100) / 100)
	}
	// MinReplicas int32
	if request.MinReplicas == 0 {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "minReplicas must not be empty.",
		})
		return false
	} else if !util.IsValidWithinRange(reflectType, int(request.MinReplicas), "minReplicas", resp) {
		return false
	}
	// MaxReplicas int32
	if request.MaxReplicas == 0 {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "maxReplicas must not be empty.",
		})
		return false
	} else if !util.IsValidWithinRange(reflectType, int(request.MaxReplicas), "maxReplicas", resp) {
		return false
	}
	// ScaleUp *scalingRules
	if request.ScaleUp == nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "scaleUp must not be empty.",
		})
		return false
	} else if !isValidScalingRules(request.ScaleUp, resp, true) {
		return false
	}
	// ScaleDown *scalingRules
	if request.ScaleDown == nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "scaleDown must not be empty.",
		})
		return false
	} else if !isValidScalingRules(request.ScaleDown, resp, true) {
		return false
	}
	return true
}

func isValidHpaModifyRequest(request *ui_autoscaler.ModifyHpaRequest, resp *restful.Response) bool {
	reflectType := reflect.TypeOf(*request)
	// TargetCpuUsage *int32
	if request.TargetCpuUsage != nil {
		if *request.TargetCpuUsage > percentUpperLimit || *request.TargetCpuUsage < percentLowerLimit {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
				Reason: "targetCpuUsage must be whole number percentage between 1 ~ 100",
			})
			return false
		}
	}
	// TargetMemoryUsage *float32
	if request.TargetMemoryUsage != nil {
		if *request.TargetMemoryUsage > resourceUpperLimit || *request.TargetMemoryUsage < resourceLowerLimit {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
				Reason: "targetMemoryUsage must be between 0.01 ~ 10000, and will be rounded to 2 decimal places",
			})
			return false
		} else {
			value := float32(math.Round(float64(*request.TargetMemoryUsage) * 100) / 100)
			request.TargetMemoryUsage = &value
		}
	}
	// MinReplicas *int32
	if request.MinReplicas != nil && !util.IsValidWithinRange(reflectType, int(*request.MinReplicas), "minReplicas", resp) {
		return false
	}
	// MaxReplicas *int32
	if request.MaxReplicas != nil && !util.IsValidWithinRange(reflectType, int(*request.MaxReplicas), "maxReplicas", resp) {
		return false
	}
	// ScaleUp *scalingRules
	if request.ScaleUp != nil && !isValidScalingRules(request.ScaleUp, resp, false) {
		return false
	}
	// ScaleDown *scalingRules
	if request.ScaleDown != nil && !isValidScalingRules(request.ScaleDown, resp, false) {
		return false
	}
	return true
}

func isValidScalingRules(rules *ui_autoscaler.ScalingRules, resp *restful.Response, isNew bool) bool {
	reflectType := reflect.TypeOf(*rules)
	// StableWindow *int32
	if rules.StableWindow == nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "stabilizationWindowSeconds must not be empty.",
		})
		return false
	} else if !util.IsValidWithinRange(reflectType, int(*rules.StableWindow), "stabilizationWindowSeconds", resp) {
		return false
	}
	// SelectPolicy string
	if rules.SelectPolicy == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "selectPolicy must not be empty.",
		})
		return false
	} else if rules.SelectPolicy != "Max" && rules.SelectPolicy != "Min" && rules.SelectPolicy != "Disabled" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "selectPolicy must be one of [\"Max\",\"Min\",\"Disabled\"]",
		})
		return false
	}
	// Policies []ScalingPolicy
	if rules.Policies == nil && isNew {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "policies must not be empty.",
		})
		return false
	} else if !isValidScalingPolicies(rules.Policies, resp, isNew) {
		return false
	}
	return true
}

func isValidScalingPolicies(policies []ui_autoscaler.ScalingPolicy, resp *restful.Response, isNew bool) bool {
	if len(policies) == 0 && isNew {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "At least one item must be provided in policies.",
		})
		return false
	} else if len(policies) > 2 {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "Each type of policy can be set at most once in policies.",
		})
		return false
	}
	for _, policy := range policies {
		reflectType := reflect.TypeOf(policy)
		// Type string
		if policy.Type == "" {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
				Reason: "type must not be empty.",
			})
			return false
		} else if policy.Type != "Pods" && policy.Type != "Percent" {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
				Reason: "type must be one of [\"Pods\",\"Percent\"]",
			})
			return false
		}
		// Value int32
		if policy.Value == 0 {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
				Reason: "value must not be empty.",
			})
			return false
		} else if !util.IsValidWithinRange(reflectType, int(policy.Value), "value", resp) {
			return false
		}
		// PeriodSeconds int32
		if policy.PeriodSeconds == 0 {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
				Reason: "periodSeconds must not be empty.",
			})
			return false
		} else if !util.IsValidWithinRange(reflectType, int(policy.PeriodSeconds), "periodSeconds", resp) {
			return false
		}
	}
	return true
}

// VPA

func isValidVpaRequest(request *ui_autoscaler.VpaRequest, resp *restful.Response) bool {
	reflectType := reflect.TypeOf(*request)
	// ResourcName string
	if request.ResourceName == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "resourceName must not be empty.",
		})
		return false
	} else if !util.IsValidKubernetesString(request.ResourceName, "resourceName", resp) {
		return false
	}
	// UpdateMode string
	if request.UpdateMode == "" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "updateMode must not be empty.",
		})
		return false
	} else if request.UpdateMode != "Auto" && request.UpdateMode != "Initial" && request.UpdateMode != "Off" {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "updateMode must be one of [\"Auto\",\"Initial\",\"Off\"]",
		})
		return false
	}
	// MinReplicas int32
	if request.MinReplicas == 0 {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "minReplicas must not be empty.",
		})
		return false
	} else if !util.IsValidWithinRange(reflectType, int(request.MinReplicas), "minReplicas", resp) {
		return false
	}
	// CPolicies []cPolicy
	if request.CPolicies == nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "containerPolicies must not be empty.",
		})
		return false
	} else if !isValidContainerPolicies(request.CPolicies, resp) {
		return false
	}
	return true
}

func isValidVpaModifyRequest(request *ui_autoscaler.ModifyVpaRequest, resp *restful.Response) bool {
	reflectType := reflect.TypeOf(*request)
	// UpdateMode *string
	if request.UpdateMode != nil {
		if *request.UpdateMode != "Auto" && *request.UpdateMode != "Initial" && *request.UpdateMode != "Off" {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
				Reason: "updateMode must be one of [\"Auto\",\"Initial\",\"Off\"]",
			})
			return false
		}
	}
	// MinReplicas *int32
	if request.MinReplicas != nil && !util.IsValidWithinRange(reflectType, int(*request.MinReplicas), "minReplicas", resp) {
		return false
	}
	// CPolicies []CPolicy
	if request.CPolicies != nil && !isValidContainerPolicies(request.CPolicies, resp){
		return false
	}
	return true
}

func isValidContainerPolicies(policies []ui_autoscaler.CPolicy, resp *restful.Response) bool {
	if len(policies) == 0 {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "Wildcard container policy must be provided in containerPolicies.",
		})
		return false
	}
	nameMap := make(map[string]bool, 0)
	for _, policy := range policies {
		// ContainerName string
		if policy.ContainerName == "" {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
				Reason: "containerName must not be empty.",
			})
			return false
		} else if policy.ContainerName != "*" && !util.IsValidKubernetesString(policy.ContainerName, "resourceName", resp) {
			return false
		}
		if _, ok := nameMap[policy.ContainerName]; ok {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
				Reason: "containerName must be unique.",
			})
			return false
		}
		nameMap[policy.ContainerName] = true
		// Mode string
		if policy.Mode == "" {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
				Reason: "mode must not be empty.",
			})
			return false
		} else if policy.Mode != "Auto" && policy.Mode != "Off" {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
				Reason: "mode must be one of [\"Auto\",\"Off\"]",
			})
			return false
		}
		// MinAllowed *Resources
		if policy.MinAllowed != nil && !isValidAllowedResources(policy.MinAllowed, "minAllowed", resp) {
			return false
		}
		// MaxAllowed *Resources
		if policy.MaxAllowed != nil && !isValidAllowedResources(policy.MaxAllowed, "maxAllowed", resp) {
			return false
		}
		// CtrledResources []string
		if policy.CtrledResources == nil || len(policy.CtrledResources) == 0 {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
				Reason: "controlledResources must not be empty.",
			})
			return false
		}
		for _, item := range policy.CtrledResources {
			if item != "cpu" && item != "memory" {
				resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
					Reason: "controlledResources must contain only \"cpu\" and/or \"memory\"",
				})
				return false
			}
		}
		// CtrledValues string
		if policy.CtrledValues == "" {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
				Reason: "controlledValues must not be empty.",
			})
			return false
		} else if policy.CtrledValues != "RequestsAndLimits" && policy.CtrledValues != "RequestsOnly" {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
				Reason: "controlledValues must be one of [\"RequestsAndLimits\",\"RequestsOnly\"]",
			})
			return false
		}
	}

	if _, ok := nameMap["*"]; !ok {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: "Wildcard container policy must be provided in containerPolicies.",
		})
		return false
	}
	return true
}

func isValidAllowedResources(resources *ui_autoscaler.Resources, fieldName string, resp *restful.Response) bool {
	// cpu float32
	if resources.Cpu == 0.0 {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: fieldName + " cpu resource must not be empty.",
		})
		return false
	} else if  resources.Cpu > resourceUpperLimit || resources.Cpu < resourceLowerLimit {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: fieldName + " cpu resource must be between 0.01 ~ 10000, and will be rounded to 2 decimal places",
		})
		return false
	} else {
		resources.Cpu = float32(math.Round(float64(resources.Cpu) * 100) / 100)
	}
	// memory float32
	if resources.Memory == 0.0 {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: fieldName + " memory resource must not be empty.",
		})
		return false
	} else if  resources.Memory > resourceUpperLimit || resources.Memory < resourceLowerLimit {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, util.BadRequestError{
			Reason: fieldName + " memory resource must be between 0.01 ~ 10000, and will be rounded to 2 decimal places",
		})
		return false
	} else {
		resources.Memory = float32(math.Round(float64(resources.Memory) * 100) / 100)
	}
	return true
}
