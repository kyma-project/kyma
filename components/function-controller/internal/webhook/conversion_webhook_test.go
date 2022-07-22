package webhook

import (
	"context"
	"fmt"
	"testing"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
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

	scheme := runtime.NewScheme()
	_ = serverlessv1alpha1.AddToScheme(scheme)
	_ = serverlessv1alpha2.AddToScheme(scheme)

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(testGitRepo).Build()
	w := NewConvertingWebhook(client, scheme)
	tests := []struct {
		name          string
		src           runtime.Object
		wantDst       runtime.Object
		wantVersion   string
		wantErr       bool
		checkRecreate bool
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
			name: "v1alpha1 to v1alpha2 inline function - with Resources",
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
			},
			wantDst: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: serverlessv1alpha2.ResourceConfiguration{
						Build: serverlessv1alpha2.ResourceRequirements{
							Resources: corev1.ResourceRequirements{
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
						Function: serverlessv1alpha2.ResourceRequirements{
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
					Runtime: serverlessv1alpha2.NodeJs12,
					Source: serverlessv1alpha2.Source{
						Inline: &serverlessv1alpha2.InlineSource{
							Source:       "test-source",
							Dependencies: "test-deps",
						},
					},
				},
			},
			wantVersion: serverlessv1alpha2.GroupVersion.String(),
		},
		{
			name: "v1alpha2 to v1alpha1 inline function - with ResourceConfiguration",
			src: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: serverlessv1alpha2.ResourceConfiguration{
						Build: serverlessv1alpha2.ResourceRequirements{
							Resources: corev1.ResourceRequirements{
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
						Function: serverlessv1alpha2.ResourceRequirements{
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
			name:          "v1alpha2 to v1alpha1 Git function - missing git repository",
			checkRecreate: true,

			src: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test", Namespace: "test",
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
					Source:  "test-recreated-repo",
					Repository: serverlessv1alpha1.Repository{
						BaseDir:   "/code/",
						Reference: "main",
					},
				},
			},

			wantVersion: serverlessv1alpha1.GroupVersion.String(),
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

				if tt.checkRecreate {
					validateRecreatedRepo(t,
						client,
						tt.src.(*serverlessv1alpha2.Function),
						*tt.src.(*serverlessv1alpha2.Function).Spec.Source.GitRepository.Auth)

					require.Equal(t, tt.wantDst.(*serverlessv1alpha1.Function).Spec.BaseDir,
						dst.(*serverlessv1alpha1.Function).Spec.BaseDir)

					require.Equal(t, tt.wantDst.(*serverlessv1alpha1.Function).Spec.Reference,
						dst.(*serverlessv1alpha1.Function).Spec.Reference)
				}
			} else if _, ok := tt.wantDst.(*serverlessv1alpha2.Function); ok {
				require.Equal(t, tt.wantDst.(*serverlessv1alpha2.Function).Spec, dst.(*serverlessv1alpha2.Function).Spec)

				require.NoError(t, checkGitRepoAnnotation(dst.(*serverlessv1alpha2.Function), testRepoName))

			}
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

func validateRecreatedRepo(t *testing.T, client ctrlclient.Client, f *serverlessv1alpha2.Function, authConfig serverlessv1alpha2.RepositoryAuth) {
	repoName := recreatedRepoName(f.Name)
	repo := &serverlessv1alpha1.GitRepository{}
	err := client.Get(context.Background(), types.NamespacedName{
		Name:      repoName,
		Namespace: f.Namespace,
	}, repo)

	require.NoError(t, err)
	require.NotNil(t, repo)

	require.Equal(t, authConfig.SecretName, repo.Spec.Auth.SecretName)
	require.Equal(t, authConfig.Type, serverlessv1alpha2.RepositoryAuthType(repo.Spec.Auth.Type))
}

func TestConvertingWebhook_convertFunctionWithGitConfigUpdates(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = serverlessv1alpha1.AddToScheme(scheme)
	_ = serverlessv1alpha2.AddToScheme(scheme)

	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	w := NewConvertingWebhook(client, scheme)
	//GIVEN
	startingFunction := &serverlessv1alpha2.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test", Namespace: "test",
		},
		Spec: serverlessv1alpha2.FunctionSpec{
			Runtime: serverlessv1alpha2.NodeJs12,
			Source: serverlessv1alpha2.Source{
				GitRepository: &serverlessv1alpha2.GitRepositorySource{
					URL: "https://github.com/kyma-project/kyma.git",
					Repository: serverlessv1alpha2.Repository{
						BaseDir:   "/code/",
						Reference: "main",
					},
					Auth: &serverlessv1alpha2.RepositoryAuth{
						Type:       serverlessv1alpha2.RepositoryAuthBasic,
						SecretName: "secret-name",
					},
				},
			},
		},
	}

	dst, err := w.allocateDstObject(serverlessv1alpha1.GroupVersion.String(), "Function")
	require.NoError(t, err)

	err = w.convertFunction(startingFunction, dst)
	require.NoError(t, err)

	repoName := recreatedRepoName(startingFunction.Name)
	repo := &serverlessv1alpha1.GitRepository{}
	err = client.Get(context.Background(), types.NamespacedName{
		Name:      repoName,
		Namespace: startingFunction.Namespace,
	}, repo)

	require.NoError(t, err)
	require.NotNil(t, repo)
	require.Equal(t, startingFunction.Spec.Source.GitRepository.URL, repo.Spec.URL)

	//WHEN
	startingFunction.Spec.Source.GitRepository.URL = "https://github.com/kyma-fork/kyma.git"

	err = w.convertFunction(startingFunction, dst)
	require.NoError(t, err)

	//THEN
	UpdatedRepo := &serverlessv1alpha1.GitRepository{}
	err = client.Get(context.Background(), types.NamespacedName{
		Name:      repoName,
		Namespace: startingFunction.Namespace,
	}, UpdatedRepo)

	require.NoError(t, err)
	require.NotNil(t, UpdatedRepo)
	require.Equal(t, startingFunction.Spec.Source.GitRepository.URL, UpdatedRepo.Spec.URL)
}
