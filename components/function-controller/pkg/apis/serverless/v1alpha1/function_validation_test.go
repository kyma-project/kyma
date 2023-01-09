package v1alpha1

import (
	"testing"

	"github.com/vrischmann/envconfig"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/onsi/gomega"
)

func TestFunctionSpec_validateResources(t *testing.T) {
	t.Setenv("RESERVED_ENVS", "K_CONFIGURATION")
	t.Setenv("FUNCTION_REPLICAS_MIN_VALUE", "1")

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
					MinReplicas: pointer.Int32(1),
					MaxReplicas: pointer.Int32(1),
					Runtime:     Nodejs16,
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
					Runtime: Nodejs16,
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
					MinReplicas: pointer.Int32(1),
					MaxReplicas: pointer.Int32(1),
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
					Runtime: Nodejs16,
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
					Runtime: Nodejs16,
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
					Runtime: Nodejs16,
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
					Runtime:     Nodejs16,
					MinReplicas: pointer.Int32(1),
					MaxReplicas: pointer.Int32(-1),
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
					Runtime:     Nodejs16,
					MinReplicas: pointer.Int32(0), // HPA needs this value to be greater then 0
					MaxReplicas: pointer.Int32(1),
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
					Runtime: Nodejs16,
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
					Runtime: Nodejs16,
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
					MinReplicas: pointer.Int32(0),
					MaxReplicas: pointer.Int32(0),
					Runtime:     Nodejs16,
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
					MinReplicas: pointer.Int32(1),
					MaxReplicas: pointer.Int32(1),
					Type:        SourceTypeGit,
					Runtime:     Nodejs16,
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
					MinReplicas: pointer.Int32(1),
					MaxReplicas: pointer.Int32(1),
					Runtime:     Nodejs16,
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

			// when
			errs := testData.givenFunc.Validate(config)
			// then
			g.Expect(errs).To(testData.expectedError)
			if testData.specifiedExpectedError != nil {
				g.Expect(errs.Error()).To(testData.specifiedExpectedError)
			}
		})
	}
}
