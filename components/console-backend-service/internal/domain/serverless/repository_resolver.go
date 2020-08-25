package serverless

import (
	"context"

	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GitRepositoryList []*v1alpha1.GitRepository

func (l *GitRepositoryList) Append() interface{} {
	e := &v1alpha1.GitRepository{}
	*l = append(*l, e)
	return e
}

func (r *NewResolver) GitRepositoriesQuery(ctx context.Context, namespace string) ([]*v1alpha1.GitRepository, error) {
	items := GitRepositoryList{}
	err := r.GitRepositoryService().ListInNamespace(namespace, &items)
	return items, err
}

func (r *NewResolver) GitRepositoryQuery(ctx context.Context, namespace, name string) (*v1alpha1.GitRepository, error) {
	var result *v1alpha1.GitRepository
	err := r.GitRepositoryService().GetInNamespace(name, namespace, &result)
	return result, err
}

func (r *NewResolver) CreateGitRepository(ctx context.Context, namespace, name string, spec v1alpha1.GitRepositorySpec) (*v1alpha1.GitRepository, error) {
	gitRepository := &v1alpha1.GitRepository{
		TypeMeta: metav1.TypeMeta{
			APIVersion: gitRepositoriesGroupVersionResource.GroupVersion().String(),
			Kind:       gitRepositoryKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: spec,
	}
	result := &v1alpha1.GitRepository{}
	err := r.GitRepositoryService().Create(gitRepository, result)
	return result, err
}

func (r *NewResolver) UpdateGitRepository(ctx context.Context, namespace, name string, spec v1alpha1.GitRepositorySpec) (*v1alpha1.GitRepository, error) {
	result := &v1alpha1.GitRepository{}
	err := r.GitRepositoryService().UpdateInNamespace(name, namespace, result, func() error {
		result.Spec = spec
		return nil
	})
	return result, err
}

func (r *NewResolver) DeleteGitRepository(ctx context.Context, namespace string, name string) (*v1alpha1.GitRepository, error) {
	result := &v1alpha1.GitRepository{}
	err := r.GitRepositoryService().DeleteInNamespace(namespace, name, result)
	return result, err
}
