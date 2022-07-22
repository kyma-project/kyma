{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "hydra-maester.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "hydra-maester.fullname" -}}
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
{{- define "hydra-maester.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "hydra-maester.labels" -}}
app.kubernetes.io/name: {{ include "hydra-maester.name" . }}
helm.sh/chart: {{ include "hydra-maester.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Get Hydra name
*/}}
{{- define "hydra-maester.getHydraName" -}}
{{- $fullName := include "hydra-maester.fullname" . -}}
{{- $nameParts := split "-" $fullName }}
{{- if eq $nameParts._0 $nameParts._1 -}}
{{- printf "%s" $nameParts._0 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" $nameParts._0 $nameParts._1 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Get Hydra admin service name
*/}}
{{- define "hydra-maester.adminService" -}}
{{- if .Values.hydraFullnameOverride -}}
{{- printf "%s-admin"  .Values.hydraFullnameOverride -}}
{{- else if contains "hydra" .Release.Name -}}
{{- printf "%s-admin" .Release.Name -}}
{{- else -}}
{{- printf "%s-%s-admin" .Release.Name "hydra" -}}
{{- end -}}
{{- end -}}

{{/*
Get Hydra secret name
*/}}
{{- define "hydra-maester.hydraSecret" -}}
{{- if hasKey .Values.config "hydraSecret" -}}
{{- printf "%s" .Values.config.hydraSecret -}}
{{- else -}}
{{- $hydra := include "hydra-maester.getHydraName" . -}}
{{- printf "%s" $hydra -}}
{{- end -}}
{{- end -}}
