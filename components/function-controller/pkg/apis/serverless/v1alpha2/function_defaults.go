package v1alpha2

import (
	"k8s.io/apimachinery/pkg/api/resource"

	corev1 "k8s.io/api/core/v1"
)

const DefaultingConfigKey = "defaulting-config"

type ReplicasPreset struct {
	Min int32
	Max int32
}

type ResourcesPreset struct {
	RequestCPU    string
	RequestMemory string
	LimitCPU      string
	LimitMemory   string
}

type FunctionReplicasDefaulting struct {
	DefaultPreset string
	Presets       map[string]ReplicasPreset
}

type FunctionResourcesDefaulting struct {
	DefaultPreset  string
	Presets        map[string]ResourcesPreset
	RuntimePresets map[string]string
}

type BuildJobResourcesDefaulting struct {
	DefaultPreset string
	Presets       map[string]ResourcesPreset
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
	Runtime  Runtime
}

func (fn *Function) Default(config *DefaultingConfig) {
	fn.Spec.defaultScaling(config)
	fn.Spec.defaultFunctionResources(config, fn)
	fn.Spec.defaultBuildResources(config, fn)
}

func (spec *FunctionSpec) defaultScaling(config *DefaultingConfig) {
	defaultingConfig := config.Function.Replicas
	replicasPreset := defaultingConfig.Presets[defaultingConfig.DefaultPreset]

	if spec.Replicas == nil {
		// TODO: The presets structure and docs should be updated to reflect the new behavior.
		spec.Replicas = &replicasPreset.Min
	}

	if spec.ScaleConfig == nil {
		return
	}

	// spec.ScaleConfig is SET, but not fully configured, for sanity, we will default MinReplicas, we will also use it as a default spec.Replica
	if spec.ScaleConfig.MinReplicas == nil {
		newMin := replicasPreset.Min
		if spec.ScaleConfig.MaxReplicas != nil && *spec.ScaleConfig.MaxReplicas < newMin {
			newMin = *spec.ScaleConfig.MaxReplicas
		}
		spec.ScaleConfig.MinReplicas = &newMin
	}
	spec.Replicas = spec.ScaleConfig.MinReplicas

	if spec.ScaleConfig.MaxReplicas == nil {
		newMax := replicasPreset.Max
		if *spec.ScaleConfig.MinReplicas > newMax {
			newMax = *spec.ScaleConfig.MinReplicas
		}

		spec.ScaleConfig.MaxReplicas = &newMax
	}
}

func (spec *FunctionSpec) defaultFunctionResources(config *DefaultingConfig, fn *Function) {
	var resources *corev1.ResourceRequirements
	var profile string
	if spec.ResourceConfiguration != nil && spec.ResourceConfiguration.Function != nil {
		functionResourceCfg := *spec.ResourceConfiguration.Function
		if functionResourceCfg.Resources != nil {
			resources = functionResourceCfg.Resources
		}
		profile = functionResourceCfg.Profile
	}
	defaultingConfig := config.Function.Resources
	calculatedResources := calculateResources(fn, resources, profile, FunctionResourcesPresetLabel, defaultingConfig.Presets, defaultingConfig.DefaultPreset, defaultingConfig.RuntimePresets)
	setFunctionResources(spec, calculatedResources)
}

func setFunctionResources(spec *FunctionSpec, resources *corev1.ResourceRequirements) {

	if spec.ResourceConfiguration == nil {
		spec.ResourceConfiguration = &ResourceConfiguration{}
	}

	if spec.ResourceConfiguration.Function == nil {
		spec.ResourceConfiguration.Function = &ResourceRequirements{}
	}

	spec.ResourceConfiguration.Function.Resources = resources
}

func (spec *FunctionSpec) defaultBuildResources(config *DefaultingConfig, fn *Function) {
	// if build resources are not set by the user we don't default them.
	// However, if only a part is set or the preset label is set, we should correctly set missing defaults.
	if shouldSkipBuildResourcesDefault(fn) {
		return
	}

	var buildResourceCfg ResourceRequirements
	if spec.ResourceConfiguration != nil && spec.ResourceConfiguration.Build != nil {
		buildResourceCfg = *spec.ResourceConfiguration.Build
	}

	defaultingConfig := config.BuildJob.Resources
	calculatedResources := calculateResources(fn, buildResourceCfg.Resources, buildResourceCfg.Profile, BuildResourcesPresetLabel, defaultingConfig.Presets, defaultingConfig.DefaultPreset, nil)

	setBuildResources(spec, calculatedResources)
}

func setBuildResources(spec *FunctionSpec, resources *corev1.ResourceRequirements) {

	if spec.ResourceConfiguration == nil {
		spec.ResourceConfiguration = &ResourceConfiguration{}
	}

	if spec.ResourceConfiguration.Build == nil {
		spec.ResourceConfiguration.Build = &ResourceRequirements{}
	}

	spec.ResourceConfiguration.Build.Resources = resources
}

func shouldSkipBuildResourcesDefault(fn *Function) bool {
	resourceCfg := fn.Spec.ResourceConfiguration.Build
	_, hasPresetLabel := fn.Labels[BuildResourcesPresetLabel]
	if hasPresetLabel {
		return false
	}

	if resourceCfg != nil {
		if resourceCfg.Profile != "" {
			return false
		}
		if resourceCfg.Resources != nil {
			return resourceCfg.Resources.Limits == nil && resourceCfg.Resources.Requests == nil
		}
	}
	return true
}

func calculateResources(fn *Function, resourceRequirements *corev1.ResourceRequirements, profile string, presetLabel string, presets map[string]ResourcesPreset, defaultPreset string, runtimePreset map[string]string) *corev1.ResourceRequirements {
	// profile has the highest priority
	preset := profile
	// we can use profile from label (deprecated) instead of new profile
	if preset == "" {
		preset = fn.GetLabels()[presetLabel]
	}
	if preset != "" {
		return presetsToRequirements(presets[preset])
	}
	// when no profile we use user defined resources
	if resourceRequirements != nil {
		return resourceRequirements
	}
	// we use default preset only when no profile and no resources
	rtmPreset, ok := runtimePreset[string(fn.Spec.Runtime)]
	if ok {
		return presetsToRequirements(presets[rtmPreset])
	}
	return presetsToRequirements(presets[defaultPreset])
}

func presetsToRequirements(preset ResourcesPreset) *corev1.ResourceRequirements {
	result := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(preset.LimitCPU),
			corev1.ResourceMemory: resource.MustParse(preset.LimitMemory),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(preset.RequestCPU),
			corev1.ResourceMemory: resource.MustParse(preset.RequestMemory),
		},
	}
	return &result
}
