{{- define "subscription-controller-knative.name" -}}
{{- printf "subscription-controller-knative" -}}
{{- end -}}

{{- define "subscription-controller-knative.fullname" -}}
{{- printf "%s-%s" .Release.Name "subscription-controller-knative" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "subscription-controller-knative.labels.standard" -}}
app: {{ template "subscription-controller-knative.name" . }}
heritage: {{ .Release.Service | quote }}
release: {{ .Release.Name | quote }}
chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end -}}
