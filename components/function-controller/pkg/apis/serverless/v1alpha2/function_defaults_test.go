package v1alpha2

import (
	"testing"

	"github.com/stretchr/testify/require"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vrischmann/envconfig"

	"github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestSetDefaults(t *testing.T) {
	zero := int32(0)
	one := int32(1)
	two := int32(2)

	functionProfiles := `
{
	"python39": "L"
}
`
	functionReplicas := `
{
"S":{"min": 1,"max": 1},
"M":{"min": 1,"max": 2},
"L":{"min": 2}
}
`
	functionResources := `
{
"S":{"requestCpu": "25m","requestMemory": "32Mi","limitCpu": "50m","limitMemory": "64Mi"},
"M":{"requestCpu": "50m","requestMemory": "64Mi","limitCpu": "100m","limitMemory": "128Mi"},
"L":{"requestCpu": "100m","requestMemory": "128Mi","limitCpu": "200m","limitMemory": "256Mi"}
}
`

	LRuntimeResources := CreateCoreV1ResourceRequirements("200m", "256Mi", "100m", "128Mi")

	MRuntimeResources := CreateCoreV1ResourceRequirements("100m", "128Mi", "50m", "64Mi")

	buildResources := `
{
"slow":{"requestCpu": "350m","requestMemory": "350Mi","limitCpu": "700m","limitMemory": "700Mi"},
"normal":{"requestCpu": "700m","requestMemory": "700Mi","limitCpu": "1100m","limitMemory": "1100Mi"},
"fast":{"requestCpu": "1100m","requestMemory": "1100Mi", "limitCpu": "1800m","limitMemory": "1800Mi"}
}
`

	fastBuildResources := CreateCoreV1ResourceRequirements("1800m", "1800Mi", "1100m", "1100Mi")

	for testName, testData := range map[string]struct {
		givenFunc    Function
		expectedFunc Function
	}{
		"Should do nothing": {
			givenFunc: Function{
				Spec: FunctionSpec{
					Runtime: NodeJs14,
					ResourceConfiguration: ResourceConfiguration{
						Function: CreateResourceRequirements("150m", "158Mi", "90m", "84Mi"),
						Build:    CreateResourceRequirements("400m", "321Mi", "374m", "300Mi"),
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: &two,
						MaxReplicas: &two,
					},
				},
			},
			expectedFunc: Function{Spec: FunctionSpec{
				Runtime: NodeJs14,
				ResourceConfiguration: ResourceConfiguration{
					Function: CreateResourceRequirements("150m", "158Mi", "90m", "84Mi"),
					Build:    CreateResourceRequirements("400m", "321Mi", "374m", "300Mi"),
				},
				ScaleConfig: &ScaleConfig{
					MinReplicas: &two,
					MaxReplicas: &two,
				},
			},
			},
		},
		"Should not change runtime type": {
			givenFunc: Function{
				Spec: FunctionSpec{
					Runtime: Python39,
					ResourceConfiguration: ResourceConfiguration{
						Function: CreateResourceRequirements("150m", "158Mi", "90m", "84Mi"),
						Build:    CreateResourceRequirements("400m", "321Mi", "374m", "300Mi"),
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
					ResourceConfiguration: ResourceConfiguration{
						Function: CreateResourceRequirements("150m", "158Mi", "90m", "84Mi"),
						Build:    CreateResourceRequirements("400m", "321Mi", "374m", "300Mi"),
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: &two,
						MaxReplicas: &two,
					},
				},
			},
		},
		"Should not change empty runtime type to default": {
			givenFunc: Function{
				Spec: FunctionSpec{
					ResourceConfiguration: ResourceConfiguration{
						Function: CreateResourceRequirements("150m", "158Mi", "90m", "84Mi"),
						Build:    CreateResourceRequirements("400m", "321Mi", "374m", "300Mi"),
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: &two,
						MaxReplicas: &two,
					},
				},
			},
			expectedFunc: Function{
				Spec: FunctionSpec{
					ResourceConfiguration: ResourceConfiguration{
						Function: CreateResourceRequirements("150m", "158Mi", "90m", "84Mi"),
						Build:    CreateResourceRequirements("400m", "321Mi", "374m", "300Mi"),
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: &two,
						MaxReplicas: &two,
					},
				},
			},
		},
		"Should return default webhook": {
			givenFunc: Function{},
			expectedFunc: Function{
				Spec: FunctionSpec{
					ResourceConfiguration: ResourceConfiguration{
						Function: CreateResourceRequirements("100m", "128Mi", "50m", "64Mi"),
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: &one,
						MaxReplicas: &one,
					},
				},
			},
		},
		"Should fill missing fields": {
			givenFunc: Function{
				Spec: FunctionSpec{
					ResourceConfiguration: ResourceConfiguration{
						Function: CreateResourceRequirements("", "", "150m", "150Mi"),
						Build:    CreateResourceRequirements("", "", "1200m", "12000Mi"),
					},
					Replicas: &two,
				},
			},
			expectedFunc: Function{
				Spec: FunctionSpec{
					ResourceConfiguration: ResourceConfiguration{
						Function: CreateResourceRequirements("150m", "150Mi", "150m", "150Mi"),
						Build:    CreateResourceRequirements("1200m", "12000Mi", "1200m", "12000Mi"),
					},
					Replicas: &two,
				},
			},
		},
		"should consider maxReplicas and limits": {
			givenFunc: Function{
				Spec: FunctionSpec{
					ResourceConfiguration: ResourceConfiguration{
						Function: CreateResourceRequirements("15m", "15Mi", "", ""),
						Build:    CreateResourceRequirements("800m", "800Mi", "", ""),
					},
					ScaleConfig: &ScaleConfig{
						MaxReplicas: &zero,
					},
				},
			},
			expectedFunc: Function{
				Spec: FunctionSpec{
					ResourceConfiguration: ResourceConfiguration{
						Function: CreateResourceRequirements("15m", "15Mi", "15m", "15Mi"),
						Build:    CreateResourceRequirements("800m", "800Mi", "700m", "700Mi"),
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: &zero,
						MaxReplicas: &zero,
					},
				},
			},
		},
	} {
		t.Run(testName, func(t *testing.T) {
			// given
			g := gomega.NewWithT(t)
			config := &DefaultingConfig{}
			err := envconfig.Init(config)
			g.Expect(err).To(gomega.BeNil())

			functionReplicasPresets, err := ParseReplicasPresets(functionReplicas)
			g.Expect(err).To(gomega.BeNil())
			config.Function.Replicas.Presets = functionReplicasPresets

			functionResourcesPresets, err := ParseResourcePresets(functionResources)
			g.Expect(err).To(gomega.BeNil())
			config.Function.Resources.Presets = functionResourcesPresets

			buildResourcesPresets, err := ParseResourcePresets(buildResources)
			g.Expect(err).To(gomega.BeNil())
			config.BuildJob.Resources.Presets = buildResourcesPresets

			// when
			testData.givenFunc.Default(config)

			// then
			g.Expect(testData.givenFunc).To(gomega.Equal(testData.expectedFunc))
		})
	}

	testCases := map[string]struct {
		givenFunc    Function
		expectedFunc Function
	}{
		"Should properly merge resources presets - case with all fields": {
			givenFunc: Function{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						ReplicasPresetLabel:          "L",
						FunctionResourcesPresetLabel: "S",
						BuildResourcesPresetLabel:    "slow",
					},
				},
				Spec: FunctionSpec{
					Runtime: NodeJs14,
					ResourceConfiguration: ResourceConfiguration{
						Function: CreateResourceRequirements("", "", "15m", "15Mi"),
						Build:    CreateResourceRequirements("", "", "250m", "250Mi"),
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: &two,
					},
				},
			},
			expectedFunc: Function{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						ReplicasPresetLabel:          "L",
						FunctionResourcesPresetLabel: "S",
						BuildResourcesPresetLabel:    "slow",
					},
				}, Spec: FunctionSpec{
					Runtime: NodeJs14,
					ResourceConfiguration: ResourceConfiguration{
						Function: CreateResourceRequirements("50m", "64Mi", "15m", "15Mi"),
						Build:    CreateResourceRequirements("700m", "700Mi", "250m", "250Mi"),
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: &two,
						MaxReplicas: &two,
					},
				},
			},
		},
		"Should properly merge resources presets - case with concatenating missing values with default preset": {
			givenFunc: Function{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						ReplicasPresetLabel:          "L",
						FunctionResourcesPresetLabel: "L",
						BuildResourcesPresetLabel:    "fast",
					},
				},
				Spec: FunctionSpec{
					Runtime: NodeJs14,
				},
			},
			expectedFunc: Function{ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					ReplicasPresetLabel:          "L",
					FunctionResourcesPresetLabel: "L",
					BuildResourcesPresetLabel:    "fast",
				},
			}, Spec: FunctionSpec{
				Runtime: NodeJs14,
				ResourceConfiguration: ResourceConfiguration{
					Function: ResourceRequirements{
						Resources: LRuntimeResources,
					},
					Build: ResourceRequirements{
						Resources: fastBuildResources,
					},
				},
				ScaleConfig: &ScaleConfig{
					MinReplicas: &two,
					MaxReplicas: &two,
				},
			},
			},
		},
		"Should set function profile to function presets M instead of default L value": {
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
			expectedFunc: Function{ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					FunctionResourcesPresetLabel: "M",
				},
			}, Spec: FunctionSpec{
				Runtime: Python39,
				ResourceConfiguration: ResourceConfiguration{
					Function: ResourceRequirements{
						Resources: MRuntimeResources,
					},
				},
				ScaleConfig: &ScaleConfig{
					MinReplicas: &one,
					MaxReplicas: &one,
				},
			}},
		},
		"Should properly merge resources presets - case with missing buildResources Requests": {
			givenFunc: Function{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						ReplicasPresetLabel:          "L",
						FunctionResourcesPresetLabel: "S",
						BuildResourcesPresetLabel:    "slow",
					},
				},
				Spec: FunctionSpec{
					Runtime: NodeJs14,
					ResourceConfiguration: ResourceConfiguration{
						Function: CreateResourceRequirements("", "", "15m", "15Mi"),
						Build:    CreateResourceRequirements("700m", "700Mi", "", ""),
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: &two,
					},
				},
			},
			expectedFunc: Function{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						ReplicasPresetLabel:          "L",
						FunctionResourcesPresetLabel: "S",
						BuildResourcesPresetLabel:    "slow",
					},
				}, Spec: FunctionSpec{
					Runtime: NodeJs14,
					ResourceConfiguration: ResourceConfiguration{
						Function: CreateResourceRequirements("50m", "64Mi", "15m", "15Mi"),
						Build:    CreateResourceRequirements("700m", "700Mi", "350m", "350Mi"),
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: &two,
						MaxReplicas: &two,
					},
				},
			},
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			// given
			g := gomega.NewWithT(t)
			config := &DefaultingConfig{}
			err := envconfig.Init(config)
			g.Expect(err).To(gomega.BeNil())

			functionReplicasPresets, err := ParseReplicasPresets(functionReplicas)
			g.Expect(err).To(gomega.BeNil())
			config.Function.Replicas.Presets = functionReplicasPresets

			functionResourcesPresets, err := ParseResourcePresets(functionResources)
			g.Expect(err).To(gomega.BeNil())
			config.Function.Resources.Presets = functionResourcesPresets

			buildResourcesPresets, err := ParseResourcePresets(buildResources)
			g.Expect(err).To(gomega.BeNil())
			config.BuildJob.Resources.Presets = buildResourcesPresets

			functionProfile, err := ParseRuntimePresets(functionProfiles)
			g.Expect(err).To(gomega.BeNil())
			config.Function.Resources.RuntimePresets = functionProfile
			// when
			testData.givenFunc.Default(config)

			// then
			//g.Expect(testData.givenFunc).To(gomega.Equal(testData.expectedFunc))
			require.EqualValues(t, testData.givenFunc, testData.expectedFunc)
		})
	}
}

func CreateCoreV1ResourceRequirements(limitsCpu, limitsMemory, requestsCpu, requestsMemory string) corev1.ResourceRequirements {
	limits := corev1.ResourceList{}
	if limitsCpu != "" {
		limits[corev1.ResourceCPU] = resource.MustParse(limitsCpu)
	}
	if limitsMemory != "" {
		limits[corev1.ResourceMemory] = resource.MustParse(limitsMemory)
	}
	if len(limits) == 0 {
		limits = nil
	}
	requests := corev1.ResourceList{}
	if requestsCpu != "" {
		requests[corev1.ResourceCPU] = resource.MustParse(requestsCpu)
	}
	if requestsMemory != "" {
		requests[corev1.ResourceMemory] = resource.MustParse(requestsMemory)
	}
	if len(requests) == 0 {
		requests = nil
	}
	return corev1.ResourceRequirements{
		Limits:   limits,
		Requests: requests,
	}
}

func CreateResourceRequirements(limitsCpu, limitsMemory, requestsCpu, requestsMemory string) ResourceRequirements {
	return ResourceRequirements{
		Resources: CreateCoreV1ResourceRequirements(limitsCpu, limitsMemory, requestsCpu, requestsMemory),
	}
}
