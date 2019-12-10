{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "rafter.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "rafter.fullname" -}}
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
{{- define "rafter.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the name of the service account
*/}}
{{- define "rafter.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "rafter.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Create the name of the rbac cluster role
*/}}
{{- define "rafter.rbacClusterRoleName" -}}
{{- if .Values.rbac.clusterScope.create -}}
    {{ default (include "rafter.fullname" .) .Values.rbac.clusterScope.role.name }}
{{- else -}}
    {{ default "default" .Values.rbac.clusterScope.role.name  }}
{{- end -}}
{{- end -}}

{{/*
Create the name of the rbac cluster role binding
*/}}
{{- define "rafter.rbacClusterRoleBindingName" -}}
{{- if .Values.rbac.clusterScope.create -}}
    {{ default (include "rafter.fullname" .) .Values.rbac.clusterScope.roleBinding.name }}
{{- else -}}
    {{ default "default" .Values.rbac.clusterScope.roleBinding.name }}
{{- end -}}
{{- end -}}

{{/*
Create the name of the rbac role
*/}}
{{- define "rafter.rbacRoleName" -}}
{{- if .Values.rbac.namespaced.create -}}
    {{ default (include "rafter.fullname" .) .Values.rbac.namespaced.role.name }}
{{- else -}}
    {{ default "default" .Values.rbac.namespaced.role.name }}
{{- end -}}
{{- end -}}

{{/*
Create the name of the rbac role binding
*/}}
{{- define "rafter.rbacRoleBindingName" -}}
{{- if .Values.rbac.namespaced.create -}}
    {{ default (include "rafter.fullname" .) .Values.rbac.namespaced.roleBinding.name }}
{{- else -}}
    {{ default "default" .Values.rbac.namespaced.roleBinding.name }}
{{- end -}}
{{- end -}}

{{/*
Create the name of the config map with webhooks
*/}}
{{- define "rafter.webhooksConfigMapName" -}}
{{- if .Values.webhooksConfigMap.create -}}
    {{ default (include "rafter.fullname" .) .Values.webhooksConfigMap.name }}
{{- else -}}
    {{ default "default" .Values.webhooksConfigMap.name }}
{{- end -}}
{{- end -}}

{{/*
Create the name of the metrics service
*/}}
{{- define "rafter.metricsServiceName" -}}
{{- if .Values.metrics.enabled -}}
    {{ default (include "rafter.fullname" .) .Values.metrics.service.name }}
{{- else -}}
    {{ default "default" .Values.metrics.service.name }}
{{- end -}}
{{- end -}}

{{/*
Create the name of the service monitor
*/}}
{{- define "rafter.serviceMonitorName" -}}
{{- if and .Values.metrics.enabled .Values.metrics.serviceMonitor.create }}
    {{ default (include "rafter.fullname" .) .Values.metrics.serviceMonitor.name }}
{{- else -}}
    {{ default "default" .Values.metrics.serviceMonitor.name }}
{{- end -}}
{{- end -}}

{{/*
Renders a value that contains template.
Usage:
{{ include "rafter.tplValue" ( dict "value" .Values.path.to.the.Value "context" $ ) }}
*/}}
{{- define "rafter.tplValue" -}}
    {{- if typeIs "string" .value }}
        {{- tpl .value .context }}
    {{- else }}
        {{- tpl (.value | toYaml) .context }}
    {{- end }}
{{- end -}}

{{/*
Renders a proper env in container
Usage:
{{ include "rafter.createEnv" ( dict "name" "APP_FOO_BAR" "value" .Values.path.to.the.Value "context" $ ) }}
*/}}
{{- define "rafter.createEnv" -}}
{{- if and .name .value }}
{{- printf "- name: %s" .name -}}
{{- include "rafter.tplValue" ( dict "value" .value "context" .context ) | nindent 2 }}
{{- end }}
{{- end -}}
