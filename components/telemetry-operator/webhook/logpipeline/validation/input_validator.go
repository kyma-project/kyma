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
	if len(containers.Include) > 0 && len(containers.Exclude) > 0 {
		return errors.New("invalid log pipeline definition: can not define both 'input.application.containers.include' and 'input.application.containers.exclude'")
	}

	var namespaces = logPipelineInput.Application.Namespaces
	if (len(namespaces.Include) > 0 && len(namespaces.Exclude) > 0) ||
		(len(namespaces.Include) > 0 && namespaces.System) ||
		(len(namespaces.Exclude) > 0 && namespaces.System) {
		return errors.New("invalid log pipeline definition: can only define one of 'input.application.namespaces' selectors: 'include', 'exclude', 'system'")
	}

	return nil
}
