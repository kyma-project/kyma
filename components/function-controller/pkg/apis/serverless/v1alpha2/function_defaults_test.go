package v1alpha2

import (
	"testing"

	"github.com/stretchr/testify/require"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestSetDefaults(t *testing.T) {
	zero := int32(0)
	one := int32(1)
	two := int32(2)

	MRuntimeResourcesBuilder := ResourceRequirementsBuilder{}.Limits("100m", "128Mi").Requests("50m", "64Mi")
	SRuntimeResourcesBuilder := ResourceRequirementsBuilder{}.Limits("50m", "64Mi").Requests("25m", "32Mi")
	LRuntimeResources := ResourceRequirementsBuilder{}.Limits("200m", "256Mi").Requests("100m", "128Mi").BuildCoreV1()
	MRuntimeResources := MRuntimeResourcesBuilder.BuildCoreV1()

	slowBuildResourcesBuilder := ResourceRequirementsBuilder{}.Limits("700m", "700Mi").Requests("350m", "350Mi")

	for testName, testData := range map[string]struct {
		givenFunc    Function
		expectedFunc Function
	}{
		"Should do nothing": {
			givenFunc: Function{
				Spec: FunctionSpec{
					Runtime: NodeJs18,
					ResourceConfiguration: &ResourceConfiguration{
						Function: ResourceRequirementsBuilder{}.Limits("150m", "158Mi").Requests("90m", "84Mi").Build(),
						Build:    ResourceRequirementsBuilder{}.Limits("400m", "321Mi").Requests("374m", "300Mi").Build(),
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: &two,
						MaxReplicas: &two,
					},
				},
			},
			expectedFunc: Function{Spec: FunctionSpec{
				Runtime: NodeJs18,
				ResourceConfiguration: &ResourceConfiguration{
					Function: ResourceRequirementsBuilder{}.Limits("150m", "158Mi").Requests("90m", "84Mi").Build(),
					Build:    ResourceRequirementsBuilder{}.Limits("400m", "321Mi").Requests("374m", "300Mi").Build(),
				},
				ScaleConfig: &ScaleConfig{
					MinReplicas: &two,
					MaxReplicas: &two,
				},

				Replicas: &two,
			},
			},
		},
		"Should not change runtime type": {
			givenFunc: Function{
				Spec: FunctionSpec{
					Runtime: Python39,
					ResourceConfiguration: &ResourceConfiguration{
						Function: ResourceRequirementsBuilder{}.Limits("150m", "158Mi").Requests("90m", "84Mi").Build(),
						Build:    ResourceRequirementsBuilder{}.Limits("400m", "321Mi").Requests("374m", "300Mi").Build(),
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: &two,
						MaxReplicas: &two,
					},
				},
			},
			expectedFunc: Function{
				Spec: FunctionSpec{
					Runtime: Python39,
					ResourceConfiguration: &ResourceConfiguration{
						Function: ResourceRequirementsBuilder{}.Limits("150m", "158Mi").Requests("90m", "84Mi").Build(),
						Build:    ResourceRequirementsBuilder{}.Limits("400m", "321Mi").Requests("374m", "300Mi").Build(),
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: &two,
						MaxReplicas: &two,
					},
					Replicas: &two,
				},
			},
		},
		"Should not change empty runtime type to default": {
			givenFunc: Function{
				Spec: FunctionSpec{
					ResourceConfiguration: &ResourceConfiguration{
						Function: ResourceRequirementsBuilder{}.Limits("150m", "158Mi").Requests("90m", "84Mi").Build(),
						Build:    ResourceRequirementsBuilder{}.Limits("400m", "321Mi").Requests("374m", "300Mi").Build(),
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: &two,
						MaxReplicas: &two,
					},
				},
			},
			expectedFunc: Function{
				Spec: FunctionSpec{
					ResourceConfiguration: &ResourceConfiguration{
						Function: ResourceRequirementsBuilder{}.Limits("150m", "158Mi").Requests("90m", "84Mi").Build(),
						Build:    ResourceRequirementsBuilder{}.Limits("400m", "321Mi").Requests("374m", "300Mi").Build(),
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: &two,
						MaxReplicas: &two,
					},
					Replicas: &two,
				},
			},
		},
		"Should default minimal function": {
			givenFunc: Function{},
			expectedFunc: Function{
				Spec: FunctionSpec{
					ResourceConfiguration: &ResourceConfiguration{
						Function: MRuntimeResourcesBuilder.Build(),
					},
					Replicas: &one,
				},
			},
		},
		"Should not fill missing resources": {
			givenFunc: Function{
				Spec: FunctionSpec{
					ResourceConfiguration: &ResourceConfiguration{
						Function: ResourceRequirementsBuilder{}.Requests("150m", "150Mi").Build(),
						Build:    ResourceRequirementsBuilder{}.Requests("1200m", "12000Mi").Build(),
					},
					Replicas: &two,
				},
			},
			expectedFunc: Function{
				Spec: FunctionSpec{
					ResourceConfiguration: &ResourceConfiguration{
						Function: ResourceRequirementsBuilder{}.Requests("150m", "150Mi").Build(),
						Build:    ResourceRequirementsBuilder{}.Requests("1200m", "12000Mi").Build(),
					},
					Replicas: &two,
				},
			},
		},
		"should consider maxReplicas and limits": {
			givenFunc: Function{
				Spec: FunctionSpec{
					ResourceConfiguration: &ResourceConfiguration{
						Function: ResourceRequirementsBuilder{}.Limits("15m", "15Mi").Build(),
						Build:    ResourceRequirementsBuilder{}.Limits("800m", "800Mi").Build(),
					},
					ScaleConfig: &ScaleConfig{
						MaxReplicas: &zero,
					},
				},
			},
			expectedFunc: Function{
				Spec: FunctionSpec{
					ResourceConfiguration: &ResourceConfiguration{
						Function: ResourceRequirementsBuilder{}.Limits("15m", "15Mi").Build(),
						Build:    ResourceRequirementsBuilder{}.Limits("800m", "800Mi").Build(),
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: &zero,
						MaxReplicas: &zero,
					},
					Replicas: &zero,
				},
			},
		},
	} {
		t.Run(testName, func(t *testing.T) {
			// given
			config := fixDefaultingConfig()

			// when
			testData.givenFunc.Default(config)

			// then
			require.EqualValues(t, testData.expectedFunc, testData.givenFunc)
		})
	}

	testCases := map[string]struct {
		givenFunc    Function
		expectedFunc Function
	}{
		"Should properly set resources presets (using labels) - case with all fields": {
			givenFunc: Function{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						FunctionResourcesPresetLabel: "S",
						BuildResourcesPresetLabel:    "slow",
					},
				},
				Spec: FunctionSpec{
					Runtime: NodeJs18,
				},
			},
			expectedFunc: Function{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						FunctionResourcesPresetLabel: "S",
						BuildResourcesPresetLabel:    "slow",
					},
				}, Spec: FunctionSpec{
					Runtime: NodeJs18,
					ResourceConfiguration: &ResourceConfiguration{
						Function: SRuntimeResourcesBuilder.Build(),
						Build:    slowBuildResourcesBuilder.Build(),
					},
					Replicas: &one,
				},
			},
		},
		"Should properly set resources presets (using ResourceConfiguration..Preset) - case with all fields": {
			givenFunc: Function{
				ObjectMeta: v1.ObjectMeta{},
				Spec: FunctionSpec{
					Runtime: NodeJs18,
					ResourceConfiguration: &ResourceConfiguration{
						Build: &ResourceRequirements{
							Profile: "slow",
						},
						Function: &ResourceRequirements{
							Profile: "S",
						},
					},
				},
			},
			expectedFunc: Function{
				ObjectMeta: v1.ObjectMeta{}, Spec: FunctionSpec{
					Runtime: NodeJs18,
					ResourceConfiguration: &ResourceConfiguration{
						Function: SRuntimeResourcesBuilder.Profile("S").Build(),
						Build:    slowBuildResourcesBuilder.Profile("slow").Build(),
					},
					Replicas: &one,
				},
			},
		},
		"Should overwrite custom resources by presets (using labels) - case with all fields": {
			givenFunc: Function{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						FunctionResourcesPresetLabel: "S",
						BuildResourcesPresetLabel:    "slow",
					},
				},
				Spec: FunctionSpec{
					Runtime: NodeJs18,
					ResourceConfiguration: &ResourceConfiguration{
						Function: ResourceRequirementsBuilder{}.Requests("15m", "15Mi").Build(),
						Build:    ResourceRequirementsBuilder{}.Requests("250m", "250Mi").Build(),
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: &two,
					},
				},
			},
			expectedFunc: Function{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						FunctionResourcesPresetLabel: "S",
						BuildResourcesPresetLabel:    "slow",
					},
				}, Spec: FunctionSpec{
					Runtime: NodeJs18,
					ResourceConfiguration: &ResourceConfiguration{
						Function: SRuntimeResourcesBuilder.Build(),
						Build:    slowBuildResourcesBuilder.Build(),
					},
					Replicas: &two,
					ScaleConfig: &ScaleConfig{
						MinReplicas: &two,
						MaxReplicas: &two,
					},
				},
			},
		},
		"Should overwrite custom resources by presets (using ResourceConfiguration..Preset) - case with all fields": {
			givenFunc: Function{
				ObjectMeta: v1.ObjectMeta{},
				Spec: FunctionSpec{
					Runtime: NodeJs18,
					ResourceConfiguration: &ResourceConfiguration{
						Function: ResourceRequirementsBuilder{}.Requests("15m", "15Mi").Profile("S").Build(),
						Build:    ResourceRequirementsBuilder{}.Requests("250m", "250Mi").Profile("slow").Build(),
					},
					Replicas: &two,
				},
			},
			expectedFunc: Function{
				ObjectMeta: v1.ObjectMeta{},
				Spec: FunctionSpec{
					Runtime: NodeJs18,
					ResourceConfiguration: &ResourceConfiguration{
						Function: SRuntimeResourcesBuilder.Profile("S").Build(),
						Build:    slowBuildResourcesBuilder.Profile("slow").Build(),
					},
					Replicas: &two,
				},
			},
		},
		"Should set function profile to function presets M instead of default L value (using labels)": {
			givenFunc: Function{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						FunctionResourcesPresetLabel: "M",
					},
				},
				Spec: FunctionSpec{
					Runtime: Python39,
				},
			},
			expectedFunc: Function{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						FunctionResourcesPresetLabel: "M",
					},
				},
				Spec: FunctionSpec{
					Runtime: Python39,
					ResourceConfiguration: &ResourceConfiguration{
						Function: &ResourceRequirements{
							Resources: &MRuntimeResources,
						},
					},
					Replicas: &one,
				}},
		},
		"Should set function profile to function presets M instead of default L value (using ResourceConfiguration..Preset)": {
			givenFunc: Function{
				ObjectMeta: v1.ObjectMeta{},
				Spec: FunctionSpec{
					Runtime: Python39,
					ResourceConfiguration: &ResourceConfiguration{
						Function: ResourceRequirementsBuilder{}.Profile("M").Build(),
					},
				},
			},
			expectedFunc: Function{
				ObjectMeta: v1.ObjectMeta{},
				Spec: FunctionSpec{
					Runtime: Python39,
					ResourceConfiguration: &ResourceConfiguration{
						Function: &ResourceRequirements{
							Profile:   "M",
							Resources: &MRuntimeResources,
						},
					},
					Replicas: &one,
				}},
		},
		"Should set function profile to function default preset L": {
			givenFunc: Function{
				ObjectMeta: v1.ObjectMeta{},
				Spec: FunctionSpec{
					Runtime: Python39,
				},
			},
			expectedFunc: Function{
				ObjectMeta: v1.ObjectMeta{},
				Spec: FunctionSpec{
					Runtime: Python39,
					ResourceConfiguration: &ResourceConfiguration{
						Function: &ResourceRequirements{
							Resources: &LRuntimeResources,
						},
					},
					Replicas: &one,
				}},
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			// given
			config := fixDefaultingConfig()
			// when
			testData.givenFunc.Default(config)

			// then
			require.EqualValues(t, testData.expectedFunc, testData.givenFunc)
		})
	}
}

