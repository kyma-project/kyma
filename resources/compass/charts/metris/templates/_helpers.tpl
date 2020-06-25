{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "metris.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "metris.fullname" -}}
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

{{/* Create chart name and version as used by the chart label. */}}
{{- define "metris.chartref" -}}
{{- replace "+" "_" .Chart.Version | printf "%s-%s" .Chart.Name -}}
{{- end }}

{{/* Generate basic labels */}}
{{- define "metris.labels" -}}
chart: {{ template "metris.chartref" . }}
release: {{ .Release.Name | quote }}
heritage: {{ .Release.Service | quote }}
{{- end }}


{{- define "metris.imagePullSecrets" -}}
{{- if .Values.image.pullSecrets }}
imagePullSecrets:
{{- range .Values.image.pullSecrets }}
  - name: {{ . }}
{{- end }}
{{- else if .Values.global }}
{{- if .Values.global.imagePullSecrets }}
imagePullSecrets:
{{- range .Values.global.imagePullSecrets }}
  - name: {{ . }}
{{- end }}
{{- end -}}
{{- end -}}
{{- end -}}


{{- define "metris.image" -}}
{{- $repository := "" -}}
{{- $tag := "" -}}
{{- if .Values.global -}}
  {{- if .Values.global.images -}}
    {{- if .Values.global.images.containerRegistry -}}
      {{- $repository = printf "%s/%smetris" .Values.global.images.containerRegistry.path (default "" .Values.global.images.metris.dir) -}}
      {{- $tag = .Values.global.images.metris.version | toString -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{- if .Values.image -}}
{{- if .Values.image.repository -}}
{{- $repository = .Values.image.repository -}}
{{- end -}}
{{- if .Values.image.tag -}}
{{- $tag = .Values.image.tag | toString -}}
{{- end -}}
{{- end -}}
{{- printf "%s:%s" $repository $tag -}}
{{- end -}}