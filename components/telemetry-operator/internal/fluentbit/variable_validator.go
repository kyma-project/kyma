package fluentbit

import (
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
)

func ValidateVariables(logPipeline *telemetryv1alpha1.LogPipeline, logPipelines *telemetryv1alpha1.LogPipelineList) error {

	for _, l := range logPipelines.Items {
		if l.Name != logPipeline.Name {
			for _, v := range l.Spec.Variables {
				err := validateSecretRefs(logPipeline, v, l.Name)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateSecretRefs(logPipeLine *telemetryv1alpha1.LogPipeline, vr telemetryv1alpha1.VariableReference, existingPipelineName string) error {
	for _, v := range logPipeLine.Spec.Variables {
		if v.Name == vr.Name {
			return fmt.Errorf("vairable with name '%s' has a been previously used in pipeline '%s'", v.Name, existingPipelineName)
		}
	}
	return nil
}
