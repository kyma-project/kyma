{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "jaeger-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "jaeger-operator.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "jaeger-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "jaeger-operator.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "jaeger-operator.labels" -}}
helm.sh/chart: {{ include "jaeger-operator.chart" . }}
{{ include "jaeger-operator.selectorLabels" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "jaeger-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "jaeger-operator.fullname" . }}-jaeger-operator
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "jaeger-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "jaeger-operator.kcproxy.groups" -}}
{{- if .Values.kcproxy.config.resources.useKymaGroups }}
{{- printf "|groups=%s,%s,%s,%s" .Values.global.kymaRuntime.adminGroup .Values.global.kymaRuntime.operatorGroup .Values.global.kymaRuntime.developerGroup .Values.global.kymaRuntime.namespaceAdminGroup -}}
{{- else if .Values.kcproxy.config.resources.groups }}
{{- printf "|groups=%s" .Values.kcproxy.config.resources.groups }}
{{- end }}
{{- end -}}

{{- define "jaeger-operator.kcproxy.methods" -}}
{{- if .Values.kcproxy.config.resources.methods }}
{{- printf "|methods=%s" .Values.kcproxy.config.resources.methods }}
{{- end }}
{{- end -}}

{{- define "jaeger-operator.kcproxy.roles" -}}
{{- if .Values.kcproxy.config.resources.roles }}
{{- printf "|roles=%s" .Values.kcproxy.config.resources.roles }}
{{- end }}
{{- end -}}