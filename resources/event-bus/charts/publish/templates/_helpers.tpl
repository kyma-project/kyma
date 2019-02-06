{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the publish subchart.
*/}}
{{- define "publish.name" -}}
{{- printf "publish" -}}
{{- end -}}

{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the publish subchart.
*/}}
{{- define "publish.fullname" -}}
{{- printf "%s-%s" .Release.Name "publish" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- /*
Credit: @technosophos
https://github.com/technosophos/common-chart/
publish.labels.standard prints the standard Helm labels.
The standard labels are frequently used in metadata.
*/ -}}
{{- define "publish.labels.standard" -}}
app: {{ template "publish.name" . }}
heritage: {{ .Release.Service | quote }}
release: {{ .Release.Name | quote }}
chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end -}}
