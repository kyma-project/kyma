package validation

import (
	"fmt"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

//go:generate mockery --name OutputValidator --filename output_validator.go
type OutputValidator interface {
	Validate(logPipeline *telemetryv1alpha1.LogPipeline) error
}

type outputValidator struct {
}

func NewOutputValidator() OutputValidator {
	return &outputValidator{}
}

func (v *outputValidator) Validate(logPipeline *telemetryv1alpha1.LogPipeline) error {
	section, err := config.ParseCustomSection(logPipeline.Spec.Output.Custom)
	if err != nil {
		return err
	}

	if section.ContainsKey("storage.total_limit_size") {
		return fmt.Errorf("log pipeline '%s' contains forbidden configuration key 'storage.total_limit_size'", logPipeline.Name)
	}

	return nil
}
