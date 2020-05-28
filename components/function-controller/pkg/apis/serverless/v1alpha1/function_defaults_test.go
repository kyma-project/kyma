package v1alpha1

import (
	"context"
	"testing"

	"github.com/vrischmann/envconfig"

	"github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestSetDefaults(t *testing.T) {
	zero := int32(0)
	one := int32(1)
	two := int32(2)
	for testName, testData := range map[string]struct {
		givenFunc    Function
		expectedFunc Function
	}{
		"Should do nothing": {
			givenFunc: Function{
				Spec: FunctionSpec{Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("150m"),
						corev1.ResourceMemory: resource.MustParse("158Mi"),
					},
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("90m"),
						corev1.ResourceMemory: resource.MustParse("84Mi"),
					},
				},
					MinReplicas: &two,
					MaxReplicas: &two},
			},
			expectedFunc: Function{
				Spec: FunctionSpec{Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("150m"),
						corev1.ResourceMemory: resource.MustParse("158Mi"),
					},
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("90m"),
						corev1.ResourceMemory: resource.MustParse("84Mi"),
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
					MinReplicas: &two,
				},
			},
			expectedFunc: Function{
				Spec: FunctionSpec{
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
					MaxReplicas: &zero,
				},
			},
			expectedFunc: Function{
				Spec: FunctionSpec{
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
			envconfig.Init(config)
			ctx := context.WithValue(nil, DefaultingConfigKey, *config)

			// when
			testData.givenFunc.SetDefaults(ctx)

			// then
			g.Expect(testData.givenFunc).To(gomega.Equal(testData.expectedFunc))
		})
	}
}
