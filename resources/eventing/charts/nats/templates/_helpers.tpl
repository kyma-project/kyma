{{- /*
nats.labels.standard prints the standard labels.

Standard labels are used in metadata.
*/ -}}
{{- define "nats.labels.standard" -}}
app.kubernetes.io/name: {{ .Values.global.nats.fullname }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: kyma
{{- end -}}

{{- /*
nats.labels.selectors prints the labels used in selectors.

Selectors use a subset of the standard labels.
*/ -}}
{{- define "nats.labels.selectors" -}}
app.kubernetes.io/name: {{ .Values.global.nats.fullname }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/* vim: set filetype=mustache: */}}