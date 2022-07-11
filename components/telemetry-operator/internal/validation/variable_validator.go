package validation

import (
	"context"
	"fmt"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockery --name VariablesValidator --filename variables_validator.go
type VariablesValidator interface {
	Validate(context context.Context, logPipeline *telemetryv1alpha1.LogPipeline, logPipelines *telemetryv1alpha1.LogPipelineList) error
}

type variablesValidator struct {
	client client.Client
}

func NewVariablesValidator(client client.Client) VariablesValidator {
	return &variablesValidator{
		client: client,
	}
}

func (v *variablesValidator) Validate(context context.Context, logPipeline *telemetryv1alpha1.LogPipeline, logPipelines *telemetryv1alpha1.LogPipelineList) error {
	if len(logPipeline.Spec.Variables) == 0 {
		return nil
	}
	for _, l := range logPipelines.Items {
		if l.Name != logPipeline.Name {
			for _, variable := range l.Spec.Variables {

				err := findConflictingVariables(logPipeline, variable, l.Name)
				if err != nil {
					return err
				}
			}
		}
	}

	for _, variable := range logPipeline.Spec.Variables {
		if validateMandatoryFieldsAreEmpty(variable) {
			return fmt.Errorf("mandatory field variable name or secretKeyRef name or secretKeyRef namespace or secretKeyRef key cannot be empty")
		}
	}
	return nil
}

func validateMandatoryFieldsAreEmpty(vr telemetryv1alpha1.VariableReference) bool {
	secretKey := vr.ValueFrom.SecretKey
	return len(vr.Name) == 0 || len(secretKey.Key) == 0 || len(secretKey.Namespace) == 0 || len(secretKey.Name) == 0

}

func findConflictingVariables(logPipeLine *telemetryv1alpha1.LogPipeline, vr telemetryv1alpha1.VariableReference, existingPipelineName string) error {
	for _, v := range logPipeLine.Spec.Variables {
		if v.Name == vr.Name {
			return fmt.Errorf("variable with name '%s' has been previously used in pipeline '%s'", v.Name, existingPipelineName)
		}
	}
	return nil
}
