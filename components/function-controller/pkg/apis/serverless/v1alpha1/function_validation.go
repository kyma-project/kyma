package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/validation"
	v1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"knative.dev/pkg/apis"
)

var reservedEnvs = []string{
	"K_REVISION",
	"K_CONFIGURATION",
	"K_SERVICE",
	"FUNC_RUNTIME",
	"FUNC_HANDLER",
	"FUNC_TIMEOUT",
	"FUNC_PORT",
	"PORT",
	"MOD_NAME",
	"NODE_PATH",
}

func (fn *Function) Validate(_ context.Context) (errors *apis.FieldError) {
	errors = errors.Also(fn.validateObjectMeta("metadata"))

	spec := fn.Spec
	errors = errors.Also(spec.validateSource("spec.source"))
	errors = errors.Also(spec.validateDeps("spec.deps"))
	errors = errors.Also(spec.validateEnv("spec.env"))
	errors = errors.Also(spec.validateLabels("spec.labels"))
	errors = errors.Also(spec.validateReplicas().ViaField("spec"))
	errors = errors.Also(spec.validateResources().ViaField("spec.resources"))

	return errors
}

func (fn *Function) validateObjectMeta(fldPath string) (apisError *apis.FieldError) {
	nameFn := validation.ValidateNameFunc(validation.NameIsDNS1035Label)
	fieldPath := field.NewPath(fldPath)
	if errs := validation.ValidateObjectMeta(&fn.ObjectMeta, true, nameFn, fieldPath); errs != nil {
		for _, err := range errs {
			apisError = apisError.Also(apis.ErrInvalidValue(err.Error(), fldPath))
		}
	}
	return apisError
}

func (spec *FunctionSpec) validateSource(fldPath string) (apisError *apis.FieldError) {
	if strings.TrimSpace(spec.Source) == "" {
		apisError = apisError.Also(apis.ErrMissingField(fldPath))
	}
	return apisError
}

func (spec *FunctionSpec) validateDeps(fldPath string) (apisError *apis.FieldError) {
	if deps := strings.TrimSpace(spec.Deps); deps != "" && (deps[0] != '{' || deps[len(deps)-1] != '}') {
		apisError = apisError.Also(apis.ErrInvalidValue("deps should start with '{' and end with '}'", fldPath))
	}
	return apisError
}

func (spec *FunctionSpec) validateEnv(fldPath string) (apisError *apis.FieldError) {
	envs := spec.Env
	for _, env := range envs {
		for _, reservedEnv := range reservedEnvs {
			if env.Name == reservedEnv {
				apisError = apisError.Also(apis.ErrInvalidValue(
					fmt.Sprintf("%s env name is reserved for the serverless domain", env.Name), fldPath))
			}
		}
	}
	return apisError
}

func (spec *FunctionSpec) validateResources() (apisError *apis.FieldError) {
	requests := spec.Resources.Requests
	limits := spec.Resources.Limits

	if requests.Cpu().Cmp(*limits.Cpu()) == 1 {
		newErr := apis.ErrInvalidValue(fmt.Sprintf("limits cpu(%s) should be higher than requests cpu(%s)",
			limits.Cpu().String(), requests.Cpu().String()), "limits.cpu")
		apisError = apisError.Also(newErr)
	}
	if requests.Memory().Cmp(*limits.Memory()) == 1 {
		newErr := apis.ErrInvalidValue(fmt.Sprintf("limits memory(%s) should be higher than requests memory(%s)",
			limits.Memory().String(), requests.Memory().String()), "limits.memory")
		apisError = apisError.Also(newErr)
	}

	return apisError
}

func (spec *FunctionSpec) validateReplicas() (apisError *apis.FieldError) {
	maxReplicas := spec.MaxReplicas
	minReplicas := spec.MinReplicas
	if maxReplicas != nil && minReplicas != nil && *minReplicas > *maxReplicas {
		apisError = apisError.Also(apis.ErrInvalidValue("maxReplicas is less than minReplicas", "maxReplicas"))
	} else if minReplicas != nil && *minReplicas < 0 {
		apisError = apisError.Also(apis.ErrInvalidValue("minReplicas is less than 0", "minReplicas"))
	} else if maxReplicas != nil && *maxReplicas < 0 {
		apisError = apisError.Also(apis.ErrInvalidValue("maxReplicas is less than 0", "maxReplicas"))
	}

	return apisError
}

func (spec *FunctionSpec) validateLabels(fldPath string) (apisError *apis.FieldError) {
	labels := spec.Labels
	fieldPath := field.NewPath(fldPath)

	errors := v1validation.ValidateLabels(labels, fieldPath)
	for _, err := range errors {
		apisError.Also(apis.ErrInvalidValue(err.Error(), fldPath))
	}
	return apisError
}
