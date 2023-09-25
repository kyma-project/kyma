package v1alpha2_test

import (
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"testing"

	"github.com/stretchr/testify/require"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func Test_EffectiveResource(t *testing.T) {

	MRuntimeResourcesBuilder := ResourceRequirementsBuilder{}.Limits("100m", "128Mi").Requests("50m", "64Mi")
	LRuntimeResources := ResourceRequirementsBuilder{}.Limits("200m", "256Mi").Requests("100m", "128Mi").BuildCoreV1()
	MRuntimeResources := MRuntimeResourcesBuilder.BuildCoreV1()

	testCases := map[string]struct {
		given    *serverlessv1alpha2.ResourceRequirements
		expected corev1.ResourceRequirements
	}{
		"Should choose custom": {
			given:    ResourceRequirementsBuilder{}.Limits("150m", "158Mi").Requests("90m", "84Mi").Build(),
			expected: ResourceRequirementsBuilder{}.Limits("150m", "158Mi").Requests("90m", "84Mi").BuildCoreV1(),
		},
		"Should choose default profile": {
			given:    nil,
			expected: MRuntimeResources,
		},
		"Should choose declared profile ": {
			given:    &serverlessv1alpha2.ResourceRequirements{Profile: "L"},
			expected: LRuntimeResources,
		},
		"Should choose default profile in case of not existing profile": {
			given:    &serverlessv1alpha2.ResourceRequirements{Profile: "NOT EXISTS"},
			expected: MRuntimeResources,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// given
			presets, defaultPreset := fixPresetsConfig()

			// when
			effectiveResource := tc.given.EffectiveResource(defaultPreset, presets)

			// then
			require.EqualValues(t, tc.expected, effectiveResource)
		})
	}
}

func fixPresetsConfig() (map[string]corev1.ResourceRequirements, string) {
	return map[string]corev1.ResourceRequirements{
		"S": {
			Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("50m"), corev1.ResourceMemory: resource.MustParse("64Mi")},
			Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("25m"), corev1.ResourceMemory: resource.MustParse("32Mi")},
		},
		"M": {
			Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m"), corev1.ResourceMemory: resource.MustParse("128Mi")},
			Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("50m"), corev1.ResourceMemory: resource.MustParse("64Mi")},
		},
		"L": {
			Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("200m"), corev1.ResourceMemory: resource.MustParse("256Mi")},
			Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m"), corev1.ResourceMemory: resource.MustParse("128Mi")},
		},
	}, "M"
}

type ResourceRequirementsBuilder struct {
	limitsCpu, limitsMemory, requestsCpu, requestsMemory, profile string
}

func (b ResourceRequirementsBuilder) Limits(cpu, memory string) ResourceRequirementsBuilder {
	b.limitsCpu = cpu
	b.limitsMemory = memory
	return b
}

func (b ResourceRequirementsBuilder) Requests(cpu, memory string) ResourceRequirementsBuilder {
	b.requestsCpu = cpu
	b.requestsMemory = memory
	return b
}

func (b ResourceRequirementsBuilder) Profile(profile string) ResourceRequirementsBuilder {
	b.profile = profile
	return b
}

func (b ResourceRequirementsBuilder) BuildCoreV1() corev1.ResourceRequirements {
	limits := corev1.ResourceList{}
	if b.limitsCpu != "" {
		limits[corev1.ResourceCPU] = resource.MustParse(b.limitsCpu)
	}
	if b.limitsMemory != "" {
		limits[corev1.ResourceMemory] = resource.MustParse(b.limitsMemory)
	}
	if len(limits) == 0 {
		limits = nil
	}
	requests := corev1.ResourceList{}
	if b.requestsCpu != "" {
		requests[corev1.ResourceCPU] = resource.MustParse(b.requestsCpu)
	}
	if b.requestsMemory != "" {
		requests[corev1.ResourceMemory] = resource.MustParse(b.requestsMemory)
	}
	if len(requests) == 0 {
		requests = nil
	}
	return corev1.ResourceRequirements{
		Limits:   limits,
		Requests: requests,
	}
}

func (b ResourceRequirementsBuilder) Build() *serverlessv1alpha2.ResourceRequirements {
	res := b.BuildCoreV1()
	return &serverlessv1alpha2.ResourceRequirements{
		Resources: &res,
		Profile:   b.profile,
	}
}
