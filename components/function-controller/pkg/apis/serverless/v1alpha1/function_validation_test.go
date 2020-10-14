package v1alpha1

import (
	"context"
	"os"
	"testing"

	"github.com/vrischmann/envconfig"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFunctionSpec_validateResources(t *testing.T) {
	minusOne := int32(-1)
	zero := int32(0)
	one := int32(1)

	g := gomega.NewWithT(t)
	err := os.Setenv("RESERVED_ENVS", "K_CONFIGURATION")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = os.Setenv("FUNCTION_REPLICAS_MIN_VALUE", "1")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	for testName, testData := range map[string]struct {
		givenFunc              Function
		expectedError          gomega.OmegaMatcher
		specifiedExpectedError gomega.OmegaMatcher
	}{
		"Should be ok": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source:      "test-source",
					MinReplicas: &one,
					MaxReplicas: &one,
					Runtime:     Nodejs12,
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
							corev1.ResourceCPU:    resource.MustParse("300m"),
							corev1.ResourceMemory: resource.MustParse("300Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("200Mi"),
						},
					},
				},
			},
			expectedError: gomega.BeNil(),
		},
		"Should validate all fields without error": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source:  "test-source",
					Deps:    " { test }     \t\n",
					Runtime: Nodejs12,
					Env: []corev1.EnvVar{
						{
							Name:  "test",
							Value: "test",
						},
						{
							Name:  "config",
							Value: "test",
						},
					},
					Labels: map[string]string{
						"shoul-be-ok": "test",
						"test":        "test",
					},
					MinReplicas: &one,
					MaxReplicas: &one,
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
							corev1.ResourceCPU:    resource.MustParse("300m"),
							corev1.ResourceMemory: resource.MustParse("300Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("200Mi"),
						},
					},
				},
			},
			expectedError: gomega.BeNil(),
		},
		"Should return errors on empty function": {
			givenFunc:     Function{},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring(
					"metadata.name",
				),
				gomega.ContainSubstring(
					"metadata.namespace",
				),
				gomega.ContainSubstring(
					"spec.source",
				),
			),
		},
		"Should return error on deps validation": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source:  "test-source",
					Runtime: Nodejs12,
					Deps:    "{",
				},
			},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring(
					"spec.deps",
				),
			),
		},
		"Should return error on env validation": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source:  "test-source",
					Runtime: Nodejs12,
					Env: []corev1.EnvVar{
						{
							Name:  "test",
							Value: "test",
						},
						{
							Name:  "K_CONFIGURATION",
							Value: "should reject this",
						},
					},
				},
			},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring(
					"spec.env",
				),
			),
		},
		"Should return error on labels validation": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source:  "test-source",
					Runtime: Nodejs12,
					Labels: map[string]string{
						"shoul-be-ok":      "test",
						"should BE not OK": "test",
					},
				},
			},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring(
					"spec.labels",
				),
			),
		},
		"Should return error on replicas validation": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source:      "test-source",
					Runtime:     Nodejs12,
					MinReplicas: &one,
					MaxReplicas: &minusOne,
				},
			},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring(
					"spec.maxReplicas",
				),
			),
		},
		"Should return error on replicas validation on 0 minReplicas set": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source:      "test-source",
					Runtime:     Nodejs12,
					MinReplicas: &zero, // HPA needs this value to be greater then 0
					MaxReplicas: &one,
				},
			},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring(
					"spec.minReplicas",
				),
			),
		},
		"Should return error on function resources validation": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source:  "test-source",
					Runtime: Nodejs12,
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("50m"),
							corev1.ResourceMemory: resource.MustParse("64Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
			},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring(
					"spec.resources.limits.cpu",
				),
				gomega.ContainSubstring(
					"spec.resources.limits.memory",
				),
			),
		},
		"Should return error on function build resources validation": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source:  "test-source",
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
							corev1.ResourceCPU:    resource.MustParse("50m"),
							corev1.ResourceMemory: resource.MustParse("64Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
			},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring("spec.buildResources.limits.cpu"),
				gomega.ContainSubstring("spec.buildResources.limits.memory"),
				gomega.ContainSubstring("spec.buildResources.requests.memory"),
				gomega.ContainSubstring("spec.buildResources.requests.cpu"),
			),
		},
		"should return errors because of minimal config values": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source:      "test-source",
					MinReplicas: &zero,
					MaxReplicas: &zero,
					Runtime:     Nodejs12,
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("9m"),
							corev1.ResourceMemory: resource.MustParse("10Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("5m"),
							corev1.ResourceMemory: resource.MustParse("6Mi"),
						},
					},
					BuildResources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("9m"),
							corev1.ResourceMemory: resource.MustParse("10Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("5m"),
							corev1.ResourceMemory: resource.MustParse("6Mi"),
						},
					},
				},
			},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring("spec.minReplicas"),
				gomega.ContainSubstring("spec.maxReplicas"),
				gomega.ContainSubstring("spec.resources.requests.cpu"),
				gomega.ContainSubstring("spec.resources.requests.memory"),
				gomega.ContainSubstring("spec.resources.limits.cpu"),
				gomega.ContainSubstring("spec.resources.limits.memory"),
				gomega.ContainSubstring("spec.buildResources.requests.cpu"),
				gomega.ContainSubstring("spec.buildResources.requests.memory"),
				gomega.ContainSubstring("spec.buildResources.limits.cpu"),
				gomega.ContainSubstring("spec.buildResources.limits.memory"),
			),
		},
		"should be OK for git sourceType": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source: "test-source",
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
							corev1.ResourceCPU:    resource.MustParse("400m"),
							corev1.ResourceMemory: resource.MustParse("400Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("300m"),
							corev1.ResourceMemory: resource.MustParse("300Mi"),
						},
					},
					MinReplicas: &one,
					MaxReplicas: &one,
					Type:        SourceTypeGit,
					Runtime:     Nodejs12,
					Repository: Repository{
						BaseDir:   "/",
						Reference: "test-me",
					},
				},
			},
			expectedError: gomega.BeNil(),
		},
		"Should return errors OK if reference and baseDir is missing": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source: "test-source",
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
					Runtime:     Nodejs12,
					Type:        SourceTypeGit,
				},
			},
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring("spec.reference"),
				gomega.ContainSubstring("spec.baseDir"),
			),
			expectedError: gomega.HaveOccurred(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			// given
			g := gomega.NewWithT(t)
			config := &ValidationConfig{}
			err := envconfig.Init(config)
			g.Expect(err).ShouldNot(gomega.HaveOccurred())

			ctx := context.WithValue(context.Background(), ValidationConfigKey, *config)

			// when
			errs := testData.givenFunc.Validate(ctx)

			// then
			g.Expect(errs).To(testData.expectedError)
			if testData.specifiedExpectedError != nil {
				g.Expect(errs.Error()).To(testData.specifiedExpectedError)
			}
		})
	}
}
