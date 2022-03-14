package functionwebhook

import (
	"context"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

type functionDefaulter struct {
	defaultingConfig *serverlessv1alpha1.DefaultingConfig
}

type functionValidator struct {
	validationConfig *serverlessv1alpha1.ValidationConfig
}
type FunctionDefaulter interface {
	Default(ctx context.Context, obj runtime.Object) error
}
type FunctionValidator interface {
	ValidateCreate(ctx context.Context, obj runtime.Object) error
	ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error
	ValidateDelete(ctx context.Context, obj runtime.Object) error
}

func NewFunctionDefaulter(cfg *serverlessv1alpha1.DefaultingConfig) FunctionDefaulter {
	return &functionDefaulter{
		defaultingConfig: cfg,
	}
}

func NewFunctionValidator(cfg *serverlessv1alpha1.ValidationConfig) FunctionValidator {
	return &functionValidator{
		validationConfig: cfg,
	}
}

func (fd *functionDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	f, ok := obj.(*serverlessv1alpha1.Function)
	if !ok {
		return errors.New("obj is not a serverless function object")
	}
	f.Default(fd.defaultingConfig)
	return nil
}

func (fv *functionValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	f, ok := obj.(*serverlessv1alpha1.Function)
	if !ok {
		return errors.New("obj is not a serverless function object")
	}
	return f.Validate(fv.validationConfig)

}

func (fv *functionValidator) ValidateUpdate(ctx context.Context, _, newObj runtime.Object) error {
	// we don't have any update specific validation logic
	return fv.ValidateCreate(ctx, newObj)
}

func (fv *functionValidator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	// We don't do delete validation
	return nil
}
