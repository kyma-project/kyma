package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"

	"k8s.io/apimachinery/pkg/api/validation"
	v1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	utilvalidation "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"knative.dev/pkg/apis"
)

const ValidationConfigKey = "validation-config"

type ValidationConfig struct {
	MinRequestCpu    string   `envconfig:"default=10m"`
	MinRequestMemory string   `envconfig:"default=16Mi"`
	MinReplicasValue int32    `envconfig:"default=1"`
	ReservedEnvs     []string `envconfig:"default={}"`
}

func (fn *Function) Validate(ctx context.Context) (errors *apis.FieldError) {
	spec := fn.Spec
	return errors.Also(
		fn.validateObjectMeta(ctx),
		spec.validateSource(ctx),
		spec.validateDeps(ctx),
		spec.validateEnv(ctx),
		spec.validateLabels(ctx),
		spec.validateReplicas(ctx),
		spec.validateResources(ctx),
	)
}

func (fn *Function) validateObjectMeta(_ context.Context) (apisError *apis.FieldError) {
	nameFn := validation.ValidateNameFunc(validation.NameIsDNS1035Label)
	fieldPath := field.NewPath("metadata")
	if errs := validation.ValidateObjectMeta(&fn.ObjectMeta, true, nameFn, fieldPath); errs != nil {
		for _, err := range errs {
			if err.Type == field.ErrorTypeRequired {
				apisError = apisError.Also(apis.ErrMissingField(err.Field))
			} else {
				apisError = apisError.Also(apis.ErrInvalidValue(err.Error(), err.Field))
			}
		}
	}
	return apisError
}

func (spec *FunctionSpec) validateSource(_ context.Context) (apisError *apis.FieldError) {
	if strings.TrimSpace(spec.Source) == "" {
		apisError = apisError.Also(apis.ErrMissingField("spec.source"))
	}
	return apisError
}

func (spec *FunctionSpec) validateDeps(_ context.Context) (apisError *apis.FieldError) {
	if deps := strings.TrimSpace(spec.Deps); deps != "" && (deps[0] != '{' || deps[len(deps)-1] != '}') {
		apisError = apisError.Also(apis.ErrInvalidValue("deps should start with '{' and end with '}'", "spec.deps"))
	}
	return apisError
}

func (spec *FunctionSpec) validateEnv(ctx context.Context) (apisError *apis.FieldError) {
	envs := spec.Env
	reservedEnvs := ctx.Value(ValidationConfigKey).(ValidationConfig).ReservedEnvs
	for _, env := range envs {
		errs := utilvalidation.IsEnvVarName(env.Name)
		for _, reservedEnv := range reservedEnvs {
			if env.Name == reservedEnv {
				errs = append(errs, "env name is reserved for the serverless domain")
			}
		}
		if len(errs) > 0 {
			apisError = apisError.Also(apis.ErrInvalidKeyName(env.Name, "spec.env", errs...))
		}
	}
	return apisError
}

func (spec *FunctionSpec) validateResources(ctx context.Context) (apisError *apis.FieldError) {
	minMemory := resource.MustParse(ctx.Value(ValidationConfigKey).(ValidationConfig).MinRequestMemory)
	minCpu := resource.MustParse(ctx.Value(ValidationConfigKey).(ValidationConfig).MinRequestCpu)
	requests := spec.Resources.Requests
	limits := spec.Resources.Limits

	if requests.Cpu().Cmp(minCpu) == -1 {
		apisError = apisError.Also(apis.ErrInvalidValue(fmt.Sprintf("requests cpu(%s) should be higher than minimal value (%s)",
			requests.Cpu().String(), minCpu.String()), "spec.resources.requests.cpu"))
	}
	if requests.Memory().Cmp(minMemory) == -1 {
		apisError = apisError.Also(apis.ErrInvalidValue(fmt.Sprintf("requests memory(%s) should be higher than minimal value (%s)",
			requests.Memory().String(), minMemory.String()), "spec.resources.requests.memory"))
	}
	if limits.Cpu().Cmp(minCpu) == -1 {
		apisError = apisError.Also(apis.ErrInvalidValue(fmt.Sprintf("limits cpu(%s) should be higher than minimal value (%s)",
			limits.Cpu().String(), minCpu.String()), "spec.resources.limits.cpu"))
	}
	if limits.Memory().Cmp(minMemory) == -1 {
		apisError = apisError.Also(apis.ErrInvalidValue(fmt.Sprintf("limits memory(%s) should be higher than minimal value (%s)",
			limits.Memory().String(), minMemory.String()), "spec.resources.limits.memory"))
	}

	if requests.Cpu().Cmp(*limits.Cpu()) == 1 {
		apisError = apisError.Also(apis.ErrInvalidValue(fmt.Sprintf("limits cpu(%s) should be higher than requests cpu(%s)",
			limits.Cpu().String(), requests.Cpu().String()), "spec.resources.limits.cpu"))
	}
	if requests.Memory().Cmp(*limits.Memory()) == 1 {
		apisError = apisError.Also(apis.ErrInvalidValue(fmt.Sprintf("limits memory(%s) should be higher than requests memory(%s)",
			limits.Memory().String(), requests.Memory().String()), "spec.resources.limits.memory"))
	}

	return apisError
}

func (spec *FunctionSpec) validateReplicas(ctx context.Context) (apisError *apis.FieldError) {
	minValue := ctx.Value(ValidationConfigKey).(ValidationConfig).MinReplicasValue
	maxReplicas := spec.MaxReplicas
	minReplicas := spec.MinReplicas
	if maxReplicas != nil && minReplicas != nil && *minReplicas > *maxReplicas {
		apisError = apisError.Also(apis.ErrInvalidValue(
			fmt.Sprintf("maxReplicas(%d) is less than minReplicas(%d)", *maxReplicas, *minReplicas), "spec.maxReplicas"))
	}
	if minReplicas != nil && *minReplicas < minValue {
		apisError = apisError.Also(apis.ErrInvalidValue(
			fmt.Sprintf("minReplicas(%d) is less than the smallest allowed value(%d)", *minReplicas, minValue), "spec.minReplicas"))
	}
	if maxReplicas != nil && *maxReplicas < minValue {
		apisError = apisError.Also(apis.ErrInvalidValue(
			fmt.Sprintf("maxReplicas(%d) is less than the smallest allowed value(%d)", *maxReplicas, minValue), "spec.maxReplicas"))
	}

	return apisError
}

func (spec *FunctionSpec) validateLabels(_ context.Context) (apisError *apis.FieldError) {
	labels := spec.Labels
	fieldPath := field.NewPath("spec.labels")

	errors := v1validation.ValidateLabels(labels, fieldPath)
	for _, err := range errors {
		apisError = apisError.Also(apis.ErrInvalidValue(err.Error(), "spec.labels"))
	}
	return apisError
}
