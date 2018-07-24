{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the push subchart.
*/}}
{{- define "push.name" -}}
{{- printf "push" -}}
{{- end -}}

{{/*
Expand the name of the push subchart.
*/}}
{{- define "push.fullname" -}}
{{- printf "%s-%s" .Release.Name "push" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- /*
Credit: @technosophos
https://github.com/technosophos/common-chart/
push.labels.standard prints the standard Helm labels.
The standard labels are frequently used in metadata.
*/ -}}
{{- define "push.labels.standard" -}}
app: {{ template "push.name" . }}
heritage: {{ .Release.Service | quote }}
release: {{ .Release.Name | quote }}
chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end -}}
