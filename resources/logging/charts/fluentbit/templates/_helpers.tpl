{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "fluent-bit.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "fluent-bit.fullname" -}}
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
Create images repositories
*/}}
{{- define "fluent-bit.imageRepository" -}}
{{- $keyName := .keyName -}}
{{- $imageName := base (index .Values.image $keyName).repository -}}
{{- $imageTag := (index .Values.image $keyName).tag -}}
{{- if .Values.global.imageRepository -}}
{{- printf "%s:%s" (printf "%s/%s" .Values.global.imageRepository $imageName) $imageTag -}}
{{- else -}}
{{- printf "%s:%s" (index .Values.image $keyName).repository $imageTag -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "fluent-bit.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "fluent-bit.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "fluent-bit.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Create meta labels
*/}}
{{- define "fluent-bit.metaLabels" -}}
{{- $labelChart := include "fluent-bit.chart" $ -}}
{{- $labelApp := include "fluent-bit.name" $ -}}
{{- if .Values.helmLabels.enabled -}}
  {{- $labels := dict "chart" $labelChart "release" .Release.Name "heritage" .Release.Service -}}
  {{ merge $labels .Values.extraLabels (default (dict "app" $labelApp) .Values.defaultLabels) | toYaml }}
{{- else -}}
  {{ merge .Values.extraLabels (default (dict "app" $labelApp) .Values.defaultLabels) | toYaml }}
{{- end -}}
{{- end -}}

{{/*
Create match labels
*/}}
{{- define "fluent-bit.matchLabels" -}}
{{- $labelApp := include "fluent-bit.name" $ -}}
{{- if .Values.helmLabels.enabled }}
  {{- $labels := dict "release" .Release.Name -}}
  {{ merge $labels .Values.extraLabels (default (dict "app" $labelApp) .Values.defaultLabels) | toYaml }}
{{- else -}}
  {{ merge .Values.extraLabels (default (dict "app" $labelApp) .Values.defaultLabels) | toYaml }}
{{- end -}}
{{- end -}}

{{/*
Return the appropriate apiVersion for deployment. (Can be tested with --kube-version)
*/}}
{{- define "fluent-bit.ds.apiVersion" -}}
{{- if (semverCompare ">=1.6-0, <1.8-0" .Capabilities.KubeVersion.GitVersion) -}}
{{- print "apps/v1beta1" -}}
{{- else if (semverCompare ">=1.8-0, <1.9-0" .Capabilities.KubeVersion.GitVersion) -}}
{{- print "apps/v1beta2" -}}
{{- else -}}
{{- print "apps/v1" -}}
{{- end -}}
{{- end -}}

{{/*
Return the arguments of the metrics-collection script
*/}}
{{- define "fluent-bit.metricsArguments" -}}
  {{ printf "--endpoint \"%s\"" .Values.prometheusPushGateway.endpoint -}}
  {{- if .Values.prometheusPushGateway.metricsEndpoint -}}
    {{ printf " --metrics-endpoint \"%s\"" .Values.prometheusPushGateway.metricsEndpoint -}}
  {{- end -}}
  {{- if .Values.prometheusPushGateway.tls.enabled -}}
    {{ printf " --enable-pg-tls" -}}
  {{- end -}}
  {{- if .Values.prometheusPushGateway.tls.auth -}}
    {{ printf " --enable-pg-auth" -}}
  {{- end -}}
  {{- if not .Values.prometheusPushGateway.tls.caCertificate -}}
    {{ printf " --insecure" -}}
  {{- end -}}
{{- end -}}

{{- define "helm-toolkit.utils.joinListWithComma" -}}
{{- $local := dict "first" true -}}
{{- range $k, $v := . -}}{{- if not $local.first -}},{{- end -}}{{- $v -}}{{- $_ := set $local "first" false -}}{{- end -}}
{{- end -}}

{{- define "loki.namespace.filter" -}}
{{- if .Values.conf.Input.Kubernetes_loki.exclude.namespaces -}}
{{- $namespaces := splitList "," .Values.conf.Input.Kubernetes_loki.exclude.namespaces -}}
{{- range $namespaces -}}
{{- printf "/var/log/containers/*_%s_*.log, " . -}}
{{- end -}}
{{- end -}}
{{- end -}}