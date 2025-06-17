/*
Copyright(c) 2025-present Accton. All rights reserved. www.accton.com
*/

package autoscaler

//import (
//	"k8s.io/api/autoscaling/v2beta2"
//)

// HPA

type HpaRequest struct {
	ResourceName      string       `json:"resourceName" description:"HPA name, use the name of the target deployment. Must be unique. Valid characters: a-z, 0-9, and -(hyphen). And must start and end with an alphanumeric character." maximum:"253"`
	TargetCpuUsage    int32        `json:"targetCpuUsage,omitempty" default:"80" description:"Target CPU usage, in percentages. At least one of CPU or memory taget usage must be set." minimum:"1" maximum:"100"`
	TargetMemoryUsage float32      `json:"targetMemoryUsage,omitempty" description:"Target memory usage, in MiBs. At least one of CPU or memory taget usage must be set." minimum:"0.01" maximum:"10000"`
	MinReplicas       int32        `json:"minReplicas" default:"1" description:"Set the minimum number of pod replicas allowed." minimum:"1" maximum:"2147483647"`
	MaxReplicas       int32        `json:"maxReplicas" default:"1" description:"Set the maximum number of pod replicas allowed. Cannot be set lower than the minimum replicas." minimum:"1" maximum:"2147483647"`
	ScaleUp           scalingRules `json:"scaleUp" description:"Scale up object, to configure the scaling up behavior."`
	ScaleDown         scalingRules `json:"scaleDown" description:"Scale down object, to configure the scaling down behavior."`
}

type scalingRules struct {
	StableWindow int32           `json:"stabilizationWindowSeconds" description:"The number of seconds for which past recommendations should be considered while scaling up. Unit is second. Default: 0 when scaling up; 300 when scaling down." minimum:"0" maximum:"3600"`
	SelectPolicy string          `json:"selectPolicy" default:"Max" description:"To specify which policy should be used. Available input: Max, Min, Disabled"`
	Policies     []scalingPolicy `json:"policies" description:"A list of potential scaling policies which can be used during scaling. At least one item must be set, and each type can be set at most once."`
}

type scalingPolicy struct {
	Type          string `json:"type" description:"Type is used to specify the scaling policy. Unique key. Available input: Pods, Percent"`
	Value         int32  `json:"value" description:"Value contains the amount of change which is permitted by the policy. In numbers or percentages." minimum:"1" maximum:"2147483647"`
	PeriodSeconds int32  `json:"periodSeconds" description:"PeriodSeconds specifies the window of time for which the policy should hold true." minimum:"1" maximum:"1800"`
}

type HpaNameResponse struct {
	ResourceName string `json:"resourceName" description:"HPA name, and the name of the target deployment."`
}

type ModifyHpaRequest struct {
	TargetCpuUsage    *int32        `json:"targetCpuUsage,omitempty" description:"Target CPU usage, in percentages. At least one of CPU or memory taget usage must be set." minimum:"1" maximum:"100"`
	TargetMemoryUsage *float32      `json:"targetMemoryUsage,omitempty" description:"Target memory usage, in MiBs. At least one of CPU or memory taget usage must be set." minimum:"0.01" maximum:"10000"`
	MinReplicas       *int32        `json:"minReplicas,omitempty" description:"Set the minimum number of pod replicas allowed." minimum:"1" maximum:"2147483647"`
	MaxReplicas       *int32        `json:"maxReplicas,omitempty" description:"Set the maximum number of pod replicas allowed. Cannot be set lower than the minimum replicas." minimum:"1" maximum:"2147483647"`
	ScaleUp           *scalingRules `json:"scaleUp,omitempty" description:"Scale up object, to configure the scaling up behavior."`
	ScaleDown         *scalingRules `json:"scaleDown,omitempty" description:"Scale down object, to configure the scaling down behavior."`
}

type HpaResponse struct {
	ResourceName      string       `json:"resourceName" description:"HPA name, and the name of the target deployment. Unique key."`
	TargetCpuUsage    int32        `json:"targetCpuUsage" description:"Target CPU usage, in percentages."`
	TargetMemoryUsage float32      `json:"targetMemoryUsage" description:"Target memory usage, in MiBs."`
	MinReplicas       int32        `json:"minReplicas" description:"Minimum number of pod replicas allowed."`
	MaxReplicas       int32        `json:"maxReplicas" description:"Maxnimum number of pod replicas allowed."`
	ScaleUp           scalingRules `json:"scaleUp" description:"Scale up object, to configure the scaling up behavior."`
	ScaleDown         scalingRules `json:"scaleDown" description:"Scale down object, to configure the scaling down behavior."`
}

