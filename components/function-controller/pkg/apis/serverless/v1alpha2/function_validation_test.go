package v1alpha2

import (
	"os"
	"testing"

	"github.com/onsi/gomega"
	"github.com/vrischmann/envconfig"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

func TestFunctionSpec_validateResources(t *testing.T) {
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
			),
		},
		"Should be ok": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source: Source{
						Inline: &InlineSource{
							Source: "test-source",
						},
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: pointer.Int32(1),
						MaxReplicas: pointer.Int32(1),
					},
					Runtime: NodeJs16,
					ResourceConfiguration: &ResourceConfiguration{
						Function: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
						},
						Build: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
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
				},
			},
			expectedError: gomega.BeNil(),
		},
		"Should validate all fields without error": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source: Source{
						Inline: &InlineSource{
							Source:       "test-source",
							Dependencies: " { test }     \t\n",
						},
					},
					Runtime: NodeJs16,
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
					Template: &Template{
						Labels: map[string]string{
							"shoul-be-ok": "test",
							"test":        "test",
						},
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: pointer.Int32(1),
						MaxReplicas: pointer.Int32(1),
					},
					ResourceConfiguration: &ResourceConfiguration{
						Function: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
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
						Build: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
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
					SecretMounts: []SecretMount{
						{
							SecretName: "secret-name-1",
							MountPath:  "/mount/path/1",
						},
						{
							SecretName: "secret-name-2",
							MountPath:  "/mount/path/2",
						},
					},
				},
			},
			expectedError: gomega.BeNil(),
		},
		"Should return error on unexpected runtime name": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Runtime: "unknown-runtime-name",
					Source: Source{
						Inline: &InlineSource{
							Source: "test-source",
						},
					},
				},
			},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring(
					"spec.runtime",
				),
			),
		},
		"Should return error on empty runtime name": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source: Source{
						Inline: &InlineSource{
							Source: "test-source",
						},
					},
				},
			},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring(
					"spec.runtime",
				),
			),
		},
		"Should return error when more than one source is filled": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Runtime: NodeJs16,
					Source: Source{
						Inline: &InlineSource{
							Source:       "test-source",
							Dependencies: "{}",
						},
						GitRepository: &GitRepositorySource{URL: "fake-url", Repository: Repository{
							BaseDir:   "/",
							Reference: "ref",
						}},
					},
				},
			},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring(
					"source",
				),
			),
		},
		"Should return error when source is not filled": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Runtime: NodeJs16,
					Source:  Source{},
				},
			},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring(
					"source",
				),
			),
		},
		"Should return error on deps validation": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Runtime: NodeJs16,
					Source: Source{
						Inline: &InlineSource{
							Source:       "test-source",
							Dependencies: "{",
						},
					},
				},
			},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring(
					"source.inline.dependencies",
				),
			),
		},
		"Should return error on env validation": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Runtime: NodeJs16,
					Source: Source{
						Inline: &InlineSource{
							Source: "test-source",
						},
					},
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
					Runtime: NodeJs16,
					Source: Source{
						Inline: &InlineSource{
							Source: "test-source",
						},
					},
					Template: &Template{
						Labels: map[string]string{
							"shoul-be-ok":      "test",
							"should BE not OK": "test",
						},
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
					Runtime: NodeJs16,
					Source: Source{
						Inline: &InlineSource{
							Source: "test-source",
						},
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: pointer.Int32(1),
						MaxReplicas: pointer.Int32(-1),
					},
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
					Runtime: NodeJs16,
					Source: Source{
						Inline: &InlineSource{
							Source: "test-source",
						},
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: pointer.Int32(0), // HPA needs this value to be greater then 0
						MaxReplicas: pointer.Int32(1),
					},
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
					Runtime: NodeJs16,
					Source: Source{
						Inline: &InlineSource{
							Source: "test-source",
						},
					},
					ResourceConfiguration: &ResourceConfiguration{
						Function: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
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
				},
			},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring(
					"spec.resourceConfiguration.function.resources.limits.cpu",
				),
				gomega.ContainSubstring(
					"spec.resourceConfiguration.function.resources.limits.memory",
				),
			),
		},
		"Should return error on function build resources validation": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Runtime: NodeJs16,
					Source: Source{
						Inline: &InlineSource{
							Source: "test-source",
						},
					},
					ResourceConfiguration: &ResourceConfiguration{
						Function: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
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
						Build: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
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
				},
			},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring("spec.resourceConfiguration.build.resources.limits.cpu"),
				gomega.ContainSubstring("spec.resourceConfiguration.build.resources.limits.memory"),
				gomega.ContainSubstring("spec.resourceConfiguration.build.resources.requests.memory"),
				gomega.ContainSubstring("spec.resourceConfiguration.build.resources.requests.cpu"),
			),
		},
		"should return errors because of minimal config values": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					ScaleConfig: &ScaleConfig{
						MinReplicas: pointer.Int32(0),
						MaxReplicas: pointer.Int32(0),
					},
					Runtime: NodeJs16,
					Source: Source{
						Inline: &InlineSource{
							Source: "test-source",
						},
					},
					ResourceConfiguration: &ResourceConfiguration{
						Function: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
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
						Build: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
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
				},
			},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring("spec.minReplicas"),
				gomega.ContainSubstring("spec.maxReplicas"),
				gomega.ContainSubstring("spec.resourceConfiguration.function.resources.requests.cpu"),
				gomega.ContainSubstring("spec.resourceConfiguration.function.resources.requests.memory"),
				gomega.ContainSubstring("spec.resourceConfiguration.function.resources.limits.cpu"),
				gomega.ContainSubstring("spec.resourceConfiguration.function.resources.limits.memory"),
				gomega.ContainSubstring("spec.resourceConfiguration.build.resources.requests.cpu"),
				gomega.ContainSubstring("spec.resourceConfiguration.build.resources.requests.memory"),
				gomega.ContainSubstring("spec.resourceConfiguration.build.resources.limits.cpu"),
				gomega.ContainSubstring("spec.resourceConfiguration.build.resources.limits.memory"),
			),
		},
		"should be OK for git sourceType": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source: Source{
						GitRepository: &GitRepositorySource{
							URL: "test-source",
							Repository: Repository{
								BaseDir:   "/",
								Reference: "test-me",
							},
						},
					},
					ResourceConfiguration: &ResourceConfiguration{
						Function: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
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
						Build: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("400m"),
									corev1.ResourceMemory: resource.MustParse("400Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("300m"),
									corev1.ResourceMemory: resource.MustParse("300Mi"),
								},
							},
						},
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: pointer.Int32(1),
						MaxReplicas: pointer.Int32(1),
					},
					Runtime: NodeJs16,
				},
			},
			expectedError: gomega.BeNil(),
		},
		"Should return errors OK if reference and baseDir is missing": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source: Source{
						GitRepository: &GitRepositorySource{
							URL: "testme",
						},
					},
					ResourceConfiguration: &ResourceConfiguration{
						Function: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
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
						Build: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("200m"),
									corev1.ResourceMemory: resource.MustParse("200Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("200m"),
									corev1.ResourceMemory: resource.MustParse("200Mi"),
								},
							},
						},
					},
					ScaleConfig: &ScaleConfig{
						MinReplicas: pointer.Int32(1),
						MaxReplicas: pointer.Int32(1),
					},
					Runtime: NodeJs16,
				},
			},
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring("spec.source.gitRepository.reference"),
				gomega.ContainSubstring("spec.source.gitRepository.baseDir"),
			),
			expectedError: gomega.HaveOccurred(),
		},
		"Should not return error when replicas field is use together with scaleConfig": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source: Source{
						Inline: &InlineSource{
							Source: "test-source",
						},
					},
					Replicas: pointer.Int32(1),
					ScaleConfig: &ScaleConfig{
						MinReplicas: pointer.Int32(1),
						MaxReplicas: pointer.Int32(1),
					},
					Runtime: NodeJs14,
					ResourceConfiguration: &ResourceConfiguration{
						Function: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
						},
						Build: &ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
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
				},
			},
			expectedError: gomega.BeNil(),
		},
		"Should validate without error Resources and Profile occurring at once in ResourceConfiguration.Function/Build": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Source: Source{
						Inline: &InlineSource{
							Source:       "test-source",
							Dependencies: " { test }",
						},
					},
					Runtime: NodeJs16,
					ResourceConfiguration: &ResourceConfiguration{
						Function: &ResourceRequirements{
							Profile: "function-profile",
							Resources: &corev1.ResourceRequirements{
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
						Build: &ResourceRequirements{
							Profile: "build-profile",
							Resources: &corev1.ResourceRequirements{
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
				},
			},
			expectedError: gomega.BeNil(),
		},
		"Should return error when validate invalid secretName in secretMounts": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Runtime: NodeJs16,
					Source: Source{
						Inline: &InlineSource{
							Source: "test-source",
						},
					},
					SecretMounts: []SecretMount{
						{
							SecretName: "secret-name-1",
							MountPath:  "/mount/path/1",
						},
						{
							SecretName: "invalid secret name - not DNS subdomain name as defined in RFC 1123",
							MountPath:  "/mount/path/2",
						},
					},
				},
			},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring("spec.secretMounts"),
				gomega.ContainSubstring("RFC 1123 subdomain"),
			),
		},
		"Should return error when validate non unique secretName in secretMounts": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Runtime: NodeJs16,
					Source: Source{
						Inline: &InlineSource{
							Source: "test-source",
						},
					},
					SecretMounts: []SecretMount{
						{
							SecretName: "secret-name-1",
							MountPath:  "/mount/path/1",
						},
						{
							SecretName: "non-unique-secret-name",
							MountPath:  "/mount/path/2",
						},
						{
							SecretName: "non-unique-secret-name",
							MountPath:  "/mount/path/3",
						},
					},
				},
			},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring("spec.secretMounts"),
				gomega.ContainSubstring("secretNames should be unique"),
			),
		},
		"Should return error when validate empty mountPath in secretMounts": {
			givenFunc: Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: FunctionSpec{
					Runtime: NodeJs16,
					Source: Source{
						Inline: &InlineSource{
							Source: "test-source",
						},
					},
					SecretMounts: []SecretMount{
						{
							SecretName: "secret-name-1",
						},
					},
				},
			},
			expectedError: gomega.HaveOccurred(),
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring("spec.secretMounts"),
				gomega.ContainSubstring("mountPath should not be empty"),
			),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			tn := testName
			t.Log(tn)
			// given
			g := gomega.NewWithT(t)
			config := &ValidationConfig{}
			err := envconfig.Init(config)
			g.Expect(err).ShouldNot(gomega.HaveOccurred())

			// when
			errs := testData.givenFunc.Validate(config)
			t.Logf("err: %s", errs)
			// then
			g.Expect(errs).To(testData.expectedError)
			if testData.specifiedExpectedError != nil {
				g.Expect(errs.Error()).To(testData.specifiedExpectedError)
			}
		})
	}
}

