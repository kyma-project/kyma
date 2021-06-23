{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "api-gateway.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "api-gateway.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "api-gateway.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "api-gateway.labels" -}}
app.kubernetes.io/name: {{ include "api-gateway.name" . }}
helm.sh/chart: {{ include "api-gateway.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Create a list of namespaced services to blocklist
*/}}
{{- define "api-gateway.serviceBlockList" -}}
{{- range $i, $e := .Values.config.serviceBlockList -}}
{{- range $e -}}
{{ printf "%s.%s," . $i -}}
{{- end }}
{{- end }}
{{- end -}}

{{/*
Create a list of domains to allowlist
*/}}
{{- define "api-gateway.domainAllowList" -}}
{{- range $domain := .Values.config.domainAllowList -}}
{{ printf "%s," $domain -}}
{{- end }}
{{- with .Values.global.ingress.domainName }}
{{- printf "%s" . -}}
{{- end }}
{{- end -}}

{{/*
Get a default domain from values if set or use the default domain name for Kyma
*/}}
{{- define "api-gateway.defaultDomain" -}}
{{ if .Values.config.defaultDomain }}
{{- printf "%s" .Values.config.defaultDomain -}}
{{ else }}
{{- printf "%s" .Values.global.ingress.domainName -}}
{{- end }}
{{- end -}}

{{- define "api-gateway.cors.allowOrigins" -}}
{{- range $i, $e := .Values.config.cors.allowOrigins -}}
{{- range $e -}}
{{ printf "%s:%s," $i . -}}
{{- end }}
{{- end }}
{{- end -}}