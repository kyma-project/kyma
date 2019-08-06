{{- define "event-publish-service.name" -}}
{{- printf "event-publish-service" -}}
{{- end -}}

{{- define "event-publish-service.fullname" -}}
{{- printf "%s-%s" .Release.Name "event-publish-service" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "event-publish-service-metrics.name" -}}
{{- printf "%s-%s" .Release.Name "event-publish-service-metrics" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "event-publish-service.labels.standard" -}}
app: {{ template "event-publish-service.name" . }}
heritage: {{ .Release.Service | quote }}
release: {{ .Release.Name | quote }}
chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end -}}
