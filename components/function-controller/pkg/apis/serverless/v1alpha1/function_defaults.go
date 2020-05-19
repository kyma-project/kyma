package v1alpha1

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"knative.dev/pkg/webhook/resourcesemantics"
)

var _ resourcesemantics.GenericCRD = (*Function)(nil)

func (fn *Function) SetDefaults(_ context.Context) {
	fn.Spec.defaultResources()
	fn.Spec.defaultReplicas()
}

func (spec *FunctionSpec) defaultReplicas() {
	if spec.MinReplicas == nil {
		one := int32(1)
		spec.MinReplicas = &one
	}
	if spec.MaxReplicas == nil {
		one := int32(1)
		spec.MaxReplicas = &one
	}
}

func (spec *FunctionSpec) defaultResources() {
	resources := spec.Resources
	if resources.Requests == nil {
		spec.Resources.Requests = corev1.ResourceList{}
	}
	if resources.Requests.Memory().IsZero() {
		spec.Resources.Requests[corev1.ResourceMemory] = resource.MustParse("64Mi")
	}
	if resources.Requests.Cpu().IsZero() {
		spec.Resources.Requests[corev1.ResourceCPU] = resource.MustParse("50m")
	}

	if resources.Limits == nil {
		spec.Resources.Limits = corev1.ResourceList{}
	}
	if resources.Limits.Memory().IsZero() {
		spec.Resources.Limits[corev1.ResourceMemory] = resource.MustParse("128Mi")
	}
	if resources.Limits.Cpu().IsZero() {
		spec.Resources.Limits[corev1.ResourceCPU] = resource.MustParse("100m")
	}
}
