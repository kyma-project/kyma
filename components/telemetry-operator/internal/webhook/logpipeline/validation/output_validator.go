package validation

import (
	"fmt"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
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
	section, err := fluentbit.ParseSection(logPipeline.Spec.Output.Custom)
	if err != nil {
		return err
	}

	if _, hasKey := section["storage.total_limit_size"]; hasKey {
		return fmt.Errorf("log pipeline '%s' contains forbidden configuration key 'storage.total_limit_size'", logPipeline.Name)
	}

	return nil
}
