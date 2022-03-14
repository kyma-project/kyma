package gitrepowebhook

import (
	"context"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
)

// type gitRepoDefaulter struct {
// 	defaultingConfig *serverlessv1alpha1.DefaultingConfig
// }

type gitRepoValidator struct {
	validationConfig *serverlessv1alpha1.ValidationConfig
}

// type GitRepoDefaulter interface {
// 	Default(ctx context.Context, obj runtime.Object) error
// }

type GitRepoValidator interface {
	ValidateCreate(ctx context.Context, obj runtime.Object) error
	ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error
	ValidateDelete(ctx context.Context, obj runtime.Object) error
}

// func NewGitRepoDefaulter(cfg *serverlessv1alpha1.DefaultingConfig) GitRepoDefaulter {
// 	return &gitRepoDefaulter{
// 		defaultingConfig: cfg,
// 	}
// }

func NewGitRepoValidator(cfg *serverlessv1alpha1.ValidationConfig) GitRepoValidator {
	return &gitRepoValidator{
		validationConfig: cfg,
	}
}

// func (grd *gitRepoDefaulter) Default(ctx context.Context, obj runtime.Object) error {
// 	return nil
// }

func (grv *gitRepoValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	return nil
}

func (grv *gitRepoValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	return nil
}

func (grv *gitRepoValidator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	return nil
}
