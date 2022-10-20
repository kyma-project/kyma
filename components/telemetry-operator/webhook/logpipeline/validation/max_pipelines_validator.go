package validation

import (
	"fmt"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

//go:generate mockery --name MaxPipelinesValidator --filename max_pipeliens_validator.go
type MaxPipelinesValidator interface {
	Validate(logPipeline *telemetryv1alpha1.LogPipeline, logPipelines *telemetryv1alpha1.LogPipelineList) error
}

type maxPipelinesValidator struct {
	maxPipelines int
}

func NewMaxPipelinesValidator(maxPipelines int) MaxPipelinesValidator {
	return &maxPipelinesValidator{
		maxPipelines: maxPipelines,
	}
}

func (m maxPipelinesValidator) Validate(logPipeline *telemetryv1alpha1.LogPipeline, logPipelines *telemetryv1alpha1.LogPipelineList) error {
	if m.maxPipelines > 0 && m.isNewPipeline(logPipeline, logPipelines) && len(logPipelines.Items) >= m.maxPipelines {
		return fmt.Errorf("the maximum number of log pipelines is %d", m.maxPipelines)
	}
	return nil
}

func (maxPipelinesValidator) isNewPipeline(logPipeline *telemetryv1alpha1.LogPipeline, logPipelines *telemetryv1alpha1.LogPipelineList) bool {
	for _, pipeline := range logPipelines.Items {
		if pipeline.Name == logPipeline.Name {
			return false
		}
	}
	return true
}
