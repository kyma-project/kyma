{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Define template equivalent to "name" template, but used for reposerver deployment/service
*/}}
{{- define "reposerver-name" -}}
{{- $rName := printf "%s-%s" .Chart.Name "reposerver" -}}
{{- default $rName  .Values.ReposerverNameOverride  | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "fullname" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Define template equivalent to "fullname" template, but used for reposerver deployment/service
*/}}
{{- define "reposerver-fullname" -}}
{{- printf "%s-%s-%s" .Release.Name .Chart.Name "reposerver" | trunc 63 | trimSuffix "-" -}}
{{- end -}}
