package gitrepo

import (
	"context"
	"testing"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testRepoName = "test-repo"
)

func TestGitRepoReconciler_updateV1Alpha2FunctionWithRepo(t *testing.T) {
	v1alpha2Function := &serverlessv1alpha2.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test", Namespace: "test",
		},
		Spec: serverlessv1alpha2.FunctionSpec{
			Runtime: serverlessv1alpha2.NodeJs16,
			Source: serverlessv1alpha2.Source{
				GitRepository: &serverlessv1alpha2.GitRepositorySource{
					URL: "https://github.com/kyma-project/old-repo.git",
					Repository: serverlessv1alpha2.Repository{
						BaseDir:   "/code/",
						Reference: "main",
					},
				},
			},
		},
	}
	v1alpha2FunctionWithAuth := v1alpha2Function.DeepCopy()
	v1alpha2FunctionWithAuth.Spec.Source.GitRepository.Auth = &serverlessv1alpha2.RepositoryAuth{
		Type:       serverlessv1alpha2.RepositoryAuthBasic,
		SecretName: "secret_name",
	}

	log := zap.NewNop().Sugar()

	scheme := runtime.NewScheme()
	_ = serverlessv1alpha1.AddToScheme(scheme)
	_ = serverlessv1alpha2.AddToScheme(scheme)

	tests := []struct {
		name             string
		givenFunction    *serverlessv1alpha2.Function
		v1alpha1Function *serverlessv1alpha1.Function
		repo             *serverlessv1alpha1.GitRepository
		wantFunction     *serverlessv1alpha2.Function
		wantErr          bool
	}{
		{
			name:          "Update GitRepo - add auth",
			givenFunction: v1alpha2Function,
			v1alpha1Function: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs16,
					Type:    serverlessv1alpha1.SourceTypeGit,
					Source:  testRepoName,
					Repository: serverlessv1alpha1.Repository{
						BaseDir:   "/code/",
						Reference: "main",
					},
				},
			},
			repo: &serverlessv1alpha1.GitRepository{
				ObjectMeta: metav1.ObjectMeta{Name: testRepoName, Namespace: "test"},
				Spec: serverlessv1alpha1.GitRepositorySpec{
					URL: "https://github.com/kyma-project/new-repo.git",
					Auth: &serverlessv1alpha1.RepositoryAuth{
						Type:       serverlessv1alpha1.RepositoryAuthBasic,
						SecretName: "secret_name",
					},
				},
			},
			wantFunction: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test", Namespace: "test",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs16,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							URL: "https://github.com/kyma-project/new-repo.git",
							Repository: serverlessv1alpha2.Repository{
								BaseDir:   "/code/",
								Reference: "main",
							},
							Auth: &serverlessv1alpha2.RepositoryAuth{
								Type:       serverlessv1alpha2.RepositoryAuthBasic,
								SecretName: "secret_name",
							},
						},
					},
				},
			},
		},
		{
			name:          "Update GitRepo - update URL",
			givenFunction: v1alpha2Function,
			v1alpha1Function: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs16,
					Type:    serverlessv1alpha1.SourceTypeGit,
					Source:  testRepoName,
					Repository: serverlessv1alpha1.Repository{
						BaseDir:   "/code/",
						Reference: "main",
					},
				},
			},
			repo: &serverlessv1alpha1.GitRepository{
				ObjectMeta: metav1.ObjectMeta{Name: testRepoName, Namespace: "test"},
				Spec: serverlessv1alpha1.GitRepositorySpec{
					URL: "https://github.com/kyma-project/new-repo.git",
				},
			},
			wantFunction: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test", Namespace: "test",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs16,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							URL: "https://github.com/kyma-project/new-repo.git",
							Repository: serverlessv1alpha2.Repository{
								BaseDir:   "/code/",
								Reference: "main",
							},
						},
					},
				},
			},
		},
		{
			name:          "Update GitRepo - remove auth",
			givenFunction: v1alpha2FunctionWithAuth,
			v1alpha1Function: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs16,
					Type:    serverlessv1alpha1.SourceTypeGit,
					Source:  testRepoName,
					Repository: serverlessv1alpha1.Repository{
						BaseDir:   "/code/",
						Reference: "main",
					},
				},
			},
			repo: &serverlessv1alpha1.GitRepository{
				ObjectMeta: metav1.ObjectMeta{Name: testRepoName, Namespace: "test"},
				Spec: serverlessv1alpha1.GitRepositorySpec{
					URL: "https://github.com/kyma-project/new-repo.git",
				},
			},
			wantFunction: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test", Namespace: "test",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs16,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							URL: "https://github.com/kyma-project/new-repo.git",
							Repository: serverlessv1alpha2.Repository{
								BaseDir:   "/code/",
								Reference: "main",
							},
						},
					},
				},
			},
		},
		{
			name:          "Update GitRepo - update auth",
			givenFunction: v1alpha2FunctionWithAuth,
			v1alpha1Function: &serverlessv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "test"},
				Spec: serverlessv1alpha1.FunctionSpec{
					Runtime: serverlessv1alpha1.Nodejs16,
					Type:    serverlessv1alpha1.SourceTypeGit,
					Source:  testRepoName,
					Repository: serverlessv1alpha1.Repository{
						BaseDir:   "/code/",
						Reference: "main",
					},
				},
			},
			repo: &serverlessv1alpha1.GitRepository{
				ObjectMeta: metav1.ObjectMeta{Name: testRepoName, Namespace: "test"},
				Spec: serverlessv1alpha1.GitRepositorySpec{
					URL: "https://github.com/kyma-project/new-repo.git",
					Auth: &serverlessv1alpha1.RepositoryAuth{
						Type:       serverlessv1alpha1.RepositoryAuthBasic,
						SecretName: "new_secret_name",
					},
				},
			},
			wantFunction: &serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test", Namespace: "test",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					Runtime: serverlessv1alpha2.NodeJs16,
					Source: serverlessv1alpha2.Source{
						GitRepository: &serverlessv1alpha2.GitRepositorySource{
							URL: "https://github.com/kyma-project/new-repo.git",
							Repository: serverlessv1alpha2.Repository{
								BaseDir:   "/code/",
								Reference: "main",
							},
							Auth: &serverlessv1alpha2.RepositoryAuth{
								Type:       serverlessv1alpha2.RepositoryAuthBasic,
								SecretName: "new_secret_name",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.givenFunction).Build()
			r := NewGitRepoUpdateController(client, log)

			if err := r.updateV1Alpha2FunctionWithRepo(context.Background(), tt.v1alpha1Function, tt.repo); (err != nil) != tt.wantErr {
				t.Errorf("GitRepoReconciler.updateV1Alpha2FunctionWithRepo() error = %v, wantErr %v", err, tt.wantErr)
			}
			gotFunction := &serverlessv1alpha2.Function{}

			err := r.client.Get(context.Background(),
				types.NamespacedName{
					Name:      v1alpha2Function.Name,
					Namespace: v1alpha2Function.Namespace,
				},
				gotFunction)
			require.NoError(t, err)

			wantGitRepo := tt.wantFunction.Spec.Source.GitRepository
			gotGitRepo := gotFunction.Spec.Source.GitRepository
			require.Equal(t, wantGitRepo.URL, gotGitRepo.URL)

			require.Equal(t, wantGitRepo.Auth, gotGitRepo.Auth)

		})
	}
}
