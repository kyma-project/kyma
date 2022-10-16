package webhook

import (
	"fmt"
	"testing"

	"go.uber.org/zap"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testRepoName = "test-repo"
)

func TestConvertingWebhook_convertFunction(t *testing.T) {
	testGitRepo := &serverlessv1alpha1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{Name: testRepoName, Namespace: "test"},

		Spec: serverlessv1alpha1.GitRepositorySpec{
			URL: "https://github.com/kyma-project/kyma.git",
			Auth: &serverlessv1alpha1.RepositoryAuth{
				Type:       serverlessv1alpha1.RepositoryAuthBasic,
				SecretName: "secret_name",
			},
		},
	}

	testTransitionTime := metav1.Now()
	scheme := runtime.NewScheme()
	_ = serverlessv1alpha1.AddToScheme(scheme)
	_ = serverlessv1alpha2.AddToScheme(scheme)

	one := int32(1)
	two := int32(2)
	three := int32(3)

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(testGitRepo).Build()
	fakeLogger := zap.NewNop().Sugar()
	w := NewConvertingWebhook(client, scheme, fakeLogger)
	tests := []struct {
		name        string
		src         runtime.Object
		wantDst     runtime.Object
		wantVersion string
		wantErr     bool
	}{
		{
			name: "v1alpha1 to v1alpha2 inline function",
			src: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs12,
					Source:  "test-source",
					Deps:    "test-deps",
				},
			},
			wantDst: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "test-source",
							Dependencies: "test-deps",
						},
					},
					Replicas: &one,
				},
			},
			wantVersion: serverlessv1alpha2.GroupVersion.String(),
		},
		{
			name: "v1alpha2 to v1alpha1 inline function",
			src: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "test-source",
							Dependencies: "test-deps",
						},
					},
				},
			},
			wantDst: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs12,
					Source:  "test-source",
					Deps:    "test-deps",
				},
			},
			wantVersion: serverlessv1alpha1.GroupVersion.String(),
		},
		{
			name: "v1alpha1 to v1alpha2 inline function - with Resources and Status",
			src: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs12,
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

					Source: "test-source",
					Deps:   "test-deps",
				},
				Status: serverlessv1alpha1.FunctionStatus{
					Conditions: []serverlessv1alpha1.Condition{
						{
							Type:               serverlessv1alpha1.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
					Source:  "test-source",
					Runtime: serverlessv1alpha1.RuntimeExtendedNodejs12,
				},
			},
			wantDst: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{
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
						Function: &serverlessv1alpha2.ResourceRequirements{
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
					},
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "test-source",
							Dependencies: "test-deps",
						},
					},
					Replicas: &one,
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Conditions: []serverlessv1alpha2.Condition{
						{
							Type:               serverlessv1alpha2.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
				},
			},
			wantVersion: serverlessv1alpha2.GroupVersion.String(),
		},
		{
			name: "v1alpha2 to v1alpha1 inline function - with ResourceConfiguration and Status",
			src: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{
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
						Function: &serverlessv1alpha2.ResourceRequirements{
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
					},
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "test-source",
							Dependencies: "test-deps",
						},
					},
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Conditions: []serverlessv1alpha2.Condition{
						{
							Type:               serverlessv1alpha2.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
				},
			},
			wantDst: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs12,
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

					Source: "test-source",
					Deps:   "test-deps",
				},
				Status: serverlessv1alpha1.FunctionStatus{
					Conditions: []serverlessv1alpha1.Condition{
						{
							Type:               serverlessv1alpha1.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
					Source:  "test-source",
					Runtime: serverlessv1alpha1.RuntimeExtendedNodejs12,
				},
			},
			wantVersion: serverlessv1alpha1.GroupVersion.String(),
		},
		{
			name: "v1alpha1 to v1alpha2 Git function",
			src: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs12,
					Type:    serverlessv1alpha1.SourceTypeGit,
					Source:  testRepoName,
					Repository: serverlessv1alpha1.Repository{
						BaseDir:   "/code/",
						Reference: "main",
					},
				},
			},
			wantDst: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test", Namespace: "test",
					Annotations: map[string]string{
						v1alpha1GitRepoNameAnnotation: testRepoName,
					},
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							URL: testGitRepo.Spec.URL,
							Repository: serverlessv1alpha2.Repository{
								BaseDir:   "/code/",
								Reference: "main",
							},
							Auth: &serverlessv1alpha2.RepositoryAuth{
								Type:       serverlessv1alpha2.RepositoryAuthType(testGitRepo.Spec.Auth.Type),
								SecretName: testGitRepo.Spec.Auth.SecretName,
							},
						},
					},
					Replicas: &one,
				},
			},
			wantVersion: serverlessv1alpha2.GroupVersion.String(),
		},
		{
			name: "v1alpha2 to v1alpha1 Git function",
			src: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test", Namespace: "test",
					Annotations: map[string]string{
						v1alpha1GitRepoNameAnnotation: testRepoName,
					},
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							URL: testGitRepo.Spec.URL,
							Repository: serverlessv1alpha2.Repository{
								BaseDir:   "/code/",
								Reference: "main",
							},
							Auth: &serverlessv1alpha2.RepositoryAuth{
								Type:       serverlessv1alpha2.RepositoryAuthType(testGitRepo.Spec.Auth.Type),
								SecretName: testGitRepo.Spec.Auth.SecretName,
							},
						},
					},
				},
			},
			wantDst: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs12,
					Type:    serverlessv1alpha1.SourceTypeGit,
					Source:  testRepoName,
					Repository: serverlessv1alpha1.Repository{
						BaseDir:   "/code/",
						Reference: "main",
					},
				},
			},
			wantVersion: serverlessv1alpha1.GroupVersion.String(),
		},
		{
			name: "v1alpha1 to v1alpha2 - with function-resources-preset label",
			src: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						serverlessv1alpha1.FunctionResourcesPresetLabel: "some-preset-value",
					},
				},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs12,
					Source:  "test-source",
					Deps:    "test-deps",
				},
				Status: serverlessv1alpha1.FunctionStatus{
					Conditions: []serverlessv1alpha1.Condition{
						{
							Type:               serverlessv1alpha1.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
					Source:  "test-source",
					Runtime: serverlessv1alpha1.RuntimeExtendedNodejs12,
				},
			},
			wantDst: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{
							Profile: "some-preset-value",
						},
					},
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "test-source",
							Dependencies: "test-deps",
						},
					},
					Replicas: &one,
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Conditions: []serverlessv1alpha2.Condition{
						{
							Type:               serverlessv1alpha2.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
				},
			},
			wantVersion: serverlessv1alpha2.GroupVersion.String(),
		},
		{
			name: "v1alpha2 to v1alpha1 - with function-resources-preset label",
			src: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						serverlessv1alpha2.FunctionResourcesPresetLabel: "some-preset-value",
					},
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "test-source",
							Dependencies: "test-deps",
						},
					},
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Conditions: []serverlessv1alpha2.Condition{
						{
							Type:               serverlessv1alpha2.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
				},
			},
			wantDst: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						serverlessv1alpha1.FunctionResourcesPresetLabel: "some-preset-value",
					},
				},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs12,
					Source:  "test-source",
					Deps:    "test-deps",
				},
				Status: serverlessv1alpha1.FunctionStatus{
					Conditions: []serverlessv1alpha1.Condition{
						{
							Type:               serverlessv1alpha1.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
					Source:  "test-source",
					Runtime: serverlessv1alpha1.RuntimeExtendedNodejs12,
				},
			},
			wantVersion: serverlessv1alpha1.GroupVersion.String(),
		},
		{
			name: "v1alpha2 to v1alpha1 - with Spec.ResourceConfiguration.Function.Profile",
			src: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{
							Profile: "another-preset-value",
						},
					},
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "test-source",
							Dependencies: "test-deps",
						},
					},
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Conditions: []serverlessv1alpha2.Condition{
						{
							Type:               serverlessv1alpha2.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
				},
			},
			wantDst: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						serverlessv1alpha1.FunctionResourcesPresetLabel: "another-preset-value",
					},
				},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs12,
					Source:  "test-source",
					Deps:    "test-deps",
				},
				Status: serverlessv1alpha1.FunctionStatus{
					Conditions: []serverlessv1alpha1.Condition{
						{
							Type:               serverlessv1alpha1.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
					Source:  "test-source",
					Runtime: serverlessv1alpha1.RuntimeExtendedNodejs12,
				},
			},
			wantVersion: serverlessv1alpha1.GroupVersion.String(),
		},
		{
			name: "v1alpha1 to v1alpha2 - with function-resources-preset label and Spec.Resources",
			src: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						serverlessv1alpha1.FunctionResourcesPresetLabel: "some-preset-value",
					},
				},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs12,
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("123m"),
							corev1.ResourceMemory: resource.MustParse("124Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("121m"),
							corev1.ResourceMemory: resource.MustParse("122Mi"),
						},
					},
					Source: "test-source",
					Deps:   "test-deps",
				},
				Status: serverlessv1alpha1.FunctionStatus{
					Conditions: []serverlessv1alpha1.Condition{
						{
							Type:               serverlessv1alpha1.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
					Source:  "test-source",
					Runtime: serverlessv1alpha1.RuntimeExtendedNodejs12,
				},
			},
			wantDst: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{
							Profile: "some-preset-value",
							Resources: &corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("123m"),
									corev1.ResourceMemory: resource.MustParse("124Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("121m"),
									corev1.ResourceMemory: resource.MustParse("122Mi"),
								},
							},
						},
					},
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "test-source",
							Dependencies: "test-deps",
						},
					},
					Replicas: &one,
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Conditions: []serverlessv1alpha2.Condition{
						{
							Type:               serverlessv1alpha2.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
				},
			},
			wantVersion: serverlessv1alpha2.GroupVersion.String(),
		},
		{
			name: "v1alpha2 to v1alpha1 - with function-resources-preset label and Spec.ResourceConfiguration.Function.Profile (should has highest priority)",
			src: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						serverlessv1alpha2.FunctionResourcesPresetLabel: "some-preset-value",
					},
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{
							Profile: "another-preset-value",
						},
					},
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "test-source",
							Dependencies: "test-deps",
						},
					},
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Conditions: []serverlessv1alpha2.Condition{
						{
							Type:               serverlessv1alpha2.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
				},
			},
			wantDst: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						serverlessv1alpha1.FunctionResourcesPresetLabel: "another-preset-value",
					},
				},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs12,
					Source:  "test-source",
					Deps:    "test-deps",
				},
				Status: serverlessv1alpha1.FunctionStatus{
					Conditions: []serverlessv1alpha1.Condition{
						{
							Type:               serverlessv1alpha1.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
					Source:  "test-source",
					Runtime: serverlessv1alpha1.RuntimeExtendedNodejs12,
				},
			},
			wantVersion: serverlessv1alpha1.GroupVersion.String(),
		},
		{
			name: "v1alpha2 to v1alpha1 - with function-resources-preset label and Spec.ResourceConfiguration.Function.Resources",
			src: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						serverlessv1alpha2.FunctionResourcesPresetLabel: "some-preset-value",
					},
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("731m"),
									corev1.ResourceMemory: resource.MustParse("732Mi"),
								},
							},
						},
					},
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "test-source",
							Dependencies: "test-deps",
						},
					},
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Conditions: []serverlessv1alpha2.Condition{
						{
							Type:               serverlessv1alpha2.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
				},
			},
			wantDst: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						serverlessv1alpha1.FunctionResourcesPresetLabel: "some-preset-value",
					},
				},
				Spec: serverlessv1alpha1.FunctionSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("731m"),
							corev1.ResourceMemory: resource.MustParse("732Mi"),
						},
					},
					Runtime: serverlessv1alpha1.Nodejs12,
					Source:  "test-source",
					Deps:    "test-deps",
				},
				Status: serverlessv1alpha1.FunctionStatus{
					Conditions: []serverlessv1alpha1.Condition{
						{
							Type:               serverlessv1alpha1.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
					Source:  "test-source",
					Runtime: serverlessv1alpha1.RuntimeExtendedNodejs12,
				},
			},
			wantVersion: serverlessv1alpha1.GroupVersion.String(),
		},
		{
			name: "v1alpha1 to v1alpha2 - with build-resources-preset label",
			src: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						serverlessv1alpha1.BuildResourcesPresetLabel: "some-preset-value",
					},
				},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs12,
					Source:  "test-source",
					Deps:    "test-deps",
				},
				Status: serverlessv1alpha1.FunctionStatus{
					Conditions: []serverlessv1alpha1.Condition{
						{
							Type:               serverlessv1alpha1.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
					Source:  "test-source",
					Runtime: serverlessv1alpha1.RuntimeExtendedNodejs12,
				},
			},
			wantDst: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{
							Profile: "some-preset-value",
						},
					},
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "test-source",
							Dependencies: "test-deps",
						},
					},
					Replicas: &one,
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Conditions: []serverlessv1alpha2.Condition{
						{
							Type:               serverlessv1alpha2.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
				},
			},
			wantVersion: serverlessv1alpha2.GroupVersion.String(),
		},
		{
			name: "v1alpha2 to v1alpha1 - with build-resources-preset label",
			src: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						serverlessv1alpha2.BuildResourcesPresetLabel: "some-preset-value",
					},
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "test-source",
							Dependencies: "test-deps",
						},
					},
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Conditions: []serverlessv1alpha2.Condition{
						{
							Type:               serverlessv1alpha2.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
				},
			},
			wantDst: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						serverlessv1alpha1.BuildResourcesPresetLabel: "some-preset-value",
					},
				},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs12,
					Source:  "test-source",
					Deps:    "test-deps",
				},
				Status: serverlessv1alpha1.FunctionStatus{
					Conditions: []serverlessv1alpha1.Condition{
						{
							Type:               serverlessv1alpha1.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
					Source:  "test-source",
					Runtime: serverlessv1alpha1.RuntimeExtendedNodejs12,
				},
			},
			wantVersion: serverlessv1alpha1.GroupVersion.String(),
		},
		{
			name: "v1alpha2 to v1alpha1 - with Spec.ResourceConfiguration.Build.Profile",
			src: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{
							Profile: "another-preset-value",
						},
					},
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "test-source",
							Dependencies: "test-deps",
						},
					},
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Conditions: []serverlessv1alpha2.Condition{
						{
							Type:               serverlessv1alpha2.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
				},
			},
			wantDst: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						serverlessv1alpha1.BuildResourcesPresetLabel: "another-preset-value",
					},
				},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs12,
					Source:  "test-source",
					Deps:    "test-deps",
				},
				Status: serverlessv1alpha1.FunctionStatus{
					Conditions: []serverlessv1alpha1.Condition{
						{
							Type:               serverlessv1alpha1.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
					Source:  "test-source",
					Runtime: serverlessv1alpha1.RuntimeExtendedNodejs12,
				},
			},
			wantVersion: serverlessv1alpha1.GroupVersion.String(),
		},
		{
			name: "v1alpha1 to v1alpha2 - with build-resources-preset label and Spec.Resources",
			src: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						serverlessv1alpha1.BuildResourcesPresetLabel: "some-preset-value",
					},
				},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs12,
					BuildResources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("123m"),
							corev1.ResourceMemory: resource.MustParse("124Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("121m"),
							corev1.ResourceMemory: resource.MustParse("122Mi"),
						},
					},
					Source: "test-source",
					Deps:   "test-deps",
				},
				Status: serverlessv1alpha1.FunctionStatus{
					Conditions: []serverlessv1alpha1.Condition{
						{
							Type:               serverlessv1alpha1.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
					Source:  "test-source",
					Runtime: serverlessv1alpha1.RuntimeExtendedNodejs12,
				},
			},
			wantDst: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{
							Profile: "some-preset-value",
							Resources: &corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("123m"),
									corev1.ResourceMemory: resource.MustParse("124Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("121m"),
									corev1.ResourceMemory: resource.MustParse("122Mi"),
								},
							},
						},
					},
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "test-source",
							Dependencies: "test-deps",
						},
					},
					Replicas: &one,
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Conditions: []serverlessv1alpha2.Condition{
						{
							Type:               serverlessv1alpha2.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
				},
			},
			wantVersion: serverlessv1alpha2.GroupVersion.String(),
		},
		{
			name: "v1alpha2 to v1alpha1 - with build-resources-preset label and Spec.ResourceConfiguration.Function.Profile (should has highest priority)",
			src: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						serverlessv1alpha2.BuildResourcesPresetLabel: "some-preset-value",
					},
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{
							Profile: "another-preset-value",
						},
					},
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "test-source",
							Dependencies: "test-deps",
						},
					},
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Conditions: []serverlessv1alpha2.Condition{
						{
							Type:               serverlessv1alpha2.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
				},
			},
			wantDst: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						serverlessv1alpha1.BuildResourcesPresetLabel: "another-preset-value",
					},
				},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs12,
					Source:  "test-source",
					Deps:    "test-deps",
				},
				Status: serverlessv1alpha1.FunctionStatus{
					Conditions: []serverlessv1alpha1.Condition{
						{
							Type:               serverlessv1alpha1.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
					Source:  "test-source",
					Runtime: serverlessv1alpha1.RuntimeExtendedNodejs12,
				},
			},
			wantVersion: serverlessv1alpha1.GroupVersion.String(),
		},
		{
			name: "v1alpha2 to v1alpha1 - with build-resources-preset label and Spec.ResourceConfiguration.Function.Resources",
			src: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						serverlessv1alpha2.BuildResourcesPresetLabel: "some-preset-value",
					},
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{
							Resources: &corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("731m"),
									corev1.ResourceMemory: resource.MustParse("732Mi"),
								},
							},
						},
					},
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "test-source",
							Dependencies: "test-deps",
						},
					},
				},
				Status: serverlessv1alpha2.FunctionStatus{
					Conditions: []serverlessv1alpha2.Condition{
						{
							Type:               serverlessv1alpha2.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
				},
			},
			wantDst: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						serverlessv1alpha1.BuildResourcesPresetLabel: "some-preset-value",
					},
				},
				Spec: serverlessv1alpha1.FunctionSpec{
					BuildResources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("731m"),
							corev1.ResourceMemory: resource.MustParse("732Mi"),
						},
					},
					Runtime: serverlessv1alpha1.Nodejs12,
					Source:  "test-source",
					Deps:    "test-deps",
				},
				Status: serverlessv1alpha1.FunctionStatus{
					Conditions: []serverlessv1alpha1.Condition{
						{
							Type:               serverlessv1alpha1.ConditionConfigurationReady,
							Status:             corev1.ConditionTrue,
							Message:            "Configured successfully",
							LastTransitionTime: testTransitionTime,
						},
					},
					Source:  "test-source",
					Runtime: serverlessv1alpha1.RuntimeExtendedNodejs12,
				},
			},
			wantVersion: serverlessv1alpha1.GroupVersion.String(),
		},
		{
			name: "v1alpha1 to v1alpha2 inline function -  with min/max replicas",
			src: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha1.FunctionSpec{
					MinReplicas: &two,
					MaxReplicas: &three,
					Runtime:     serverlessv1alpha1.Nodejs12,
					Source:      "test-source",
					Deps:        "test-deps",
				},
			},
			wantDst: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "test-source",
							Dependencies: "test-deps",
						},
					},
					Replicas: &two,
					ScaleConfig: &serverlessv1alpha2.ScaleConfig{
						MinReplicas: &two,
						MaxReplicas: &three,
					},
				},
			},
			wantVersion: serverlessv1alpha2.GroupVersion.String(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst, err := w.allocateDstObject(tt.wantVersion, "Function")
			require.NoError(t, err)
			if err := w.convertFunction(tt.src, dst); (err != nil) != tt.wantErr {
				t.Errorf("ConvertingWebhook.convertFunction() error = %v, wantErr %v", err, tt.wantErr)
			}
			if _, ok := tt.wantDst.(*serverlessv1alpha1.Function); ok {
				require.Equal(t, tt.wantDst.(*serverlessv1alpha1.Function).Spec, dst.(*serverlessv1alpha1.Function).Spec)
				require.Equal(t, tt.wantDst.(*serverlessv1alpha1.Function).ObjectMeta, dst.(*serverlessv1alpha1.Function).ObjectMeta)
			} else if _, ok := tt.wantDst.(*serverlessv1alpha2.Function); ok {
				require.Equal(t, tt.wantDst.(*serverlessv1alpha2.Function).Spec, dst.(*serverlessv1alpha2.Function).Spec)
				require.Equal(t, tt.wantDst.(*serverlessv1alpha2.Function).ObjectMeta, dst.(*serverlessv1alpha2.Function).ObjectMeta)

				require.NoError(t, checkGitRepoAnnotation(dst.(*serverlessv1alpha2.Function), testRepoName))

			}
		})
	}
}

