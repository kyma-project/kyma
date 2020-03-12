{{- define "mesh-namespaces-migration.name" -}}
{{- printf "mesh-namespaces-migration" -}}
{{- end -}}

{{- define "mesh-namespaces-migration.fullname" -}}
{{- printf "%s-%s" .Release.Name "mesh-namespaces-migration" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "mesh-namespaces-migration-metrics.name" -}}
{{- printf "%s-%s" .Release.Name "mesh-namespaces-migration-metrics" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- /*
mesh-namespaces-migration.labels.standard prints the standard labels.

Standard labels are used in metadata.
*/ -}}
{{- define "mesh-namespaces-migration.labels.standard" -}}
app.kubernetes.io/name: {{ template "mesh-namespaces-migration.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: kyma
{{- end -}}
