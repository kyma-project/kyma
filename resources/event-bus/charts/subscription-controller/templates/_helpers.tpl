{{- define "subscription-controller.name" -}}
{{- printf "subscription-controller" -}}
{{- end -}}

{{- define "subscription-controller.fullname" -}}
{{- printf "%s-%s" .Release.Name "subscription-controller" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- /*
subscription-controller.labels.standard prints the standard labels.

Standard labels are used in metadata.
*/ -}}
{{- define "subscription-controller.labels.standard" -}}
app.kubernetes.io/name: {{ template "subscription-controller.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: kyma
{{- end -}}

{{- /*
subscription-controller.labels.selectors prints the labels used in selectors.

Selectors use a subset of the standard labels.
*/ -}}
{{- define "subscription-controller.labels.selectors" -}}
app.kubernetes.io/name: {{ template "subscription-controller.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- /*
subscription-controller.labels.kyma prints Kyma-specific labels.

Kyma labels are set on various objects to integrate with other technical components (monitoring, ...).
*/ -}}
{{- define "subscription-controller.labels.kyma" -}}
kyma-grafana: {{ .Values.monitoring.grafana }}
kyma-alerts: {{ .Values.monitoring.alerts }}
{{- end -}}
