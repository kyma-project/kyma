package validation

import (
	"errors"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

//go:generate mockery --name InputValidator --filename input_validator.go
type InputValidator interface {
	Validate(logPipelineInput *telemetryv1alpha1.Input) error
}

type inputValidator struct {
}

func NewInputValidator() InputValidator {
	return &inputValidator{}
}

func (v *inputValidator) Validate(logPipelineInput *telemetryv1alpha1.Input) error {
	if logPipelineInput == nil {
		return nil
	}

	var containers = logPipelineInput.Application.Containers
	var excludeContainers = logPipelineInput.Application.ExcludeContainers
	if containers != nil && excludeContainers != nil {
		if len(containers) > 0 && len(excludeContainers) > 0 {
			return errors.New("invalid log pipeline definition: can not define both 'input.application.containers' and 'input.application.excludeContainers'")
		}
	}

	var namespaces = logPipelineInput.Application.Namespaces
	var excludeNamespaces = logPipelineInput.Application.ExcludeNamespaces
	if namespaces != nil && excludeNamespaces != nil {
		if len(namespaces) > 0 && len(excludeNamespaces) > 0 {
			return errors.New("invalid log pipeline definition: can not define both 'input.application.namespaces' and 'input.application.excludeNamespaces'")
		}
	}

	return nil
}
