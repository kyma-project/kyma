{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the nats-streaming subchart.
*/}}
{{- define "nats-streaming.fullname" -}}
{{- printf "%s-%s" .Release.Name "nats-streaming" | trunc 63 | trimSuffix "-" -}}
{{- end -}}