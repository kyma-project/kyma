{{- define "subscription-controller.name" -}}
{{- printf "subscription-controller" -}}
{{- end -}}

{{- define "subscription-controller.fullname" -}}
{{- printf "%s-%s" .Release.Name "subscription-controller" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "subscription-controller-metrics.name" -}}
{{- printf "%s-%s" .Release.Name "subscription-controller-metrics" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "subscription-controller.labels.standard" -}}
app: {{ template "subscription-controller.name" . }}
heritage: {{ .Release.Service | quote }}
release: {{ .Release.Name | quote }}
chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end -}}
