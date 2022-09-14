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

	LRuntimeResources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("200m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		},
	}

	MRuntimeResources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("50m"),
			corev1.ResourceMemory: resource.MustParse("64Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
	}

	buildResources := `
{
"slow":{"requestCpu": "350m","requestMemory": "350Mi","limitCpu": "700m","limitMemory": "700Mi"},
"normal":{"requestCpu": "700m","requestMemory": "700Mi","limitCpu": "1100m","limitMemory": "1100Mi"},
"fast":{"requestCpu": "1100m","requestMemory": "1100Mi", "limitCpu": "1800m","limitMemory": "1800Mi"}
}
`

	fastBuildResources := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1800m"),
			corev1.ResourceMemory: resource.MustParse("1800Mi"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1100m"),
			corev1.ResourceMemory: resource.MustParse("1100Mi"),
		},
	}

	for testName, testData := range map[string]struct {
		givenFunc    Function
		expectedFunc Function
	}{
		"Should do nothing": {
			givenFunc: Function{
				Spec: FunctionSpec{
					Runtime: NodeJs14,
					ResourceConfiguration: ResourceConfiguration{
						Function: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("150m"),
									corev1.ResourceMemory: resource.MustParse("158Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("90m"),
									corev1.ResourceMemory: resource.MustParse("84Mi"),
								},
							},
						},
						Build: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("400m"),
									corev1.ResourceMemory: resource.MustParse("321Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("374m"),
									corev1.ResourceMemory: resource.MustParse("300Mi"),
								},
							},
						},
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
					Function: ResourceRequirements{
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("150m"),
								corev1.ResourceMemory: resource.MustParse("158Mi"),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("90m"),
								corev1.ResourceMemory: resource.MustParse("84Mi"),
							},
						},
					},
					Build: ResourceRequirements{
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("400m"),
								corev1.ResourceMemory: resource.MustParse("321Mi"),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("374m"),
								corev1.ResourceMemory: resource.MustParse("300Mi"),
							},
						},
					},
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
						Function: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("150m"),
									corev1.ResourceMemory: resource.MustParse("158Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("90m"),
									corev1.ResourceMemory: resource.MustParse("84Mi"),
								},
							},
						},
						Build: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("400m"),
									corev1.ResourceMemory: resource.MustParse("321Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("374m"),
									corev1.ResourceMemory: resource.MustParse("300Mi"),
								},
							},
						},
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
						Function: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("150m"),
									corev1.ResourceMemory: resource.MustParse("158Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("90m"),
									corev1.ResourceMemory: resource.MustParse("84Mi"),
								},
							},
						},
						Build: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("400m"),
									corev1.ResourceMemory: resource.MustParse("321Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("374m"),
									corev1.ResourceMemory: resource.MustParse("300Mi"),
								},
							},
						},
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
						Function: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("150m"),
									corev1.ResourceMemory: resource.MustParse("158Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("90m"),
									corev1.ResourceMemory: resource.MustParse("84Mi"),
								},
							},
						},
						Build: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("400m"),
									corev1.ResourceMemory: resource.MustParse("321Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("374m"),
									corev1.ResourceMemory: resource.MustParse("300Mi"),
								},
							},
						},
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
						Function: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("150m"),
									corev1.ResourceMemory: resource.MustParse("158Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("90m"),
									corev1.ResourceMemory: resource.MustParse("84Mi"),
								},
							},
						},
						Build: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("400m"),
									corev1.ResourceMemory: resource.MustParse("321Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("374m"),
									corev1.ResourceMemory: resource.MustParse("300Mi"),
								},
							},
						},
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
						Function: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
						},
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
						Function: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("150m"),
									corev1.ResourceMemory: resource.MustParse("150Mi"),
								},
							},
						},
						Build: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1200m"),
									corev1.ResourceMemory: resource.MustParse("12000Mi"),
								},
							},
						},
					},
					Replicas: &two,
				},
			},
			expectedFunc: Function{
				Spec: FunctionSpec{
					ResourceConfiguration: ResourceConfiguration{
						Function: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("150m"),
									corev1.ResourceMemory: resource.MustParse("150Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("150m"),
									corev1.ResourceMemory: resource.MustParse("150Mi"),
								},
							},
						},
						Build: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1200m"),
									corev1.ResourceMemory: resource.MustParse("12000Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1200m"),
									corev1.ResourceMemory: resource.MustParse("12000Mi"),
								},
							},
						},
					},
					Replicas: &two,
				},
			},
		},
		"should consider maxReplicas and limits": {
			givenFunc: Function{
				Spec: FunctionSpec{
					ResourceConfiguration: ResourceConfiguration{
						Function: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("15m"),
									corev1.ResourceMemory: resource.MustParse("15Mi"),
								},
							},
						},
						Build: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("800m"),
									corev1.ResourceMemory: resource.MustParse("800Mi"),
								},
							},
						},
					},
					ScaleConfig: &ScaleConfig{
						MaxReplicas: &zero,
					},
				},
			},
			expectedFunc: Function{
				Spec: FunctionSpec{
					ResourceConfiguration: ResourceConfiguration{
						Function: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("15m"),
									corev1.ResourceMemory: resource.MustParse("15Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("15m"),
									corev1.ResourceMemory: resource.MustParse("15Mi"),
								},
							},
						},
						Build: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("800m"),
									corev1.ResourceMemory: resource.MustParse("800Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("700m"),
									corev1.ResourceMemory: resource.MustParse("700Mi"),
								},
							},
						},
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
						Function: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("15m"),
									corev1.ResourceMemory: resource.MustParse("15Mi"),
								},
							},
						},
						Build: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("250m"),
									corev1.ResourceMemory: resource.MustParse("250Mi"),
								},
							},
						},
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
						Function: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("15m"),
									corev1.ResourceMemory: resource.MustParse("15Mi"),
								},
							},
						},
						Build: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("700m"),
									corev1.ResourceMemory: resource.MustParse("700Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("250m"),
									corev1.ResourceMemory: resource.MustParse("250Mi"),
								},
							},
						},
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
						Function: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("15m"),
									corev1.ResourceMemory: resource.MustParse("15Mi"),
								},
							},
						},
						Build: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("700m"),
									corev1.ResourceMemory: resource.MustParse("700Mi"),
								},
							},
						},
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
						Function: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("15m"),
									corev1.ResourceMemory: resource.MustParse("15Mi"),
								},
							},
						},
						Build: ResourceRequirements{
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("700m"),
									corev1.ResourceMemory: resource.MustParse("700Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("350m"),
									corev1.ResourceMemory: resource.MustParse("350Mi"),
								},
							},
						},
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
