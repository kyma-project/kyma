{{- define "publish-knative.name" -}}
{{- printf "publish-knative" -}}
{{- end -}}

{{- define "publish-knative.fullname" -}}
{{- printf "%s-%s" .Release.Name "publish-knative" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "publish-knative.labels.standard" -}}
app: {{ template "publish-knative.name" . }}
heritage: {{ .Release.Service | quote }}
release: {{ .Release.Name | quote }}
chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end -}}
