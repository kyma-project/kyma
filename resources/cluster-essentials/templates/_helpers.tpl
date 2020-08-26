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
*/}}
{{- define "fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}


{{/*
False if Gardener or local env
*/}}
{{- define "components.installCertManager" -}}
{{- if ( or (.Capabilities.APIVersions.Has "cert.gardener.cloud/v1alpha1") (.Values.global.isLocalEnv)) }}
  {{- false -}}
{{- else }}
  {{- true -}}
{{- end }}
{{- end -}}
