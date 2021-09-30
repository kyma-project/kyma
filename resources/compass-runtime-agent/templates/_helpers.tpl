{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
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

