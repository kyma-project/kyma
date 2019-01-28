{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the nats-streaming subchart.
*/}}

{{- /*
Credit: @technosophos
https://github.com/technosophos/common-chart/
nats-streaming.labels.standard prints the standard Helm labels.
The standard labels are frequently used in metadata.
*/ -}}
{{- define "nats-streaming.labels.standard" -}}
app: {{ .Values.global.natsStreaming.name }}
heritage: {{ .Release.Service | quote }}
release: {{ .Release.Name | quote }}
chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end -}}
