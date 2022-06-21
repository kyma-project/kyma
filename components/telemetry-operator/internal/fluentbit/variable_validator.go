package fluentbit

import (
	"context"
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockery --name PluginValidator --filename plugin_validator.go
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
	for _, l := range logPipelines.Items {
		if l.Name != logPipeline.Name {
			for _, variable := range l.Spec.Variables {
				err := validateSecretRefs(logPipeline, variable, l.Name)
				if err != nil {
					return err
				}

			}
		}
	}

	for _, variable := range logPipeline.Spec.Variables {
		err := v.validateSecretKeysExist(context, variable.ValueFrom.SecretKey)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *variablesValidator) validateSecretKeysExist(ctx context.Context, secretKeyRef telemetryv1alpha1.SecretKeyRef) error {
	var referencedSecret corev1.Secret
	if err := v.client.Get(ctx, types.NamespacedName{Name: secretKeyRef.Name, Namespace: secretKeyRef.Namespace}, &referencedSecret); err != nil {
		return fmt.Errorf("failed reading secret %s from namespace %s", secretKeyRef.Name, secretKeyRef.Namespace)
	}

	if _, ok := referencedSecret.Data[secretKeyRef.Key]; !ok {
		return fmt.Errorf("failed to find the key: %s in the secret %s", secretKeyRef.Key, secretKeyRef.Name)
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
