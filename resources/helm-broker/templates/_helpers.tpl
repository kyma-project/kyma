{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "name" -}}
{{- default .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "fullname" -}}
{{- printf "%s" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a URL for container images
*/}}
{{- define "imageurl" -}}
{{- $registry := default $.reg.path $.img.containerRegistryPath -}}
{{- if hasKey $.img "directory" -}}
{{- printf "%s/%s/%s:%s" $registry $.img.directory $.img.name $.img.version -}}
{{- else -}}
{{- printf "%s/%s:%s" $registry $.img.name $.img.version -}}
{{- end -}}
{{- end -}}
