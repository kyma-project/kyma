package v1alpha2

import (
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"net/url"
	"regexp"
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
	MinValue int32 `yaml:"minValue"`
}

type MinFunctionResourcesValues struct {
	MinRequestCPU    string `yaml:"minRequestCPU"`
	MinRequestMemory string `yaml:"minRequestMemory"`
}

type MinBuildJobResourcesValues struct {
	MinRequestCPU    string `yaml:"minRequestCPU"`
	MinRequestMemory string `yaml:"minRequestMemory"`
}

type MinFunctionValues struct {
	Replicas  MinFunctionReplicasValues  `yaml:"replicas"`
	Resources MinFunctionResourcesValues `yaml:"resources"`
}

type MinBuildJobValues struct {
	Resources MinBuildJobResourcesValues `yaml:"resources"`
}

type ValidationConfig struct {
	ReservedEnvs []string          `yaml:"reservedEnvs"`
	Function     MinFunctionValues `yaml:"function"`
	BuildJob     MinBuildJobValues `yaml:"buildJob"`
}

type validationFunction func(*ValidationConfig) error

func (fn *Function) getBasicValidations() []validationFunction {
	return []validationFunction{
		fn.validateObjectMeta,
		fn.Spec.validateRuntime,
		fn.Spec.validateEnv,
		fn.Spec.validateLabels,
		fn.Spec.validateAnnotations,
		fn.Spec.validateReplicas,
		fn.Spec.validateFunctionResources,
		fn.Spec.validateBuildResources,
		fn.Spec.validateSources,
		fn.Spec.validateSecretMounts,
	}
}

var (
	ErrUnknownFunctionType = fmt.Errorf("unknown function source type")
)

func (fn *Function) Validate(vc *ValidationConfig) error {
	validations := fn.getBasicValidations()

	switch {
	case fn.TypeOf(FunctionTypeInline):
		validations = append(validations, fn.Spec.validateInlineSrc, fn.Spec.validateInlineDeps)
		return runValidations(vc, validations...)

	case fn.TypeOf(FunctionTypeGit):
		gitAuthValidators := fn.Spec.gitAuthValidations()
		validations = append(validations, gitAuthValidators...)
		return runValidations(vc, validations...)

	default:
		validations = append(validations, unknownFunctionTypeValidator)
		return runValidations(vc, validations...)
	}
}

func unknownFunctionTypeValidator(_ *ValidationConfig) error {
	return ErrUnknownFunctionType
}

func runValidations(vc *ValidationConfig, vFuns ...validationFunction) error {
	allErrs := []string{}
	for _, vFun := range vFuns {
		if err := vFun(vc); err != nil {
			allErrs = append(allErrs, err.Error())
		}
	}
	return returnAllErrs("", allErrs)
}

func (fn *Function) validateObjectMeta(_ *ValidationConfig) error {
	nameFn := validation.ValidateNameFunc(validation.NameIsDNS1035Label)
	fieldPath := field.NewPath("metadata")
	if errs := validation.ValidateObjectMeta(&fn.ObjectMeta, true, nameFn, fieldPath); errs != nil {
		return errs.ToAggregate()
	}
	return nil
}

func (spec *FunctionSpec) validateGitRepoURL(_ *ValidationConfig) error {
	if urlIsSSH(spec.Source.GitRepository.URL) {
		return nil
	} else if _, err := url.ParseRequestURI(spec.Source.GitRepository.URL); err != nil {
		return fmt.Errorf("invalid source.gitRepository.URL value: %v", err)
	}
	return nil
}

func (spec *FunctionSpec) validateInlineSrc(_ *ValidationConfig) error {
	if spec.Source.Inline.Source == "" {
		return fmt.Errorf("empty source.inline.source value")
	}
	return nil
}

func (spec *FunctionSpec) validateInlineDeps(_ *ValidationConfig) error {
	if err := ValidateDependencies(spec.Runtime, spec.Source.Inline.Dependencies); err != nil {
		return errors.Wrap(err, "invalid source.inline.dependencies value")
	}
	return nil
}

func (spec *FunctionSpec) gitAuthValidations() []validationFunction {
	if spec.Source.GitRepository.Auth == nil {
		return []validationFunction{
			spec.validateRepository,
		}
	}
	return []validationFunction{
		spec.validateRepository,
		spec.validateGitAuthType,
		spec.validateGitAuthSecretName,
		spec.validateGitRepoURL,
	}
}

func (spec *FunctionSpec) validateGitAuthSecretName(_ *ValidationConfig) error {
	if strings.TrimSpace(spec.Source.GitRepository.Auth.SecretName) == "" {
		return errors.New("spec.source.gitRepository.auth.secretName is required")
	}
	return nil
}

var ErrInvalidGitRepositoryAuthType = fmt.Errorf("invalid git repository authentication type")

func (spec *FunctionSpec) validateGitAuthType(_ *ValidationConfig) error {
	switch spec.Source.GitRepository.Auth.Type {
	case RepositoryAuthBasic, RepositoryAuthSSHKey:
		return nil
	default:
		return ErrInvalidGitRepositoryAuthType
	}
}

