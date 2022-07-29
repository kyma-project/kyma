package v1alpha2

import (
	"encoding/json"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const DefaultingConfigKey = "defaulting-config"

type ReplicasPreset struct {
	Min int32 `json:"min,omitempty"`
	Max int32 `json:"max,omitempty"`
}

type ResourcesPreset struct {
	RequestCPU    string `json:"requestCpu,omitempty"`
	RequestMemory string `json:"requestMemory,omitempty"`
	LimitCPU      string `json:"limitCpu,omitempty"`
	LimitMemory   string `json:"limitMemory,omitempty"`
}

type FunctionReplicasDefaulting struct {
	DefaultPreset string                    `envconfig:"default=S"`
	Presets       map[string]ReplicasPreset `envconfig:"-"`
	PresetsMap    string                    `envconfig:"default={}"`
}

type FunctionResourcesDefaulting struct {
	DefaultPreset     string                     `envconfig:"default=M"`
	Presets           map[string]ResourcesPreset `envconfig:"-"`
	PresetsMap        string                     `envconfig:"default={}"`
	RuntimePresets    map[string]string          `envconfig:"-"`
	RuntimePresetsMap string                     `envconfig:"default={}"`
}

type BuildJobResourcesDefaulting struct {
	DefaultPreset string                     `envconfig:"default=normal"`
	Presets       map[string]ResourcesPreset `envconfig:"-"`
	PresetsMap    string                     `envconfig:"default={}"`
}

type FunctionDefaulting struct {
	Replicas  FunctionReplicasDefaulting
	Resources FunctionResourcesDefaulting
}

type BuildJobDefaulting struct {
	Resources BuildJobResourcesDefaulting
}

type DefaultingConfig struct {
	Function FunctionDefaulting
	BuildJob BuildJobDefaulting
	Runtime  Runtime `envconfig:"default=nodejs14"`
}

func (fn *Function) Default(config *DefaultingConfig) {
	fn.Spec.defaultReplicas(config, fn)
	fn.Spec.defaultFunctionResources(config, fn)
	fn.Spec.defaultBuildResources(config, fn)
	fn.Spec.defaultRuntime(config)
}

func (spec *FunctionSpec) defaultReplicas(config *DefaultingConfig, fn *Function) {
	defaultingConfig := config.Function.Replicas
	replicasPreset := mergeReplicasPreset(fn, defaultingConfig.Presets, defaultingConfig.DefaultPreset)

	if spec.MinReplicas == nil {
		newMin := replicasPreset.Min
		if spec.MaxReplicas != nil && *spec.MaxReplicas < newMin {
			newMin = *spec.MaxReplicas
		}

		spec.MinReplicas = &newMin
	}
	if spec.MaxReplicas == nil {
		newMax := replicasPreset.Max
		if *spec.MinReplicas > newMax {
			newMax = *spec.MinReplicas
		}

		spec.MaxReplicas = &newMax
	}
}

func (spec *FunctionSpec) defaultFunctionResources(config *DefaultingConfig, fn *Function) {
	resources := spec.ResourceConfiguration.Function.Resources
	defaultingConfig := config.Function.Resources
	resourcesPreset := mergeResourcesPreset(fn, FunctionResourcesPresetLabel, defaultingConfig.Presets, defaultingConfig.DefaultPreset, defaultingConfig.RuntimePresets)

	spec.ResourceConfiguration.Function.Resources = defaultResources(&resources, resourcesPreset.RequestMemory, resourcesPreset.RequestCPU, resourcesPreset.LimitMemory, resourcesPreset.LimitCPU)
}

func (spec *FunctionSpec) defaultBuildResources(config *DefaultingConfig, fn *Function) {
	buildResourceCfg := spec.ResourceConfiguration.Build
	// if build resources are not set by the user we don't default them.
	// However, if only a part is set or the preset label is set, we should correctly set missing defaults.
	if shouldSkipBuildResourcesDefault(fn) {
		return
	}

	defaultingConfig := config.BuildJob.Resources
	resourcesPreset := mergeResourcesPreset(fn, BuildResourcesPresetLabel, defaultingConfig.Presets, defaultingConfig.DefaultPreset, nil)
	spec.ResourceConfiguration.Build.Resources = defaultResources(&buildResourceCfg.Resources, resourcesPreset.RequestMemory, resourcesPreset.RequestCPU, resourcesPreset.LimitMemory, resourcesPreset.LimitCPU)
}

func (spec *FunctionSpec) defaultRuntime(config *DefaultingConfig) {
	if spec.Runtime == "" {
		spec.Runtime = config.Runtime
	}
}

func shouldSkipBuildResourcesDefault(fn *Function) bool {
	resourceCfg := fn.Spec.ResourceConfiguration.Build
	_, hasPresetLabel := fn.Labels[BuildResourcesPresetLabel]

	if resourceCfg.Resources.Limits == nil && resourceCfg.Resources.Requests == nil && !hasPresetLabel {
		return true
	}
	return false
}

func defaultResources(res *corev1.ResourceRequirements, requestMemory, requestCPU, limitMemory, limitCPU string) corev1.ResourceRequirements {
	copiedRes := res.DeepCopy()

	if copiedRes.Requests == nil {
		copiedRes.Requests = corev1.ResourceList{}
	}
	if copiedRes.Requests.Memory().IsZero() {
		newResource := resource.MustParse(requestMemory)
		if !copiedRes.Limits.Memory().IsZero() && copiedRes.Limits.Memory().Cmp(newResource) == -1 {
			newResource = *copiedRes.Limits.Memory()
		}

		copiedRes.Requests[corev1.ResourceMemory] = newResource
	}
	if copiedRes.Requests.Cpu().IsZero() {
		newResource := resource.MustParse(requestCPU)
		if !copiedRes.Limits.Cpu().IsZero() && copiedRes.Limits.Cpu().Cmp(newResource) == -1 {
			newResource = *copiedRes.Limits.Cpu()
		}

		copiedRes.Requests[corev1.ResourceCPU] = newResource
	}

	if copiedRes.Limits == nil {
		copiedRes.Limits = corev1.ResourceList{}
	}
	if copiedRes.Limits.Memory().IsZero() {
		newResource := resource.MustParse(limitMemory)
		if copiedRes.Requests.Memory().Cmp(newResource) == 1 {
			newResource = *copiedRes.Requests.Memory()
		}

		copiedRes.Limits[corev1.ResourceMemory] = newResource
	}
	if copiedRes.Limits.Cpu().IsZero() {
		newResource := resource.MustParse(limitCPU)
		if copiedRes.Requests.Cpu().Cmp(newResource) == 1 {
			newResource = *copiedRes.Requests.Cpu()
		}

		copiedRes.Limits[corev1.ResourceCPU] = newResource
	}

	return *copiedRes
}

func mergeReplicasPreset(fn *Function, presets map[string]ReplicasPreset, defaultPreset string) ReplicasPreset {
	replicas := ReplicasPreset{}

	preset := fn.GetLabels()[ReplicasPresetLabel]
	if preset == "" {
		return presets[defaultPreset]
	}

	replicasPreset := presets[preset]
	replicasDefaultPreset := presets[defaultPreset]

	replicas.Min = replicasPreset.Min
	if replicas.Min == 0 {
		replicas.Min = replicasDefaultPreset.Min
	}

	replicas.Max = replicasPreset.Max
	if replicas.Max == 0 {
		replicas.Max = replicasDefaultPreset.Max
	}

	return replicas
}

func mergeResourcesPreset(fn *Function, presetLabel string, presets map[string]ResourcesPreset, defaultPreset string, runtimePreset map[string]string) ResourcesPreset {
	resources := ResourcesPreset{}

	preset := fn.GetLabels()[presetLabel]
	if preset == "" {
		rtmPreset, ok := runtimePreset[string(fn.Spec.Runtime)]
		if ok {
			return presets[rtmPreset]
		}
		return presets[defaultPreset]
	}

	resourcesPreset := presets[preset]
	resourcesDefaultPreset := presets[defaultPreset]

	resources.RequestCPU = resourcesPreset.RequestCPU
	if resources.RequestCPU == "" {
		resources.RequestCPU = resourcesDefaultPreset.RequestCPU
	}

	resources.RequestMemory = resourcesPreset.RequestMemory
	if resources.RequestMemory == "" {
		resources.RequestMemory = resourcesDefaultPreset.RequestMemory
	}

	resources.LimitCPU = resourcesPreset.LimitCPU
	if resources.LimitCPU == "" {
		resources.LimitCPU = resourcesDefaultPreset.LimitCPU
	}

	resources.LimitMemory = resourcesPreset.LimitMemory
	if resources.LimitMemory == "" {
		resources.LimitMemory = resourcesDefaultPreset.LimitMemory
	}

	return resources
}

func ParseReplicasPresets(presetsMap string) (map[string]ReplicasPreset, error) {
	var presets map[string]ReplicasPreset
	if err := json.Unmarshal([]byte(presetsMap), &presets); err != nil {
		return presets, errors.Wrap(err, "while parsing resources presets")
	}
	return presets, nil
}

func ParseResourcePresets(presetsMap string) (map[string]ResourcesPreset, error) {
	var presets map[string]ResourcesPreset
	if err := json.Unmarshal([]byte(presetsMap), &presets); err != nil {
		return presets, errors.Wrap(err, "while parsing resources presets")
	}
	return presets, nil
}

func ParseRuntimePresets(presetsMap string) (map[string]string, error) {
	var presets map[string]string
	if err := json.Unmarshal([]byte(presetsMap), &presets); err != nil {
		return presets, errors.Wrap(err, "while parsing runtime presets")
	}
	return presets, nil
}
