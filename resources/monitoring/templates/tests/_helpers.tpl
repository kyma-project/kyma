{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s" $.Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "prometheus-test.labels" -}}
helm.sh/chart: {{ include "prometheus-test.chart" . }}
{{ include "prometheus-test.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "prometheus-test.selectorLabels" -}}
app.kubernetes.io/name: {{ include "prometheus-test.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}