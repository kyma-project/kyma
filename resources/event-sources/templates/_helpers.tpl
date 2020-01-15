{{- define "controller-manager.fullname" -}}
{{- printf "%s-%s" .Release.Name .Values.global.controllerManager.name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- /*
controller-manager.labels.standard prints the standard labels.

Standard labels are used in metadata.
*/ -}}
{{- define "controller-manager.labels.standard" -}}
app.kubernetes.io/name: {{ template "controller-manager.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: kyma
kyma-project.io/event-mesh: true
{{- end -}}

{{- /*
controller-manager.labels.selectors prints the labels used in selectors.

Selectors use a subset of the standard labels.
*/ -}}
{{- define "controller-manager.labels.selectors" -}}
app.kubernetes.io/name: {{ template "controller-manager.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
