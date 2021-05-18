{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "publisher-proxy.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "publisher-proxy.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "publisher-proxy.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "publisher-proxy.labels" -}}
helm.sh/chart: {{ include "publisher-proxy.chart" . }}
{{ include "publisher-proxy.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "publisher-proxy.selectorLabels" -}}
app.kubernetes.io/name: {{ include "publisher-proxy.fullname" . }}
kyma-project.io/dashboard: eventing
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "publisher-proxy.serviceAccountName" -}}
{{- default (include "publisher-proxy.fullname" .) .Values.serviceAccount.name }}
{{- end }}

{{/*
Publisher service Name
*/}}
{{- define "publisher-nats.serviceName" -}}
{{- printf "%s-publisher-proxy" .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}
