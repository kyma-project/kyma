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

{{/* vim: set filetype=mustache: */}}