func TestFunctionSpec_validateGitRepoURL(t *testing.T) {

	tests := []struct {
		name    string
		spec    FunctionSpec
		wantErr bool
	}{
		{
			name: "Invalid http",
			spec: FunctionSpec{
				Source: Source{
					GitRepository: &GitRepositorySource{
						URL: "github.com/kyma-project/kyma.git",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Valid http",
			spec: FunctionSpec{
				Source: Source{
					GitRepository: &GitRepositorySource{
						URL: "https://github.com/kyma-project/kyma.git",
					},
				},
			},
		},
		{
			name: "Invalid ssh",
			spec: FunctionSpec{
				Source: Source{
					GitRepository: &GitRepositorySource{
						URL: "g0t@github.com:kyma-project/kyma.git",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Valid ssh",
			spec: FunctionSpec{
				Source: Source{
					GitRepository: &GitRepositorySource{
						URL: "git@github.com:kyma-project/kyma.git",
					},
				},
			},
		},
		{
			name: "Valid ssh without .git extension",
			spec: FunctionSpec{
				Source: Source{
					GitRepository: &GitRepositorySource{
						URL: "git@github.com:kyma-project/kyma",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := tt.spec.validateGitRepoURL(&ValidationConfig{}); (err != nil) != tt.wantErr {
				t.Errorf("FunctionSpec.validateGitRepoURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
