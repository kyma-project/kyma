{{/*
Expand the name of the chart.
*/}}
{{- define "nats-controller.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "nats-controller.fullname" -}}
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
{{- define "nats-controller.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "nats-controller.labels" -}}
helm.sh/chart: {{ include "nats-controller.chart" . }}
{{ include "nats-controller.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "nats-controller.selectorLabels" -}}
app.kubernetes.io/name: {{ include "nats-controller.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
kyma-project.io/dashboard: eventing
{{- end }}

{{/*
Nats server service Name
*/}}
{{- define "nats-controller.natsServer.url" -}}
{{- printf "%s-nats.%s.svc.cluster.local" .Release.Name .Release.Namespace | trunc 63 | trimSuffix "-" }}
{{- end }}
