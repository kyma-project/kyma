package domain

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

func (r *mutationResolver) CreateGitRepository(ctx context.Context, namespace string, name string, spec v1alpha1.GitRepositorySpec) (*v1alpha1.GitRepository, error) {
	return r.newServerless.CreateGitRepository(ctx, namespace, name, spec)
}

func (r *mutationResolver) UpdateGitRepository(ctx context.Context, namespace string, name string, spec v1alpha1.GitRepositorySpec) (*v1alpha1.GitRepository, error) {
	return r.newServerless.UpdateGitRepository(ctx, namespace, name, spec)
}

func (r *mutationResolver) DeleteGitRepository(ctx context.Context, namespace string, name string) (*v1alpha1.GitRepository, error) {
	return r.newServerless.DeleteGitRepository(ctx, namespace, name)
}

func (r *queryResolver) GitRepositories(ctx context.Context, namespace string) ([]*v1alpha1.GitRepository, error) {
	return r.newServerless.GitRepositoriesQuery(ctx, namespace)
}

func (r *queryResolver) GitRepository(ctx context.Context, namespace string, name string) (*v1alpha1.GitRepository, error) {
	return r.newServerless.GitRepositoryQuery(ctx, namespace, name)
}