type ListRepositoryResponse struct {
	TotalCount int           `json:"total_count" description:"Total number of HPAs."`
	Items      []HpaResponse `json:"items" description:"List of HPAs. Key is items[].resourceName"`
}


// VPA

type VpaRequest struct {
	ResourceName string    `json:"resourceName" description:"VPA name, use the name of the target deployment. Must be unique. Valid characters: a-z, 0-9, and -(hyphen). And must start and end with an alphanumeric character." maximum:"253"`
	UpdateMode   string    `json:"updateMode" default:"Auto" description:"Controls the way autoscaler applies changes to the pod resources. 'Initial' mode will only apply changes on pod creation. 'Auto' mode not only applies changes on pod creation but also will try to delete and recreate pod to apply changes. 'Off' mode will turn off updating function. Available input: Initial, Auto, Off"`
	MinReplicas  int32     `json:"minReplicas" default:"2" description:"Minimum number of replicas needed to update pod resources. If set to 1, service provided by pod may be interrupted due to pod recreation." minimum:"1" maximum:"2147483647"`
	CPolicies    []cPolicy `json:"containerPolicies" description:"Container resource policy array. This array may have multiple policies; each policy corresponds to 1 unique container name. A wildcard policy must be specified in the array when creating VPA."`
}

type cPolicy struct {
	ContainerName   string    `json:"containerName" description:"Container name for this policy. Unique key." maximum:"64"`
	Mode            string    `json:"mode" default:"Auto" description:"Whether autoscaler is enabled in this policy. Available input: Auto, Off"`
	MinAllowed      resources `json:"minAllowed" description:"Specify the minimum resources can be assigned to container. Can be an empty object."`
	MaxAllowed      resources `json:"maxAllowed" description:"Specify the maximum resources can be assigned to container. Can be an empty object."`
	CtrledResources []string  `json:"controlledResources" description:"Array of strings, must not be empty. Specifies which resource types will be controlled by autoscaler. Available input: cpu, memory"`
	CrtledValues    string    `json:"controlledValues" default:"RequestsAndLimits" description:"Specifies which resource values should be controlled. Available input: RequestsAndLimits, RequestsOnly"`
}

type resources struct {
	Cpu    float32 `json:"cpu,omitempty" description:"Minimum/Maximum amount of cpu resource that can be assigned to the container, unit is cores. If not set, autoscaler will assign as low/high as possible resource to the container." minimum:"0.01" maximum:"10000"`
	Memory float32 `json:"memory,omitempty" description:"Minimum/Maximum amount of memory resources that can be assigned to the container unit is MiBs. If not set, autoscaler will assign as low/high as possible resource to the container." minimum:"0.01" maximum:"10000"`
}

type VpaNameResponse struct {
	ResourceName string `json:"resourceName" description:"VPA name, and the name of the target deployment."`
}

type ModifyVpaRequest struct {
	UpdateMode  *string   `json:"updateMode,omitempty" description:"Controls the way autoscaler applies changes to the pod resources. 'Initial' mode will only apply changes on pod creation. 'Auto' mode not only applies changes on pod creation but also will try to delete and recreate pod to apply changes. 'Off' mode will turn off updating function. Available input: Initial, Auto, Off"`
	MinReplicas *int32    `json:"minReplicas,omitempty" description:"Minimum number of replicas needed to update pod resources. If set to 1, service provided by pod may be interrupted due to pod recreation." minimum:"1" maximum:"2147483647"`
	CPolicies   []cPolicy `json:"containerPolicies,omitempty" description:"Container resource policy array. This array may have multiple policies; each policy corresponds to 1 unique container name. If this array is provided, must not remove the wildcard policy in the array when updating VPA."`
}

type VpaResponse struct {
	ResourceName string    `json:"resourceName" description:"VPA name, and the name of the target deployment."`
	UpdateMode   string    `json:"updateMode" description:"Controls the way autoscaler applies changes to the pod resources. 'Initial' mode will only apply changes on pod creation. 'Auto' mode not only applies changes on pod creation but also will try to delete and recreate pod to apply changes. 'Off' mode will turn off updating function."`
	MinReplicas  int32     `json:"minReplicas" description:"Minimum number of replicas needed to update pod resources. If set to 1, service provided by pod may be interrupted due to pod recreation."`
	CPolicies    []cPolicy `json:"containerPolicies" description:"Container resource policy array. This array may have multiple policies; each policy corresponds to 1 unique container name. Must have at least one wildcard policy in the array when creating VPA."`
}

type ListVpaResponse struct {
	TotalCount int           `json:"total_count" description:"Total number of VPAs."`
	Items      []VpaResponse `json:"items" description:"List of VPAs. Key is items[].resourceName"`
}