func (spec *FunctionSpec) validateRuntime(_ *ValidationConfig) error {
	runtimeName := spec.Runtime
	switch runtimeName {
	case Python39, NodeJs16, NodeJs18:
		return nil
	}
	return fmt.Errorf("spec.runtime contains unsupported value")
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
	return returnAllErrs("invalid spec.env keys/values", allErrs)
}

func (spec *FunctionSpec) validateFunctionResources(vc *ValidationConfig) error {
	minMemory := resource.MustParse(vc.Function.Resources.MinRequestMemory)
	minCPU := resource.MustParse(vc.Function.Resources.MinRequestCPU)
	if spec.ResourceConfiguration != nil && spec.ResourceConfiguration.Function != nil && spec.ResourceConfiguration.Function.Resources != nil {
		return validateResources(*spec.ResourceConfiguration.Function.Resources, minMemory, minCPU, "spec.resourceConfiguration.function.resources")
	}
	return nil
}

func (spec *FunctionSpec) validateBuildResources(vc *ValidationConfig) error {
	minMemory := resource.MustParse(vc.BuildJob.Resources.MinRequestMemory)
	minCPU := resource.MustParse(vc.BuildJob.Resources.MinRequestCPU)
	if spec.ResourceConfiguration != nil && spec.ResourceConfiguration.Build != nil && spec.ResourceConfiguration.Build.Resources != nil {
		return validateResources(*spec.ResourceConfiguration.Build.Resources, minMemory, minCPU, "spec.resourceConfiguration.build.resources")
	}
	return nil
}

func (spec *FunctionSpec) validateSources(vc *ValidationConfig) error {
	sources := 0
	if spec.Source.GitRepository != nil {
		sources++
	}

	if spec.Source.Inline != nil {
		sources++
	}
	if sources == 1 {
		return nil
	}
	return errors.Errorf("spec.source should contains only 1 configuration of function")
}

func validateResources(resources corev1.ResourceRequirements, minMemory, minCPU resource.Quantity, parent string) error {
	limits := resources.Limits
	requests := resources.Requests
	allErrs := []string{}

	if requests != nil {
		allErrs = append(allErrs, validateRequests(resources, minMemory, minCPU, parent)...)
	}

	if limits != nil {
		allErrs = append(allErrs, validateLimits(resources, minMemory, minCPU, parent)...)
	}
	return returnAllErrs("invalid function resources", allErrs)
}

func validateRequests(resources corev1.ResourceRequirements, minMemory, minCPU resource.Quantity, parent string) []string {
	limits := resources.Limits
	requests := resources.Requests
	allErrs := []string{}

	if requests.Cpu().Cmp(minCPU) == -1 {
		allErrs = append(allErrs, fmt.Sprintf("%s.requests.cpu(%s) should be higher than minimal value (%s)",
			parent, requests.Cpu().String(), minCPU.String()))
	}
	if requests.Memory().Cmp(minMemory) == -1 {
		allErrs = append(allErrs, fmt.Sprintf("%s.requests.memory(%s) should be higher than minimal value (%s)",
			parent, requests.Memory().String(), minMemory.String()))
	}

	if limits == nil {
		return allErrs
	}

	if requests.Cpu().Cmp(*limits.Cpu()) == 1 {
		allErrs = append(allErrs, fmt.Sprintf("%s.limits.cpu(%s) should be higher than %s.requests.cpu(%s)",
			parent, limits.Cpu().String(), parent, requests.Cpu().String()))
	}
	if requests.Memory().Cmp(*limits.Memory()) == 1 {
		allErrs = append(allErrs, fmt.Sprintf("%s.limits.memory(%s) should be higher than %s.requests.memory(%s)",
			parent, limits.Memory().String(), parent, requests.Memory().String()))
	}

	return allErrs
}

func validateLimits(resources corev1.ResourceRequirements, minMemory, minCPU resource.Quantity, parent string) []string {
	limits := resources.Limits
	allErrs := []string{}

	if limits != nil {
		if limits.Cpu().Cmp(minCPU) == -1 {
			allErrs = append(allErrs, fmt.Sprintf("%s.limits.cpu(%s) should be higher than minimal value (%s)",
				parent, limits.Cpu().String(), minCPU.String()))
		}
		if limits.Memory().Cmp(minMemory) == -1 {
			allErrs = append(allErrs, fmt.Sprintf("%s.limits.memory(%s) should be higher than minimal value (%s)",
				parent, limits.Memory().String(), minMemory.String()))
		}
	}
	return allErrs
}

