package v1alpha2

import (
	"fmt"
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
	MinValue int32 `envconfig:"default=1"`
}

type MinFunctionResourcesValues struct {
	MinRequestCPU    string `envconfig:"default=10m"`
	MinRequestMemory string `envconfig:"default=16Mi"`
}

type MinBuildJobResourcesValues struct {
	MinRequestCPU    string `envconfig:"default=200m"`
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

func (fn *Function) getBasicValidations() []validationFunction {
	return []validationFunction{
		fn.validateObjectMeta,
		fn.Spec.validateRuntime,
		fn.Spec.validateEnv,
		fn.Spec.validateLabels,
		fn.Spec.validateReplicas,
		fn.Spec.validateFunctionResources,
		fn.Spec.validateBuildResources,
		fn.Spec.validateSources,
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

var ErrInvalidGitRepositoryAuthType = fmt.Errorf("invalid git reposiotry authentication type")

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
	case Python39, NodeJs12, NodeJs14, NodeJs16:
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
	if spec.Replicas != nil && spec.ScaleConfig != nil {
		allErrs = append(allErrs, "spec.replicas and spec.scaleConfig are use at the same time")
	}
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
	var labels map[string]string
	if spec.Templates != nil && spec.Templates.FunctionPod != nil && spec.Templates.FunctionPod.Metadata != nil && spec.Templates.FunctionPod.Metadata.Labels != nil {
		labels = spec.Templates.FunctionPod.Metadata.Labels
	}
	fieldPath := field.NewPath("spec.labels")

	errs := v1validation.ValidateLabels(labels, fieldPath)
	return errs.ToAggregate()
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

func (spec *FunctionSpec) validateTemplates(vc *ValidationConfig) error {
	if spec.Templates == nil {
		return nil
	}

	return runValidations(vc, spec.validateVolumes)
}

func (spec *FunctionSpec) validateVolumes(_ *ValidationConfig) error {
	templates := []*PodTemplate{
		spec.Templates.BuildJob,
		spec.Templates.FunctionPod,
	}
	allErrs := []string{}
	for _, t := range templates {
		allErrs = append(allErrs, t.validateTemplateVolumes()...)
	}
	return returnAllErrs("invalid volume spec", allErrs)
}

func (template *PodTemplate) validateTemplateVolumes() []string {
	allErrs := []string{}
	if !template.HasVolumes() && !template.HasVolumeMounts() {
		return nil
	}
	if !template.HasVolumes() {
		return append(allErrs, "volumes can't be empty if the volumeMounts are set.")
	}
	if !template.HasVolumeMounts() {
		return append(allErrs, "volumeMounts can't be empty if the volumes are set.")
	}

	allErrs = append(allErrs, validateVolumes(template)...)
	allErrs = append(allErrs, validateVolumeMounts(template)...)

	return allErrs
}

func validateVolumes(template *PodTemplate) []string {
	vols := template.Volumes
	allErrs := []string{}
	for _, vol := range vols {
		if vol.VolumeSource.Secret == nil && vol.VolumeSource.ConfigMap == nil {
			allErrs = append(allErrs, fmt.Sprintf("invalid volume source for volume [%s], only Secret and ConfigMap sources are supported.", vol.Name))
		}
		if vol.Name == "" {
			allErrs = append(allErrs, "volume name can't be empty.")
		}
	}
	return allErrs
}

func validateVolumeMounts(template *PodTemplate) []string {
	vols := template.Volumes
	volMounts := template.Spec.VolumeMounts
	allErrs := []string{}
	if volMounts == nil {
		allErrs = append(allErrs, "volumeMounts must be set if volumes are used.")
	}
	if len(vols) != len(volMounts) {
		allErrs = append(allErrs, "number of volumes and volumeMounts must be the same.")
	}

	volNames := map[string]bool{}
	for _, vol := range vols {
		volNames[vol.Name] = false
	}

	for _, volMount := range volMounts {
		_, exits := volNames[volMount.Name]
		if !exits {
			allErrs = append(allErrs, fmt.Sprintf("volume spec [%s] for volumeMount [%s] is not set.", volMount.Name, volMount.Name))
			continue
		}
		volNames[volMount.Name] = true
		if volMount.MountPath == "" {
			allErrs = append(allErrs, fmt.Sprintf("mountPath for volumeMount [%s] can't be empty.", volMount.Name))
		}
	}

	for vol, hasMount := range volNames {
		if !hasMount {
			allErrs = append(allErrs, fmt.Sprintf("volumeMount spec [%s] for volume [%s] is not set.", vol, vol))
		}
	}
	return allErrs
}
