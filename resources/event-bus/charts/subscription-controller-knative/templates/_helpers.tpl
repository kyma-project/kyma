{{- define "subscription-controller-knative.name" -}}
{{- printf "subscription-controller-knative" -}}
{{- end -}}

{{- define "subscription-controller-knative.fullname" -}}
{{- printf "%s-%s" .Release.Name "subscription-controller-knative" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "subscription-controller-knative-metrics.name" -}}
{{- printf "%s-%s" .Release.Name "subscription-controller-knative-metrics" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "subscription-controller-knative-metrics-service.name" -}}
{{- printf "%s-%s" .Release.Name "subscription-controller-knative-metrics-service" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "subscription-controller-knative-metrics-service-monitor.name" -}}
{{- printf "%s-%s" .Release.Name "subscription-controller-knative-metrics-service-monitor" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "subscription-controller-knative-metrics-destination-rule.name" -}}
{{- printf "%s-%s" .Release.Name "subscription-controller-knative-metrics-destination-rule" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "subscription-controller-knative.labels.standard" -}}
app: {{ template "subscription-controller-knative.name" . }}
heritage: {{ .Release.Service | quote }}
release: {{ .Release.Name | quote }}
chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end -}}
