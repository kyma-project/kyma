{{- define "event-mesh-dashboard.fullname" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- /*
event-mesh-dashboard.labels.standard prints the standard labels.

Standard labels are used in metadata.
*/ -}}
{{- define "event-mesh-dashboard.labels.standard" -}}
app.kubernetes.io/name: {{ template "event-mesh-dashboard.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: kyma
kyma-project.io/dashboard: event-mesh
{{- end -}}
