{{- define "fullname" -}}
{{ if eq .Release.Name .Chart.Name}}
{{- printf "%s" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{ else }}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
