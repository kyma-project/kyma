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
	errors = fn.validateObjectMeta(errors, "metadata")

	spec := fn.Spec
	errors = spec.validateSource(errors, "spec.source")
	errors = spec.validateDeps(errors, "spec.deps")
	errors = spec.validateEnv(errors, "spec.env")
	errors = spec.validateLabels(errors, "spec.labels")
	errors = spec.validateReplicas(errors, "spec")
	errors = spec.validateResources(errors, "spec.resources")

	return errors
}

func (fn *Function) validateObjectMeta(apisError *apis.FieldError, fldPath string) *apis.FieldError {
	nameFn := validation.ValidateNameFunc(validation.NameIsDNS1035Label)
	fieldPath := field.NewPath(fldPath)
	if errs := validation.ValidateObjectMeta(&fn.ObjectMeta, true, nameFn, fieldPath); errs != nil {
		for _, err := range errs {
			apisError = apisError.Also(apis.ErrInvalidValue(err.Error(), err.Field))
		}
	}
	return apisError
}

func (spec *FunctionSpec) validateSource(apisError *apis.FieldError, fldPath string) *apis.FieldError {
	if strings.TrimSpace(spec.Source) == "" {
		apisError = apisError.Also(apis.ErrMissingField(fldPath))
	}
	return apisError
}

func (spec *FunctionSpec) validateDeps(apisError *apis.FieldError, fldPath string) *apis.FieldError {
	if deps := strings.TrimSpace(spec.Deps); deps != "" && (deps[0] != '{' || deps[len(deps)-1] != '}') {
		apisError = apisError.Also(apis.ErrInvalidValue("deps should start with '{' and end with '}'", fldPath))
	}
	return apisError
}

func (spec *FunctionSpec) validateEnv(apisError *apis.FieldError, fldPath string) *apis.FieldError {
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

func (spec *FunctionSpec) validateResources(apisError *apis.FieldError, fldPathPrefix string) *apis.FieldError {
	requests := spec.Resources.Requests
	limits := spec.Resources.Limits

	if requests.Cpu().Cmp(*limits.Cpu()) == 1 {
		newErr := apis.ErrInvalidValue(fmt.Sprintf("limits cpu(%s) should be higher than requests cpu(%s)",
			limits.Cpu().String(), requests.Cpu().String()), fmt.Sprintf("%s.limits.cpu", fldPathPrefix))
		apisError = apisError.Also(newErr)
	}
	if requests.Memory().Cmp(*limits.Memory()) == 1 {
		newErr := apis.ErrInvalidValue(fmt.Sprintf("limits memory(%s) should be higher than requests memory(%s)",
			limits.Memory().String(), requests.Memory().String()), fmt.Sprintf("%s.limits.memory", fldPathPrefix))
		apisError = apisError.Also(newErr)
	}

	return apisError
}

func (spec *FunctionSpec) validateReplicas(apisError *apis.FieldError, fldPathPrefix string) *apis.FieldError {
	maxReplicas := spec.MaxReplicas
	minReplicas := spec.MinReplicas
	if maxReplicas != nil && minReplicas != nil && *minReplicas > *maxReplicas {
		apisError = apisError.Also(apis.ErrInvalidValue(
			"maxReplicas is less than minReplicas", fmt.Sprintf("%s.maxReplicas", fldPathPrefix)))
	}
	if minReplicas != nil && *minReplicas < 0 {
		apisError = apisError.Also(apis.ErrInvalidValue("minReplicas is less than 0", fmt.Sprintf("%s.minReplicas", fldPathPrefix)))
	}
	if maxReplicas != nil && *maxReplicas < 0 {
		apisError = apisError.Also(apis.ErrInvalidValue("maxReplicas is less than 0", fmt.Sprintf("%s.maxReplicas", fldPathPrefix)))
	}

	return apisError
}

func (spec *FunctionSpec) validateLabels(apisError *apis.FieldError, fldPath string) *apis.FieldError {
	labels := spec.Labels
	fieldPath := field.NewPath(fldPath)

	errors := v1validation.ValidateLabels(labels, fieldPath)
	for _, err := range errors {
		apisError = apisError.Also(apis.ErrInvalidValue(err.Error(), fldPath))
	}
	return apisError
}
