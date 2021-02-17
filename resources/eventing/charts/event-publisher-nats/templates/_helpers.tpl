{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "event-publisher-nats.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "event-publisher-nats.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "event-publisher-nats.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "event-publisher-nats.labels" -}}
helm.sh/chart: {{ include "event-publisher-nats.chart" . }}
{{ include "event-publisher-nats.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "event-publisher-nats.selectorLabels" -}}
app.kubernetes.io/name: {{ include "event-publisher-nats.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}


{{/*
Create the name of the service account to use
*/}}
{{- define "event-publisher-nats.serviceAccountName" -}}
{{- default (include "event-publisher-nats.fullname" .) .Values.serviceAccount.name }}
{{- end }}

{{/*
Nats server service Name
*/}}
{{- define "nats-controller.natsServer.url" -}}
{{- printf "%s-nats.%s.svc.cluster.local" .Release.Name .Release.Namespace | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Publisher service Name
*/}}
{{- define "event-publisher-nats.serviceName" -}}
{{- printf "%s-event-publisher-proxy" .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}