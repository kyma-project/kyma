{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "istio.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "istio.fullname" -}}
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
{{- define "istio.chart" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a fully qualified configmap name.
*/}}
{{- define "istio.configmap.fullname" -}}
{{- printf "%s-%s" .Release.Name "istio-mesh-config" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Configmap checksum.
*/}}
{{- define "istio.configmap.checksum" -}}
{{- print $.Template.BasePath "/configmap.yaml" | sha256sum -}}
{{- end -}}

{{/*
Create a URL for container images, without version number
*/}}
{{- define "shortimageurl" -}}
{{- $registry := default $.reg.path $.img.containerRegistryPath -}}
{{- if hasKey $.img "directory" -}}
{{- printf "%s/%s/%s" $registry $.img.directory $.img.name -}}
{{- else -}}
{{- printf "%s/%s" $registry $.img.name -}}
{{- end -}}
{{- end -}}

{{/*
Create a URL for container images
*/}}
{{- define "imageurl" -}}
{{- $registry := default $.reg.path $.img.containerRegistryPath -}}
{{- $path := ternary (print $registry) (print $registry "/" $.img.directory) (empty $.img.directory) -}}
{{- $version := ternary (print ":" $.img.version) (print "@sha256:" $.img.sha) (empty $.img.sha) -}}
{{- print $path "/" $.img.name $version -}}
{{- end -}}
