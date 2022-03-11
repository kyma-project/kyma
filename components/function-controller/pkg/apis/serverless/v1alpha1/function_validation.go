package v1alpha1

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"k8s.io/apimachinery/pkg/api/validation"
	v1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/errors"
	utilvalidation "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

const ValidationConfigKey = "validation-config"

type MinFunctionReplicasValues struct {
	MinValue int32 `envconfig:"default=1"`
}

type MinFunctionResourcesValues struct {
	MinRequestCpu    string `envconfig:"default=10m"`
	MinRequestMemory string `envconfig:"default=16Mi"`
}

type MinBuildJobResourcesValues struct {
	MinRequestCpu    string `envconfig:"default=200m"`
	MinRequestMemory string `envconfig:"default=200Mi"`
}

type MinFunctionValues struct {
	Replicas  MinFunctionReplicasValues
	Resources MinFunctionResourcesValues
}

type MinBuildJobValues struct {
	Resources MinBuildJobResourcesValues
}

type ValidationConfig struct {
	ReservedEnvs []string `envconfig:"default={}"`
	Function     MinFunctionValues
	BuildJob     MinBuildJobValues
}
type validationFunction func(*ValidationConfig) error

func (fn *Function) performBasicValidation(vc *ValidationConfig) error {
	validationFns := []validationFunction{
		fn.validateObjectMeta,
		fn.Spec.validateSource,
		fn.Spec.validateEnv,
		fn.Spec.validateLabels,
		fn.Spec.validateReplicas,
		fn.Spec.validateFunctionResources,
		fn.Spec.validateBuildResources,
	}
	return runValidations(validationFns, vc)
}

func (fn *Function) Validate(vc *ValidationConfig) error {
	spec := fn.Spec

	if spec.Type == SourceTypeGit {
		return runValidations(
			[]validationFunction{
				fn.performBasicValidation,
				spec.validateRepository,
			},
			vc)
	}

	return runValidations(
		[]validationFunction{
			fn.performBasicValidation,
			spec.validateDeps,
		},
		vc)
}

func runValidations(vFuns []validationFunction, vc *ValidationConfig) error {
	if vFuns == nil {
		return nil
	}
	var allErrs []error

	for _, vFun := range vFuns {
		if err := vFun(vc); err != nil {
			allErrs = append(allErrs, err)
		}
	}
	return errors.NewAggregate(allErrs)
}

func (fn *Function) validateObjectMeta(_ *ValidationConfig) error {
	nameFn := validation.ValidateNameFunc(validation.NameIsDNS1035Label)
	fieldPath := field.NewPath("metadata")
	if errs := validation.ValidateObjectMeta(&fn.ObjectMeta, true, nameFn, fieldPath); errs != nil {
		return errs.ToAggregate()
	}
	return nil
}

func (spec *FunctionSpec) validateSource(_ *ValidationConfig) error {
	if strings.TrimSpace(spec.Source) == "" {
		return field.Required(field.NewPath("spec.Source"), "spec.Source is required")
	}
	return nil
}

func (spec *FunctionSpec) validateDeps(_ *ValidationConfig) error {
	if err := ValidateDependencies(spec.Runtime, spec.Deps); err != nil {
		return field.Invalid(field.NewPath("spec.Deps"), spec.Deps, "invalid spec.Deps value")
	}
	return nil
}

func (spec *FunctionSpec) validateEnv(vc *ValidationConfig) error {
	var allErrs []error
	envs := spec.Env
	reservedEnvs := vc.ReservedEnvs
	for _, env := range envs {
		errs := utilvalidation.IsEnvVarName(env.Name)
		for _, reservedEnv := range reservedEnvs {
			if env.Name == reservedEnv {
				errs = append(errs, "env name is reserved for the serverless domain")
			}
		}
		if len(errs) > 0 {
			allErrs = append(allErrs,
				field.Invalid(
					field.NewPath("spec.Env"),
					spec.Env,
					fmt.Sprintf("invalid spec.Env values: %v", errs),
				),
			)
		}
	}
	return errors.NewAggregate(allErrs)
}

func (spec *FunctionSpec) validateFunctionResources(vc *ValidationConfig) error {
	minMemory := resource.MustParse(vc.Function.Resources.MinRequestMemory)
	minCpu := resource.MustParse(vc.Function.Resources.MinRequestCpu)

	if err := validateResources(spec.Resources, minMemory, minCpu); err != nil {
		return field.Invalid(field.NewPath("spec.resources"), spec.Resources, err.Error())
	}
	return nil
}

func (spec *FunctionSpec) validateBuildResources(vc *ValidationConfig) error {
	minMemory := resource.MustParse(vc.BuildJob.Resources.MinRequestMemory)
	minCpu := resource.MustParse(vc.BuildJob.Resources.MinRequestCpu)

	if err := validateResources(spec.BuildResources, minMemory, minCpu); err != nil {
		return field.Invalid(field.NewPath("spec.buildResources"), spec.BuildResources, err.Error())
	}
	return nil
}

