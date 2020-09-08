package v1alpha1

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"knative.dev/pkg/webhook/resourcesemantics"
)

var _ resourcesemantics.GenericCRD = (*Function)(nil)

const DefaultingConfigKey = "defaulting-config"

type FunctionDefaultResources struct {
	RequestCpu    string `envconfig:"default=50m"`
	RequestMemory string `envconfig:"default=64Mi"`
	LimitsCpu     string `envconfig:"default=100m"`
	LimitsMemory  string `envconfig:"default=128Mi"`
}

type BuildDefaultResources struct {
	RequestCpu    string `envconfig:"default=50m"`
	RequestMemory string `envconfig:"default=64Mi"`
	LimitsCpu     string `envconfig:"default=100m"`
	LimitsMemory  string `envconfig:"default=128Mi"`
}

type DefaultingConfig struct {
	Function    FunctionDefaultResources
	Build       BuildDefaultResources
	MinReplicas int32   `envconfig:"default=1"`
	MaxReplicas int32   `envconfig:"default=1"`
	Runtime     Runtime `envconfig:"default=nodejs12"`
}

func (fn *Function) SetDefaults(ctx context.Context) {
	config := ctx.Value(DefaultingConfigKey).(DefaultingConfig)

	fn.Spec.defaultReplicas(ctx)
	fn.Spec.defaultFunctionResources(ctx)
	fn.Spec.defaultBuildResources(ctx)
	fn.Spec.defaultRuntime(config)
}

func (spec *FunctionSpec) defaultReplicas(ctx context.Context) {
	if spec.MinReplicas == nil {
		newMin := ctx.Value(DefaultingConfigKey).(DefaultingConfig).MinReplicas
		if spec.MaxReplicas != nil && *spec.MaxReplicas < newMin {
			newMin = *spec.MaxReplicas
		}

		spec.MinReplicas = &newMin
	}
	if spec.MaxReplicas == nil {
		newMax := ctx.Value(DefaultingConfigKey).(DefaultingConfig).MaxReplicas
		if *spec.MinReplicas > newMax {
			newMax = *spec.MinReplicas
		}

		spec.MaxReplicas = &newMax
	}
}

func defaultResources(res corev1.ResourceRequirements, requestMemory, requestCpu, limitsMemory, limitsCpu string) corev1.ResourceRequirements {
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
		newResource := resource.MustParse(requestCpu)
		if !copiedRes.Limits.Cpu().IsZero() && copiedRes.Limits.Cpu().Cmp(newResource) == -1 {
			newResource = *copiedRes.Limits.Cpu()
		}

		copiedRes.Requests[corev1.ResourceCPU] = newResource
	}

	if copiedRes.Limits == nil {
		copiedRes.Limits = corev1.ResourceList{}
	}
	if copiedRes.Limits.Memory().IsZero() {
		newResource := resource.MustParse(limitsMemory)
		if copiedRes.Requests.Memory().Cmp(newResource) == 1 {
			newResource = *copiedRes.Requests.Memory()
		}

		copiedRes.Limits[corev1.ResourceMemory] = newResource
	}
	if copiedRes.Limits.Cpu().IsZero() {
		newResource := resource.MustParse(limitsCpu)
		if copiedRes.Requests.Cpu().Cmp(newResource) == 1 {
			newResource = *copiedRes.Requests.Cpu()
		}

		copiedRes.Limits[corev1.ResourceCPU] = newResource
	}

	return *copiedRes
}

func (spec *FunctionSpec) defaultFunctionResources(ctx context.Context) {
	resources := spec.Resources
	defaultingConfig := ctx.Value(DefaultingConfigKey).(DefaultingConfig).Function

	spec.Resources = defaultResources(resources, defaultingConfig.RequestMemory, defaultingConfig.RequestCpu, defaultingConfig.LimitsMemory, defaultingConfig.LimitsCpu)
}

func (spec *FunctionSpec) defaultBuildResources(ctx context.Context) {
	resources := spec.BuildResources
	defaultingConfig := ctx.Value(DefaultingConfigKey).(DefaultingConfig).Build

	spec.BuildResources = defaultResources(resources, defaultingConfig.RequestMemory, defaultingConfig.RequestCpu, defaultingConfig.LimitsMemory, defaultingConfig.LimitsCpu)
}

func (spec *FunctionSpec) defaultRuntime(config DefaultingConfig) {
	if spec.Runtime == "" {
		spec.Runtime = config.Runtime
	}
}
