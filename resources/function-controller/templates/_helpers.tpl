{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "fullname" -}}
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
{{- define "chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Expand the name of the tests resources.
*/}}
{{- define "tests.name" -}}
{{- printf "%s-tests" (include "name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Expand the fullname of the tests resources.
*/}}
{{- define "tests.fullname" -}}
{{- printf "%s-tests" (include "fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Renders a value that contains template.
Usage:
{{- include "tplValue" ( dict "value" .Values.path.to.the.Value "context" $ ) }}
*/}}
{{- define "tplValue" -}}
    {{- if typeIs "string" .value }}
        {{- tpl .value .context }}
    {{- else }}
        {{- tpl (.value | toYaml) .context }}
    {{- end }}
{{- end -}}

{{/*
Renders a proper env in container
Usage:
{{ include "createEnv" ( dict "name" "APP_FOO_BAR" "value" .Values.path.to.the.Value "context" $ ) }}
*/}}
{{- define "createEnv" -}}
{{- if and .name .value }}
{{- printf "- name: %s" .name -}}
{{- include "tplValue" ( dict "value" .value "context" .context ) | nindent 2 }}
{{- end }}
{{- end -}}

{{/*
Expand the name of the chart.
*/}}
{{- define "registry-name" -}}
{{- "docker-registry" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "registry-fullname" -}}
{{- "function-controller-docker-registry" -}}
{{- end -}}

