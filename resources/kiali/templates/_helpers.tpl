{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "kiali.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kiali.fullname" -}}
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
Common labels
*/}}
{{- define "kiali.labels" -}}
helm.sh/chart: {{ include "kiali.chart" . }}
{{ include "kiali.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kiali.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kiali.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "kiali.chart" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "kiali.kcproxy.groups" -}}
{{- if .Values.kcproxy.config.resources.useKymaGroups }}
{{- printf "|groups=%s,%s,%s,%s" .Values.global.kymaRuntime.adminGroup .Values.global.kymaRuntime.operatorGroup .Values.global.kymaRuntime.developerGroup .Values.global.kymaRuntime.namespaceAdminGroup -}}
{{- else if .Values.kcproxy.config.resources.groups }}
{{- printf "|groups=%s" .Values.kcproxy.config.resources.groups }}
{{- end }}
{{- end -}}

{{- define "kiali.kcproxy.methods" -}}
{{- if .Values.kcproxy.config.resources.methods }}
{{- printf "|methods=%s" .Values.kcproxy.config.resources.methods }}
{{- end }}
{{- end -}}

{{- define "kiali.kcproxy.roles" -}}
{{- if .Values.kcproxy.config.resources.roles }}
{{- printf "|roles=%s" .Values.kcproxy.config.resources.roles }}
{{- end }}
{{- end -}}