func (spec *FunctionSpec) validateReplicas(vc *ValidationConfig) error {
	minValue := vc.Function.Replicas.MinValue
	var maxReplicas *int32
	var minReplicas *int32
	if spec.ScaleConfig == nil {
		return nil
	}
	maxReplicas = spec.ScaleConfig.MaxReplicas
	minReplicas = spec.ScaleConfig.MinReplicas
	allErrs := []string{}

	if spec.Replicas == nil && spec.ScaleConfig == nil {
		allErrs = append(allErrs, "spec.replicas and spec.scaleConfig are empty at the same time")
	}
	if maxReplicas != nil && minReplicas != nil && *minReplicas > *maxReplicas {
		allErrs = append(allErrs, fmt.Sprintf("spec.maxReplicas(%d) is less than spec.minReplicas(%d)",
			*maxReplicas, *minReplicas))
	}
	if minReplicas != nil && *minReplicas < minValue {
		allErrs = append(allErrs, fmt.Sprintf("spec.minReplicas(%d) is less than the smallest allowed value(%d)",
			*minReplicas, minValue))
	}
	if maxReplicas != nil && *maxReplicas < minValue {
		allErrs = append(allErrs, fmt.Sprintf("spec.maxReplicas(%d) is less than the smallest allowed value(%d)",
			*maxReplicas, minValue))
	}
	return returnAllErrs("invalid values", allErrs)
}

func (spec *FunctionSpec) validateLabels(_ *ValidationConfig) error {
	var templateLabels map[string]string
	if spec.Template != nil {
		templateLabels = spec.Template.Labels
	}

	errs := field.ErrorList{}
	errs = append(errs, validateFunctionLabels(spec.Labels, "spec.labels")...)
	errs = append(errs, validateFunctionLabels(templateLabels, "spec.template.labels")...)
	errs = append(errs, validateLabelConflicts(templateLabels, "spec.template.labels", spec.Labels, "spec.labels")...)

	return errs.ToAggregate()
}

func validateLabelConflicts(lab1 map[string]string, path1 string, lab2 map[string]string, path2 string) field.ErrorList {
	errs := field.ErrorList{}
	if labels.Conflicts(lab1, lab2) {
		fieldPath1 := field.NewPath(path1)
		fieldPath2 := field.NewPath(path2)
		errs = append(errs, field.Invalid(fieldPath1, fieldPath2, "conflict between labels"))
	}
	return errs
}

func validateFunctionLabels(labels map[string]string, path string) field.ErrorList {
	errs := field.ErrorList{}

	fieldPath := field.NewPath(path)
	errs = append(errs, v1validation.ValidateLabels(labels, fieldPath)...)
	errs = append(errs, validateFunctionLabelsByOwnGroup(labels, fieldPath)...)

	return errs
}

func validateFunctionLabelsByOwnGroup(labels map[string]string, fieldPath *field.Path) field.ErrorList {
	forbiddenPrefix := FunctionGroup + "/"
	errorMessage := fmt.Sprintf("label from domain %s is not allowed", FunctionGroup)
	allErrs := field.ErrorList{}
	for k := range labels {
		if strings.HasPrefix(k, forbiddenPrefix) {
			allErrs = append(allErrs, field.Invalid(fieldPath, k, errorMessage))
		}
	}
	return allErrs
}

func (spec *FunctionSpec) validateAnnotations(_ *ValidationConfig) error {
	fieldPath := field.NewPath("spec.annotations")
	errs := validation.ValidateAnnotations(spec.Annotations, fieldPath)

	return errs.ToAggregate()
}

func (spec *FunctionSpec) validateSecretMounts(_ *ValidationConfig) error {
	var allErrs []string
	secretMounts := spec.SecretMounts
	for _, secretMount := range secretMounts {
		allErrs = append(allErrs,
			utilvalidation.IsDNS1123Subdomain(secretMount.SecretName)...)
	}

	if !secretNamesAreUnique(secretMounts) {
		allErrs = append(allErrs, "secretNames should be unique")
	}

	if !secretMountPathAreNotEmpty(secretMounts) {
		allErrs = append(allErrs, "mountPath should not be empty")
	}

	return returnAllErrs("invalid spec.secretMounts", allErrs)
}

func secretNamesAreUnique(secretMounts []SecretMount) bool {
	uniqueSecretNames := make(map[string]bool)
	for _, secretMount := range secretMounts {
		uniqueSecretNames[secretMount.SecretName] = true
	}
	return len(uniqueSecretNames) == len(secretMounts)
}

func secretMountPathAreNotEmpty(secretMounts []SecretMount) bool {
	for _, secretMount := range secretMounts {
		if secretMount.MountPath == "" {
			return false
		}
	}
	return true
}

type property struct {
	name  string
	value string
}

func (spec *FunctionSpec) validateRepository(_ *ValidationConfig) error {
	return validateIfMissingFields([]property{
		{name: "spec.source.gitRepository.baseDir", value: spec.Source.GitRepository.BaseDir},
		{name: "spec.source.gitRepository.reference", value: spec.Source.GitRepository.Reference},
	}...)
}

func urlIsSSH(repoURL string) bool {
	exp, err := regexp.Compile(`((git|ssh?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(/)?`)
	if err != nil {
		panic(err)
	}

	return exp.MatchString(repoURL)
}
