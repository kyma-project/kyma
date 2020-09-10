{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- /*
application-broker-eventing-migration.labels.standard prints the standard labels.
Standard labels are used in metadata.
*/ -}}
{{- define "application-broker-eventing-migration.name" -}}
{{- printf "application-broker-eventing-migration" -}}
{{- end -}}

{{- define "application-broker-eventing-migration.fullname" -}}
{{- printf "%s-%s" .Release.Name "application-broker-eventing-migration" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "application-broker-eventing-migration.labels.standard" -}}
app.kubernetes.io/name: {{ template "application-broker-eventing-migration.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: kyma
{{- end -}}
