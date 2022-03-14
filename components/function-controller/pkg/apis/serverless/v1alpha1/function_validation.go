package v1alpha1

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"k8s.io/apimachinery/pkg/api/validation"
	v1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
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
	for _, vFun := range vFuns {
		if err := vFun(vc); err != nil {
			return err
		}
	}
	return nil
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
		return errors.New("spec.Source is required")
	}
	return nil
}

func (spec *FunctionSpec) validateDeps(_ *ValidationConfig) error {
	if err := ValidateDependencies(spec.Runtime, spec.Deps); err != nil {
		return errors.Wrap(err, "invalid spec.Deps value")
	}
	return nil
}

func (spec *FunctionSpec) validateEnv(vc *ValidationConfig) error {
	var allErrs []string
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
				errs...,
			)
		}
	}
	if len(allErrs) > 0 {
		return fmt.Errorf("invalid spec.Env keys/values: %v", allErrs)
	}
	return nil
}

func (spec *FunctionSpec) validateFunctionResources(vc *ValidationConfig) error {
	minMemory := resource.MustParse(vc.Function.Resources.MinRequestMemory)
	minCpu := resource.MustParse(vc.Function.Resources.MinRequestCpu)

	return validateResources(spec.Resources, minMemory, minCpu)
}

func (spec *FunctionSpec) validateBuildResources(vc *ValidationConfig) error {
	minMemory := resource.MustParse(vc.BuildJob.Resources.MinRequestMemory)
	minCpu := resource.MustParse(vc.BuildJob.Resources.MinRequestCpu)

	return validateResources(spec.BuildResources, minMemory, minCpu)
}

func validateResources(resources corev1.ResourceRequirements, minMemory, minCpu resource.Quantity) error {
	limits := resources.Limits
	requests := resources.Requests

	if requests.Cpu().Cmp(minCpu) == -1 {
		return fmt.Errorf("requests cpu(%s) should be higher than minimal value (%s)",
			requests.Cpu().String(), minCpu.String())
	}
	if requests.Memory().Cmp(minMemory) == -1 {
		return fmt.Errorf("requests memory(%s) should be higher than minimal value (%s)",
			requests.Memory().String(), minMemory.String())
	}
	if limits.Cpu().Cmp(minCpu) == -1 {
		return fmt.Errorf("limits cpu(%s) should be higher than minimal value (%s)",
			limits.Cpu().String(), minCpu.String())
	}
	if limits.Memory().Cmp(minMemory) == -1 {
		return fmt.Errorf("limits memory(%s) should be higher than minimal value (%s)",
			limits.Memory().String(), minMemory.String())
	}

	if requests.Cpu().Cmp(*limits.Cpu()) == 1 {
		return fmt.Errorf("limits cpu(%s) should be higher than requests cpu(%s)",
			limits.Cpu().String(), requests.Cpu().String())
	}
	if requests.Memory().Cmp(*limits.Memory()) == 1 {
		return fmt.Errorf("limits memory(%s) should be higher than requests memory(%s)",
			limits.Memory().String(), requests.Memory().String())
	}

	return nil
}

func (spec *FunctionSpec) validateReplicas(vc *ValidationConfig) error {
	minValue := vc.Function.Replicas.MinValue
	maxReplicas := spec.MaxReplicas
	minReplicas := spec.MinReplicas

	if maxReplicas != nil && minReplicas != nil && *minReplicas > *maxReplicas {
		return fmt.Errorf("maxReplicas(%d) is less than minReplicas(%d)",
			*maxReplicas, *minReplicas)
	}
	if minReplicas != nil && *minReplicas < minValue {
		return fmt.Errorf("minReplicas(%d) is less than the smallest allowed value(%d)",
			*minReplicas, minValue)
	}
	if maxReplicas != nil && *maxReplicas < minValue {
		return fmt.Errorf("maxReplicas(%d) is less than the smallest allowed value(%d)",
			*maxReplicas, minValue)
	}

	return nil
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
	var allErrs []string
	for _, item := range properties {
		if strings.TrimSpace(item.value) != "" {
			continue
		}
		allErrs = append(allErrs, fmt.Sprintf("%s is required", item.name))
	}
	if len(allErrs) > 0 {
		return fmt.Errorf("missing required fields: %v", allErrs)
	}
	return nil
}

func (in *Repository) validateRepository(_ *ValidationConfig) error {
	return validateIfMissingFields([]property{
		{name: "spec.baseDir", value: in.BaseDir},
		{name: "spec.reference", value: in.Reference},
	}...)
}
