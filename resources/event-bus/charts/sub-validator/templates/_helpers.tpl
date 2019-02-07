{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the sub-validator subchart.
*/}}
{{- define "sub-validator.name" -}}
{{- printf "sub-validator" -}}
{{- end -}}

{{/*
Expand the name of the sub-validator subchart.
*/}}
{{- define "sub-validator.fullname" -}}
{{- printf "%s-%s" .Release.Name "sub-validator" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- /*
Credit: @technosophos
https://github.com/technosophos/common-chart/
sub-validator.labels.standard prints the standard Helm labels.
The standard labels are frequently used in metadata.
*/ -}}
{{- define "sub-validator.labels.standard" -}}
app: {{ template "sub-validator.name" . }}
heritage: {{ .Release.Service | quote }}
release: {{ .Release.Name | quote }}
chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end -}}
