{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "rafterFrontMatterService.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "rafterFrontMatterService.fullname" -}}
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
{{- define "rafterFrontMatterService.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the name of the service
*/}}
{{- define "rafterFrontMatterService.serviceName" -}}
{{- if .Values.service.name -}}
{{- include .Values.service.name . | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- include "rafterFrontMatterService.fullname" . -}}
{{- end -}}
{{- end -}}

{{/*
Create the name of the service monitor
*/}}
{{- define "rafterFrontMatterService.serviceMonitorName" -}}
{{- if .Values.serviceMonitor.create }}
    {{ default (include "rafterFrontMatterService.fullname" .) .Values.serviceMonitor.name }}
{{- else -}}
    {{ default "default" .Values.serviceMonitor.name }}
{{- end -}}
{{- end -}}

{{/*
Renders a value that contains template.
Usage:
{{ include "rafterFrontMatterService.tplValue" ( dict "value" .Values.path.to.the.Value "context" $ ) }}
*/}}
{{- define "rafterFrontMatterService.tplValue" -}}
    {{- if typeIs "string" .value }}
        {{- tpl .value .context }}
    {{- else }}
        {{- tpl (.value | toYaml) .context }}
    {{- end }}
{{- end -}}

{{/*
Renders a proper env in container
Usage:
{{ include "rafterFrontMatterService.createEnv" ( dict "name" "APP_FOO_BAR" "value" .Values.path.to.the.Value "context" $ ) }}
*/}}
{{- define "rafterFrontMatterService.createEnv" -}}
{{- if and .name .value }}
{{- printf "- name: %s" .name -}}
{{- include "rafterFrontMatterService.tplValue" ( dict "value" .value "context" .context ) | nindent 2 }}
{{- end }}
{{- end -}}
