{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "gcloud-sqlproxy.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "gcloud-sqlproxy.fullname" -}}
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
{{- define "gcloud-sqlproxy.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "gcloud-sqlproxy.labels" -}}
helm.sh/chart: {{ include "gcloud-sqlproxy.chart" . }}
{{ include "gcloud-sqlproxy.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "gcloud-sqlproxy.selectorLabels" -}}
app.kubernetes.io/name: {{ include "gcloud-sqlproxy.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Generate gcp service account name
*/}}
{{- define "gcloud-sqlproxy.gcpServiceAccountName" -}}
{{ default (include "gcloud-sqlproxy.fullname" .) .Values.serviceAccountName }}
{{- end -}}

{{/*
Generate key name in the secret
*/}}
{{- define "gcloud-sqlproxy.secretKey" -}}
{{ default "credentials.json" .Values.existingSecretKey }}
{{- end -}}

{{/*
Generate the secret name
*/}}
{{- define "gcloud-sqlproxy.secretName" -}}
{{ default (include "gcloud-sqlproxy.fullname" .) .Values.existingSecret }}
{{- end -}}

{{/*
Check if any types of credentials are defined
*/}}
{{- define "gcloud-sqlproxy.hasCredentials" -}}
{{ or .Values.serviceAccountKey ( or .Values.existingSecret .Values.usingGCPController ) -}}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "gcloud-sqlproxy.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "gcloud-sqlproxy.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}
