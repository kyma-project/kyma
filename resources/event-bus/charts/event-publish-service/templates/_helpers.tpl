{{- define "event-publish-service.name" -}}
{{- printf "event-publish-service" -}}
{{- end -}}

{{- define "event-publish-service.fullname" -}}
{{- printf "%s-%s" .Release.Name "event-publish-service" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "event-publish-service-metrics.name" -}}
{{- printf "%s-%s" .Release.Name "event-publish-service-metrics" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- /*
event-publish-service.labels.standard prints the standard labels.

Standard labels are used in metadata.
*/ -}}
{{- define "event-publish-service.labels.standard" -}}
app.kubernetes.io/name: {{ template "event-publish-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: kyma
{{- end -}}

{{- /*
event-publish-service.labels.selectors prints the labels used in selectors.

Selectors use a subset of the standard labels.
*/ -}}
{{- define "event-publish-service.labels.selectors" -}}
app.kubernetes.io/name: {{ template "event-publish-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- /*
event-publish-service.labels.kyma prints Kyma-specific labels.

Kyma labels are set on various objects to integrate with other technical components (monitoring, ...).
*/ -}}
{{- define "event-publish-service.labels.kyma" -}}
kyma-grafana: {{ .Values.monitoring.grafana }}
kyma-alerts: {{ .Values.monitoring.alerts }}
{{- end -}}
