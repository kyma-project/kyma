{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "loki.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "loki.fullname" -}}
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
{{- define "loki.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the name of the service account
*/}}
{{- define "loki.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "loki.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "loki.selectorLabels" -}}
app.kubernetes.io/name: {{ include "fluent-bit.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "loki.labels" -}}
helm.sh/chart: {{ include "loki.chart" . }}
{{ include "loki.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Create the app name of loki clients. Defaults to the same logic as "loki.fullname", and default client expects "promtail".
*/}}
{{- define "client.name" -}}
{{- if .Values.client.name -}}
{{- .Values.client.name -}}
{{- else if .Values.client.fullnameOverride -}}
{{- .Values.client.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default "promtail" .Values.client.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "kyma.auth.groups" -}}
{{- if .Values.kyma.auth.useKymaGroups }}
{{- printf "%s,%s,%s,%s" .Values.global.kymaRuntime.adminGroup .Values.global.kymaRuntime.operatorGroup .Values.global.kymaRuntime.developerGroup .Values.global.kymaRuntime.namespaceAdminGroup -}}
{{- else if .Values.kyma.auth.groups }}
{{- printf "%s" .Values.kyma.auth.groups }}
{{- end }}
{{- end -}}
