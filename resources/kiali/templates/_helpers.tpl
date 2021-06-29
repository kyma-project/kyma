{{/* vim: set filetype=mustache: */}}

{{/*
Expand the name of the chart.
*/}}
{{- define "kiali-server.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kiali-server.fullname" -}}
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
{{- define "kiali-server.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Identifies the log_level with the old verbose_mode and the new log_level considered.
*/}}
{{- define "kiali-server.logLevel" -}}
{{- if .Values.kiali.spec.deployment.verbose_mode -}}
{{- .Values.kiali.spec.deployment.verbose_mode -}}
{{- else -}}
{{- .Values.kiali.spec.deployment.logger.log_level -}}
{{- end -}}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kiali-server.labels" -}}
helm.sh/chart: {{ include "kiali-server.chart" . }}
app: {{ include "kiali-server.name" . }}
{{ include "kiali-server.selectorLabels" . }}
version: {{ .Values.kiali.spec.deployment.version_label | default .Chart.AppVersion | quote }}
app.kubernetes.io/version: {{ .Values.kiali.spec.deployment.version_label | default .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: "kiali"
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kiali-server.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kiali-server.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Used to determine if a custom dashboard (defined in .Template.Name) should be deployed.
*/}}
{{- define "kiali-server.isDashboardEnabled" -}}
{{- if .Values.kiali.spec.external_services.custom_dashboards.enabled }}
  {{- $includere := "" }}
  {{- range $_, $s := .Values.kiali.spec.deployment.custom_dashboards.includes }}
    {{- if $s }}
      {{- if $includere }}
        {{- $includere = printf "%s|^%s$" $includere ($s | replace "*" ".*" | replace "?" ".") }}
      {{- else }}
        {{- $includere = printf "^%s$" ($s | replace "*" ".*" | replace "?" ".") }}
      {{- end }}
    {{- end }}
  {{- end }}
  {{- $excludere := "" }}
  {{- range $_, $s := .Values.kiali.spec.deployment.custom_dashboards.excludes }}
    {{- if $s }}
      {{- if $excludere }}
        {{- $excludere = printf "%s|^%s$" $excludere ($s | replace "*" ".*" | replace "?" ".") }}
      {{- else }}
        {{- $excludere = printf "^%s$" ($s | replace "*" ".*" | replace "?" ".") }}
      {{- end }}
    {{- end }}
  {{- end }}
  {{- if (and (mustRegexMatch (default "no-matches" $includere) (base .Template.Name)) (not (mustRegexMatch (default "no-matches" $excludere) (base .Template.Name)))) }}
    {{- print "enabled" }}
  {{- else }}
    {{- print "" }}
  {{- end }}
{{- else }}
  {{- print "" }}
{{- end }}
{{- end }}

{{/*
Determine the default login token signing key.
*/}}
{{- define "kiali-server.login_token.signing_key" -}}
{{- if .Values.kiali.spec.login_token.signing_key }}
  {{- .Values.kiali.spec.login_token.signing_key }}
{{- else }}
  {{- randAlphaNum 16 }}
{{- end }}
{{- end }}

{{/*
Determine the default web root.
*/}}
{{- define "kiali-server.server.web_root" -}}
{{- if .Values.kiali.spec.server.web_root  }}
  {{- .Values.kiali.spec.server.web_root | trimSuffix "/" }}
{{- else }}
  {{- if .Capabilities.APIVersions.Has "route.openshift.io/v1" }}
    {{- "/" }}
  {{- else }}
    {{- "/kiali" }}
  {{- end }}
{{- end }}
{{- end }}

{{/*
Determine the default identity cert file. There is no default if on k8s; only on OpenShift.
*/}}
{{- define "kiali-server.identity.cert_file" -}}
{{- if hasKey .Values.kiali.spec.identity "cert_file" }}
  {{- .Values.kiali.spec.identity.cert_file }}
{{- else }}
  {{- if .Capabilities.APIVersions.Has "route.openshift.io/v1" }}
    {{- "/kiali-cert/tls.crt" }}
  {{- else }}
    {{- "" }}
  {{- end }}
{{- end }}
{{- end }}

{{/*
Determine the default identity private key file. There is no default if on k8s; only on OpenShift.
*/}}
{{- define "kiali-server.identity.private_key_file" -}}
{{- if hasKey .Values.kiali.spec.identity "private_key_file" }}
  {{- .Values.kiali.spec.identity.private_key_file }}
{{- else }}
  {{- if .Capabilities.APIVersions.Has "route.openshift.io/v1" }}
    {{- "/kiali-cert/tls.key" }}
  {{- else }}
    {{- "" }}
  {{- end }}
{{- end }}
{{- end }}

{{/*
Determine the istio namespace - default is where Kiali is installed.
*/}}
{{- define "kiali-server.istio_namespace" -}}
{{- if .Values.kiali.spec.istio_namespace }}
  {{- .Values.kiali.spec.istio_namespace }}
{{- else }}
  {{- .Release.Namespace }}
{{- end }}
{{- end }}

{{/*
Determine the auth strategy to use - default is "token" on Kubernetes and "openshift" on OpenShift.
*/}}
{{- define "kiali-server.auth.strategy" -}}
{{- if .Values.kiali.spec.auth.strategy }}
  {{- if (and (eq .Values.kiali.spec.auth.strategy "openshift") (not .Values.kiali.spec.kiali_route_url)) }}
    {{- fail "You did not define what the Kiali Route URL will be (--set kiali_route_url=...). Without this set, the openshift auth strategy will not work. Either set that or use a different auth strategy via the --set auth.strategy=... option." }}
  {{- end }}
  {{- .Values.kiali.spec.auth.strategy }}
{{- else }}
  {{- if .Capabilities.APIVersions.Has "route.openshift.io/v1" }}
    {{- if not .Values.kiali.spec.kiali_route_url }}
      {{- fail "You did not define what the Kiali Route URL will be (--set kiali_route_url=...). Without this set, the openshift auth strategy will not work. Either set that or explicitly indicate another auth strategy you want via the --set auth.strategy=... option." }}
    {{- end }}
    {{- "openshift" }}
  {{- else }}
    {{- "token" }}
  {{- end }}
{{- end }}
{{- end }}