func TestConvertingWebhook_convertFunctionWithRemovedAuth(t *testing.T) {
	testGitRepoNoAuth := &serverlessv1alpha1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{Name: testRepoName, Namespace: "test"},

		Spec: serverlessv1alpha1.GitRepositorySpec{
			URL: "https://github.com/kyma-project/kyma.git",
		},
	}

	scheme := runtime.NewScheme()
	_ = serverlessv1alpha1.AddToScheme(scheme)
	_ = serverlessv1alpha2.AddToScheme(scheme)

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(testGitRepoNoAuth).Build()
	fakeLogger := zap.NewNop().Sugar()
	w := NewConvertingWebhook(client, scheme, fakeLogger)
	tests := []struct {
		name        string
		src         runtime.Object
		repo        serverlessv1alpha1.GitRepository
		wantDst     runtime.Object
		wantVersion string
		wantErr     bool
	}{
		{
			name: "v1alpha1 to v1alpha2 Git function without auth",
			src: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs12,
					Type:    serverlessv1alpha1.SourceTypeGit,
					Source:  testRepoName,
					Repository: serverlessv1alpha1.Repository{
						BaseDir:   "/code/",
						Reference: "main",
					},
				},
			},
			wantDst: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test", Namespace: "test",
					Annotations: map[string]string{
						v1alpha1GitRepoNameAnnotation: testRepoName,
					},
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							URL: testGitRepoNoAuth.Spec.URL,
							Repository: serverlessv1alpha2.Repository{
								BaseDir:   "/code/",
								Reference: "main",
							},
						},
					},
				},
			},
			wantVersion: serverlessv1alpha2.GroupVersion.String(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst, err := w.allocateDstObject(tt.wantVersion, "Function")
			require.NoError(t, err)
			if err := w.convertFunction(tt.src, dst); (err != nil) != tt.wantErr {
				t.Errorf("ConvertingWebhook.convertFunction() error = %v, wantErr %v", err, tt.wantErr)
			}
			desSpec := dst.(*serverlessv1alpha2.Function).Spec
			wantSpec := tt.wantDst.(*serverlessv1alpha2.Function).Spec
			require.Equal(t, wantSpec, wantSpec)
			require.Nil(t, desSpec.Source.GitRepository.Auth)

		})
	}
}

func checkGitRepoAnnotation(f *serverlessv1alpha2.Function, repoName string) error {
	if !f.TypeOf(serverlessv1alpha2.FunctionTypeGit) {
		return nil
	}

	if n, ok := f.Annotations[v1alpha1GitRepoNameAnnotation]; !ok || n != repoName {
		return fmt.Errorf("can't find the GitRepo annotation or the value is incorrect")
	}
	return nil
}
