package v1alpha1

import (
	"context"
	"testing"

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
	buildResources := `
{
"slow":{"requestCpu": "350m","requestMemory": "350Mi","limitsCpu": "700m","limitsMemory": "700Mi"},
"normal":{"requestCpu": "700m","requestMemory": "700Mi","limitsCpu": "1100m","limitsMemory": "1100Mi"},
"fast":{"limitsCpu": "1800m","limitsMemory": "1800Mi"}
}
`

	for testName, testData := range map[string]struct {
		givenFunc    Function
		expectedFunc Function
	}{
		"Should do nothing": {
			givenFunc: Function{
				Spec: FunctionSpec{
					Runtime: Nodejs12,
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
					BuildResources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("400m"),
							corev1.ResourceMemory: resource.MustParse("321Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("374m"),
							corev1.ResourceMemory: resource.MustParse("300Mi"),
						},
					},
					MinReplicas: &two,
					MaxReplicas: &two,
				},
			},
			expectedFunc: Function{Spec: FunctionSpec{
				Runtime: Nodejs12,
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
				BuildResources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("400m"),
						corev1.ResourceMemory: resource.MustParse("321Mi"),
					},
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("374m"),
						corev1.ResourceMemory: resource.MustParse("300Mi"),
					},
				},
				MinReplicas: &two,
				MaxReplicas: &two,
			},
			},
		},
		"Should not change runtime type": {
			givenFunc: Function{
				Spec: FunctionSpec{
					Runtime: Python38,
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
					BuildResources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("400m"),
							corev1.ResourceMemory: resource.MustParse("321Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("374m"),
							corev1.ResourceMemory: resource.MustParse("300Mi"),
						},
					},
					MinReplicas: &two,
					MaxReplicas: &two,
				},
			},
			expectedFunc: Function{
				Spec: FunctionSpec{
					Runtime: Python38,
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
					BuildResources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("400m"),
							corev1.ResourceMemory: resource.MustParse("321Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("374m"),
							corev1.ResourceMemory: resource.MustParse("300Mi"),
						},
					},
					MinReplicas: &two,
					MaxReplicas: &two,
				},
			},
		},
		"Should change empty runtime type to default nodejs12": {
			givenFunc: Function{
				Spec: FunctionSpec{
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
					BuildResources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("400m"),
							corev1.ResourceMemory: resource.MustParse("321Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("374m"),
							corev1.ResourceMemory: resource.MustParse("300Mi"),
						},
					},
					MinReplicas: &two,
					MaxReplicas: &two,
				},
			},
			expectedFunc: Function{
				Spec: FunctionSpec{
					Runtime: Nodejs12,
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
					BuildResources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("400m"),
							corev1.ResourceMemory: resource.MustParse("321Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("374m"),
							corev1.ResourceMemory: resource.MustParse("300Mi"),
						},
					},
					MinReplicas: &two,
					MaxReplicas: &two},
			},
		},
		"Should return default webhook": {
			givenFunc: Function{},
			expectedFunc: Function{
				Spec: FunctionSpec{
					Runtime: Nodejs12,
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
					BuildResources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1100m"),
							corev1.ResourceMemory: resource.MustParse("1100Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("700m"),
							corev1.ResourceMemory: resource.MustParse("700Mi"),
						},
					},
					MinReplicas: &one,
					MaxReplicas: &one,
				},
			},
		},
		"Should fill missing fields": {
			givenFunc: Function{
				Spec: FunctionSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("150m"),
							corev1.ResourceMemory: resource.MustParse("150Mi"),
						},
					},
					BuildResources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1200m"),
							corev1.ResourceMemory: resource.MustParse("12000Mi"),
						},
					},
					MinReplicas: &two,
				},
			},
			expectedFunc: Function{
				Spec: FunctionSpec{
					Runtime: Nodejs12,
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
					BuildResources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1200m"),
							corev1.ResourceMemory: resource.MustParse("12000Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1200m"),
							corev1.ResourceMemory: resource.MustParse("12000Mi"),
						},
					},
					MinReplicas: &two,
					MaxReplicas: &two,
				},
			},
		},
		"should consider maxReplicas and limits": {
			givenFunc: Function{
				Spec: FunctionSpec{
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("15m"),
							corev1.ResourceMemory: resource.MustParse("15Mi"),
						},
					},
					BuildResources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("800m"),
							corev1.ResourceMemory: resource.MustParse("800Mi"),
						},
					},
					MaxReplicas: &zero,
				},
			},
			expectedFunc: Function{
				Spec: FunctionSpec{
					Runtime: Nodejs12,
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
					BuildResources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("800m"),
							corev1.ResourceMemory: resource.MustParse("800Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("700m"),
							corev1.ResourceMemory: resource.MustParse("700Mi"),
						},
					},
					MinReplicas: &zero,
					MaxReplicas: &zero,
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
			presets, err := ParseResourcePresets(buildResources)
			config.BuildJob.Presets = presets
			g.Expect(err).To(gomega.BeNil())
			ctx := context.WithValue(context.Background(), DefaultingConfigKey, *config)

			// when
			testData.givenFunc.SetDefaults(ctx)

			// then
			g.Expect(testData.givenFunc).To(gomega.Equal(testData.expectedFunc))
		})
	}

	for testName, testData := range map[string]struct {
		givenFunc    Function
		expectedFunc Function
	}{
		"Should properly merge resources presets - case with all fields": {
			givenFunc: Function{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						BuildResourcesPresetLabel: "slow",
					},
				},
				Spec: FunctionSpec{
					Runtime: Nodejs12,
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
					BuildResources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("250m"),
							corev1.ResourceMemory: resource.MustParse("250Mi"),
						},
					},
					MinReplicas: &two,
					MaxReplicas: &two,
				},
			},
			expectedFunc: Function{ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					BuildResourcesPresetLabel: "slow",
				},
			}, Spec: FunctionSpec{
				Runtime: Nodejs12,
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
				BuildResources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("700m"),
						corev1.ResourceMemory: resource.MustParse("700Mi"),
					},
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("250m"),
						corev1.ResourceMemory: resource.MustParse("250Mi"),
					},
				},
				MinReplicas: &two,
				MaxReplicas: &two,
			},
			},
		},
		"Should properly merge resources presets - case with concatenating missing values with default preset": {
			givenFunc: Function{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						BuildResourcesPresetLabel: "fast",
					},
				},
				Spec: FunctionSpec{
					Runtime: Nodejs12,
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
					MinReplicas: &two,
					MaxReplicas: &two,
				},
			},
			expectedFunc: Function{ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					BuildResourcesPresetLabel: "fast",
				},
			}, Spec: FunctionSpec{
				Runtime: Nodejs12,
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
				BuildResources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1800m"),
						corev1.ResourceMemory: resource.MustParse("1800Mi"),
					},
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("700m"),
						corev1.ResourceMemory: resource.MustParse("700Mi"),
					},
				},
				MinReplicas: &two,
				MaxReplicas: &two,
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
			presets, err := ParseResourcePresets(buildResources)
			config.BuildJob.Presets = presets
			g.Expect(err).To(gomega.BeNil())
			ctx := context.WithValue(context.Background(), DefaultingConfigKey, *config)

			// when
			testData.givenFunc.SetDefaults(ctx)

			// then
			g.Expect(testData.givenFunc).To(gomega.Equal(testData.expectedFunc))
		})
	}
}
