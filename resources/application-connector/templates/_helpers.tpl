{{- define "recreate-service-instances.name" -}}
{{- printf "recreate-service-instances" -}}
{{- end -}}

{{- define "recreate-service-instances.fullname" -}}
{{- printf "%s-%s" .Release.Name "recreate-service-instances" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "recreate-service-instances-metrics.name" -}}
{{- printf "%s-%s" .Release.Name "recreate-service-instances-metrics" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- /*
recreate-service-instances.labels.standard prints the standard labels.

Standard labels are used in metadata.
*/ -}}
{{- define "recreate-service-instances.labels.standard" -}}
app.kubernetes.io/name: {{ template "recreate-service-instances.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: kyma
{{- end -}}