func validateResources(resources corev1.ResourceRequirements, minMemory, minCpu resource.Quantity) error {
	limits := resources.Limits
	requests := resources.Requests
	var allErrs []error

	if requests.Cpu().Cmp(minCpu) == -1 {
		allErrs = append(allErrs,
			field.Invalid(
				field.NewPath("requests.cpu"),
				requests.Cpu(),
				fmt.Sprintf("requests cpu(%s) should be higher than minimal value (%s)",
					requests.Cpu().String(), minCpu.String()),
			),
		)
	}
	if requests.Memory().Cmp(minMemory) == -1 {
		allErrs = append(allErrs,
			field.Invalid(
				field.NewPath("requests.memory"),
				requests.Memory(),
				fmt.Sprintf("requests memory(%s) should be higher than minimal value (%s)",
					requests.Memory().String(), minMemory.String()),
			),
		)
	}
	if limits.Cpu().Cmp(minCpu) == -1 {
		allErrs = append(allErrs,
			field.Invalid(
				field.NewPath("limits.cpu"),
				limits.Cpu(),
				fmt.Sprintf("limits cpu(%s) should be higher than minimal value (%s)",
					limits.Cpu().String(), minCpu.String()),
			),
		)
	}
	if limits.Memory().Cmp(minMemory) == -1 {
		allErrs = append(allErrs,
			field.Invalid(
				field.NewPath("limits.memory"),
				limits.Memory(),
				fmt.Sprintf("limits memory(%s) should be higher than minimal value (%s)",
					limits.Memory().String(), minMemory.String()),
			),
		)
	}

	if requests.Cpu().Cmp(*limits.Cpu()) == 1 {
		allErrs = append(allErrs,
			field.Invalid(
				field.NewPath("limits.cpu"),
				limits.Cpu(),
				fmt.Sprintf("limits cpu(%s) should be higher than requests cpu(%s)",
					limits.Cpu().String(), requests.Cpu().String()),
			),
		)
	}
	if requests.Memory().Cmp(*limits.Memory()) == 1 {
		allErrs = append(allErrs,
			field.Invalid(
				field.NewPath("limits.memory"),
				limits.Memory(),
				fmt.Sprintf("limits memory(%s) should be higher than requests memory(%s)",
					limits.Memory().String(), requests.Memory().String()),
			),
		)
	}

	return errors.NewAggregate(allErrs)
}

func (spec *FunctionSpec) validateReplicas(vc *ValidationConfig) error {
	minValue := vc.Function.Replicas.MinValue
	maxReplicas := spec.MaxReplicas
	minReplicas := spec.MinReplicas
	var allErrs []error

	if maxReplicas != nil && minReplicas != nil && *minReplicas > *maxReplicas {
		allErrs = append(allErrs,
			field.Invalid(
				field.NewPath("spec.maxReplicas"),
				maxReplicas,
				fmt.Sprintf("maxReplicas(%d) is less than minReplicas(%d)",
					*maxReplicas, *minReplicas),
			),
		)

	}
	if minReplicas != nil && *minReplicas < minValue {
		allErrs = append(allErrs,
			field.Invalid(
				field.NewPath("spec.minReplicas"),
				minReplicas,
				fmt.Sprintf("minReplicas(%d) is less than the smallest allowed value(%d)",
					*minReplicas, minValue),
			),
		)
	}
	if maxReplicas != nil && *maxReplicas < minValue {
		allErrs = append(allErrs,
			field.Invalid(
				field.NewPath("spec.maxReplicas"),
				maxReplicas,
				fmt.Sprintf("maxReplicas(%d) is less than the smallest allowed value(%d)",
					*maxReplicas, minValue),
			),
		)

	}

	return errors.NewAggregate(allErrs)
}

func (spec *FunctionSpec) validateLabels(_ *ValidationConfig) error {
	labels := spec.Labels
	fieldPath := field.NewPath("spec.labels")

	errors := v1validation.ValidateLabels(labels, fieldPath)
	return errors.ToAggregate()
}

type property struct {
	name  string
	value string
}

func validateIfMissingFields(properties ...property) error {
	var allErrs []error
	for _, item := range properties {
		if strings.TrimSpace(item.value) != "" {
			continue
		}
		// err := apis.ErrMissingField(item.name)
		// apisError = apisError.Also(err)
		allErrs = append(allErrs, field.Required(field.NewPath(item.name), fmt.Sprintf("%s is required", item.name)))
	}
	return errors.NewAggregate(allErrs)
}

func (in *Repository) validateRepository(_ *ValidationConfig) error {
	return validateIfMissingFields([]property{
		{name: "spec.baseDir", value: in.BaseDir},
		{name: "spec.reference", value: in.Reference},
	}...)
}