func fixDefaultingConfig() *DefaultingConfig {
	return &DefaultingConfig{
		Function: FunctionDefaulting{
			Replicas: FunctionReplicasDefaulting{
				DefaultPreset: "S",
				Presets:       map[string]ReplicasPreset{"S": {Min: 1, Max: 1}},
			},
			Resources: FunctionResourcesDefaulting{
				DefaultPreset: "M",
				Presets: map[string]ResourcesPreset{
					"S": {RequestCPU: "25m", RequestMemory: "32Mi", LimitCPU: "50m", LimitMemory: "64Mi"},
					"M": {RequestCPU: "50m", RequestMemory: "64Mi", LimitCPU: "100m", LimitMemory: "128Mi"},
					"L": {RequestCPU: "100m", RequestMemory: "128Mi", LimitCPU: "200m", LimitMemory: "256Mi"},
				},
				RuntimePresets: map[string]string{"python39": "L"},
			},
		},
		BuildJob: BuildJobDefaulting{
			Resources: BuildJobResourcesDefaulting{
				DefaultPreset: "normal",
				Presets: map[string]ResourcesPreset{
					"slow":   {RequestCPU: "350m", RequestMemory: "350Mi", LimitCPU: "700m", LimitMemory: "700Mi"},
					"normal": {RequestCPU: "700m", RequestMemory: "700Mi", LimitCPU: "1100m", LimitMemory: "1100Mi"},
					"fast":   {RequestCPU: "1100m", RequestMemory: "1100Mi", LimitCPU: "1800m", LimitMemory: "1800Mi"},
				},
			},
		},
	}
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

func (b ResourceRequirementsBuilder) Build() *ResourceRequirements {
	res := b.BuildCoreV1()
	return &ResourceRequirements{
		Resources: &res,
		Profile:   b.profile,
	}
}
