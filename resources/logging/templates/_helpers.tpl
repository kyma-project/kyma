{{/* vim: set filetype=mustache: */}}
{{/* Expand the name of the chart. This is suffixed with -alertmanager, which means subtract 13 from longest 63 available */}}
{{- define "kube-logging-stack.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 50 | trimSuffix "-" -}}
{{- end }}

{{/* Create chart name and version as used by the chart label. */}}
{{- define "kube-logging-stack.chartref" -}}
{{- replace "+" "_" .Chart.Version | printf "%s-%s" .Chart.Name -}}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kube-logging-stack.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kube-logging-stack.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/* Generate basic labels */}}
{{- define "kube-logging-stack.labels" }}
chart: {{ template "kube-logging-stack.chartref" . }}
release: {{ $.Release.Name | quote }}
{{- if .Values.commonLabels}}
{{ toYaml .Values.commonLabels }}
{{- end }}
helm.sh/chart: {{ include "kube-logging-stack.chartref" . }}
{{ include "kube-logging-stack.selectorLabels" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}


