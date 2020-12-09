{{- define "knative-eventing.fullname" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- /*
knative-eventing.labels.standard prints the standard labels.

Standard labels are used in metadata.
*/ -}}
{{- define "knative-eventing.labels.standard" -}}
app.kubernetes.io/name: {{ template "knative-eventing.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}
