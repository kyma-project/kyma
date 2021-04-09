{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "hydra.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "hydra.fullname" -}}
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
Create chart name and version as used by the chart label.
*/}}
{{- define "hydra.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Ensure there is always a way to track down source of the deployment.
It is unlikely AppVersion will be missing, but we will fallback on the
chart's version in that case.
*/}}
{{- define "hydra.version" -}}
{{- if .Chart.AppVersion }}
{{- .Chart.AppVersion -}}
{{- else -}}
{{- printf "v%s" .Chart.Version -}}
{{- end -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "hydra.labels" -}}
"app.kubernetes.io/name": {{ include "hydra.name" . | quote }}
"app.kubernetes.io/instance": {{ .Release.Name | quote }}
"app.kubernetes.io/version": {{ include "hydra.version" . | quote }}
"app.kubernetes.io/managed-by": {{ .Release.Service | quote }}
"helm.sh/chart": {{ include "hydra.chart" . | quote }}
{{- end -}}

{{/*
Generate the dsn value
*/}}
{{- define "hydra.dsn" -}}
{{- if .Values.demo -}}
memory
{{- else if .Values.hydra.config.dsn -}}
{{- .Values.hydra.config.dsn }}
{{- end -}}
{{- end -}}

{{/*
Get the name of the secret resource containing secrets
*/}}
{{- define "hydra.secretname" -}}
{{- if .Values.hydra.existingSecret }}
    {{- printf "%s" .Values.hydra.existingSecret -}}
{{- else -}}
    {{- printf "%s" (include "hydra.fullname" .) -}}
{{- end -}}
{{- end -}}

{{/*
Generate the secrets.system value
*/}}
{{- define "hydra.secrets.system" -}}
{{- if .Values.hydra.config.secrets.system -}}
{{- .Values.hydra.config.secrets.system }}
{{- else if .Values.demo -}}
a-very-insecure-secret-for-checking-out-the-demo
{{- end -}}
{{- end -}}

{{/*
Generate the secrets.cookie value
*/}}
{{- define "hydra.secrets.cookie" -}}
{{- if .Values.hydra.config.secrets.cookie -}}
{{- .Values.hydra.config.secrets.cookie }}
{{- else -}}
{{- include "hydra.secrets.system" . }}
{{- end -}}
{{- end -}}

{{/*
Return true if a secret object should be created
*/}}
{{- define "hydra.createSecret" -}}
{{- if .Values.hydra.existingSecret }}
{{- else -}}
    {{- true -}}
{{- end -}}
{{- end -}}

{{/*
Generate the configmap data, redacting secrets
*/}}
{{- define "hydra.configmap" -}}
{{- $configCopy := .Values.hydra.config | mustDeepCopy -}}
{{- $result := unset $configCopy "dsn" -}}
{{- $result := unset $result "secrets" -}}
{{- $result := unset $result "secretAnnotations" -}}
{{- toYaml $result -}}
{{- end -}}

{{/*
Generate the urls.issuer value
*/}}
{{- define "hydra.config.urls.issuer" -}}
{{- if .Values.hydra.config.urls.self.issuer -}}
{{- .Values.hydra.config.urls.self.issuer }}
{{- else if .Values.ingress.public.enabled -}}
{{- $host := index .Values.ingress.public.hosts 0 -}}
http{{ if $.Values.ingress.public.tls }}s{{ end }}://{{ $host.host }}
{{- else if contains "ClusterIP" .Values.service.public.type -}}
http://127.0.0.1:{{ .Values.service.public.port }}/
{{- end -}}
{{- end -}}

{{/*
Check overrides consistency
*/}}
{{- define "hydra.check.override.consistency" -}}
{{- if and .Values.maester.enabled .Values.fullnameOverride -}}
{{- if not .Values.maester.hydraFullnameOverride -}}
{{ fail "hydra fullname has been overridden, but the new value has not been provided to maester. Set maester.hydraFullnameOverride" }}
{{- else if not (eq .Values.maester.hydraFullnameOverride .Values.fullnameOverride) -}}
{{ fail (tpl "hydra fullname has been overridden, but a different value was provided to maester. {{ .Values.maester.hydraFullnameOverride }} different of {{ .Values.fullnameOverride }}" . ) }}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "hydra.utils.joinListWithComma" -}}
{{- $local := dict "first" true -}}
{{- range $k, $v := . -}}{{- if not $local.first -}},{{- end -}}{{- $v -}}{{- $_ := set $local "first" false -}}{{- end -}}
{{- end -}}
