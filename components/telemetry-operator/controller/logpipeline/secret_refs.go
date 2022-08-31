package logpipeline

import (
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/utils/envvar"
)

type fieldDescriptor struct {
	targetSecretKey string
	secretKeyRef    telemetryv1alpha1.SecretKeyRef
}

func lookupSecretRefFields(pipeline *telemetryv1alpha1.LogPipeline) []fieldDescriptor {
	var result []fieldDescriptor

	for _, v := range pipeline.Spec.Variables {
		if !v.ValueFrom.IsSecretKeyRef() {
			continue
		}

		result = append(result, fieldDescriptor{
			targetSecretKey: v.Name,
			secretKeyRef:    *v.ValueFrom.SecretKeyRef,
		})
	}

	output := pipeline.Spec.Output
	if output.IsHTTPDefined() {
		result = appendOutputFieldIfHasSecret(result, pipeline.Name, output.HTTP.Host)
		result = appendOutputFieldIfHasSecret(result, pipeline.Name, output.HTTP.User)
		result = appendOutputFieldIfHasSecret(result, pipeline.Name, output.HTTP.Password)
	}

	if output.IsLokiDefined() {
		result = appendOutputFieldIfHasSecret(result, pipeline.Name, output.Loki.URL)
	}

	return result
}

func appendOutputFieldIfHasSecret(fields []fieldDescriptor, pipelineName string, valueType telemetryv1alpha1.ValueType) []fieldDescriptor {
	if valueType.Value == "" && valueType.ValueFrom != nil && valueType.ValueFrom.IsSecretKeyRef() {
		fields = append(fields, fieldDescriptor{
			targetSecretKey: envvar.GenerateName(pipelineName, *valueType.ValueFrom.SecretKeyRef),
			secretKeyRef:    *valueType.ValueFrom.SecretKeyRef,
		})
	}

	return fields
}

func hasSecretRef(pipeline *telemetryv1alpha1.LogPipeline, secretName, secretNamespace string) bool {
	secretRefFields := lookupSecretRefFields(pipeline)
	for _, field := range secretRefFields {
		if field.secretKeyRef.Name == secretName && field.secretKeyRef.Namespace == secretNamespace {
			return true
		}
	}

	return false
}
