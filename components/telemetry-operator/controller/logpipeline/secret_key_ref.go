package logpipeline

import (
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

type fieldDescriptor struct {
	jsonPath     string
	secretKeyRef telemetryv1alpha1.SecretKeyRef
}

func listSecretRefFields(pipeline *telemetryv1alpha1.LogPipeline) []fieldDescriptor {
	var result []fieldDescriptor

	for _, v := range pipeline.Spec.Variables {
		if !v.ValueFrom.IsSecretKeyRef() {
			continue
		}

		result = append(result, fieldDescriptor{
			jsonPath:     "spec.variables.name",
			secretKeyRef: *v.ValueFrom.SecretKeyRef,
		})
	}

	output := pipeline.Spec.Output
	if !output.IsHTTPDefined() {
		return result
	}
	if output.HTTP.Host.ValueFrom != nil && output.HTTP.Host.ValueFrom.IsSecretKeyRef() {
		result = append(result, fieldDescriptor{
			jsonPath:     "spec.output.http.host",
			secretKeyRef: *output.HTTP.Host.ValueFrom.SecretKeyRef,
		})
	}
	if output.HTTP.User.ValueFrom != nil && output.HTTP.User.ValueFrom.IsSecretKeyRef() {
		result = append(result, fieldDescriptor{
			jsonPath:     "spec.output.http.user",
			secretKeyRef: *output.HTTP.User.ValueFrom.SecretKeyRef,
		})
	}
	if output.HTTP.Password.ValueFrom != nil && output.HTTP.Password.ValueFrom.IsSecretKeyRef() {
		result = append(result, fieldDescriptor{
			jsonPath:     "spec.output.http.password",
			secretKeyRef: *output.HTTP.Password.ValueFrom.SecretKeyRef,
		})
	}

	return result
}

func hasSecretReference(pipeline *telemetryv1alpha1.LogPipeline, secretName, secretNamespace string) bool {
	secretRefFields := listSecretRefFields(pipeline)
	for _, field := range secretRefFields {
		if field.secretKeyRef.Name == secretName && field.secretKeyRef.Namespace == secretNamespace {
			return true
		}
	}

	return false
}
