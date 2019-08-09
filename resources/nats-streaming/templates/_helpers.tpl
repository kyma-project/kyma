{{- /*
nats-streaming.labels.standard prints the standard labels.

Standard labels are used in metadata.
*/ -}}
{{- define "nats-streaming.labels.standard" -}}
app.kubernetes.io/name: {{ .Values.global.natsStreaming.fullname }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: kyma
{{- end -}}

{{- /*
nats-streaming.labels.selectors prints the labels used in selectors.

Selectors use a subset of the standard labels.
*/ -}}
{{- define "nats-streaming.labels.selectors" -}}
app.kubernetes.io/name: {{ .Values.global.natsStreaming.fullname }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- /*
nats-streaming.labels.kyma prints Kyma-specific labels.

Kyma labels are set on various objects to integrate with other technical components (monitoring, ...).
*/ -}}
{{- define "nats-streaming.labels.kyma" -}}
kyma-grafana: {{ .Values.monitoring.grafana }}
kyma-alerts: {{ .Values.monitoring.alerts }}
alertcpu: {{ .Values.monitoring.alertcpu | quote }}
alertmem: {{ .Values.monitoring.alertmem | quote }}
{{- end -}}

{{/* vim: set filetype=mustache: */}}
