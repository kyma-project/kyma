package v1alpha1

import (
	"context"
	"fmt"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/webhook/resourcesemantics"
)

var _ resourcesemantics.GenericCRD = (*Function)(nil)

func (fn *Function) SetDefaults(_ context.Context) {
	fn.Spec.defaultReplicas()
}

func (fn *Function) Validate(_ context.Context) *apis.FieldError {
	var allErr *apis.FieldError

	if err := fn.Spec.validateReplicas(); err != nil {
		allErr = err.ViaField("spec")
	}

	if err := fn.Spec.validateResources(); err != nil {
		allErr = allErr.Also(err.ViaField("spec"))
	}

	return allErr
}

func (spec *FunctionSpec) defaultReplicas() {
	if spec.MinReplicas == nil {
		one := int32(1)
		spec.MinReplicas = &one
	}
	if spec.MaxReplicas == nil {
		*spec.MaxReplicas = *spec.MinReplicas
	}
}

func (spec *FunctionSpec) validateReplicas() *apis.FieldError {
	maxReplicas := spec.MaxReplicas
	minReplicas := spec.MinReplicas
	if maxReplicas != nil && minReplicas != nil && *minReplicas > *maxReplicas {
		return apis.ErrInvalidValue("maxReplicas is less than minReplicas", "maxReplicas")
	}
	return nil
}

func (spec *FunctionSpec) validateResources() *apis.FieldError {
	requests := spec.Resources.Requests
	limits := spec.Resources.Limits

	var errs *apis.FieldError

	if requests.Cpu().Cmp(*limits.Cpu()) == 1 {
		errs = apis.ErrInvalidValue("limits cpu should be higher than requests cpu", "limits.cpu")
	}
	if requests.Memory().Cmp(*limits.Memory()) == 1 {
		err := apis.ErrInvalidValue("limits memory should be higher than requests memory", "limits.memory")
		errs = errs.Also(err)
	}
	if errs != nil {
		fmt.Println("----")
		fmt.Println(errs.Error())
		fmt.Println("----")
	}

	return errs
}
