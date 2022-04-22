
{{- define "fluent-bit.pod" -}}
serviceAccountName: {{ include "fluent-bit.serviceAccountName" . }}
{{- with .Values.imagePullSecrets }}
imagePullSecrets:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- if or .Values.priorityClassName .Values.global.highPriorityClassName }}
priorityClassName: {{ coalesce .Values.priorityClassName .Values.global.highPriorityClassName }}
{{- end }}
{{- with .Values.podSecurityContext }}
securityContext:
  {{- toYaml . | nindent 2 }}
{{- end }}
hostNetwork: {{ .Values.hostNetwork }}
dnsPolicy: {{ .Values.dnsPolicy }}
{{- with .Values.dnsConfig }}
dnsConfig:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- with .Values.hostAliases }}
hostAliases:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- with .Values.initContainers }}
initContainers:
{{- if kindIs "string" . }}
  {{- tpl . $ | nindent 2 }}
{{- else }}
  {{-  toYaml . | nindent 2 }}
{{- end -}}
{{- end }}
containers:
  - name: {{ .Chart.Name }}
  {{- with .Values.securityContext }}
    securityContext:
      {{- toYaml . | nindent 6 }}
  {{- end }}
    image: "{{ include "imageurl" (dict "reg" .Values.global.containerRegistry "img" .Values.global.images.fluent_bit) }}"
    imagePullPolicy: {{ .Values.image.pullPolicy }}
  {{- if or .Values.env .Values.envWithTpl }}
    env:
    {{- with .Values.env }}
      {{- toYaml . | nindent 6 }}
    {{- end }}
    {{- range $item := .Values.envWithTpl }}
      - name: {{ $item.name }}
        value: {{ tpl $item.value $ | quote }}
    {{- end }}
  {{- end }}
  {{- if .Values.envFrom }}
    envFrom:
      {{- toYaml .Values.envFrom | nindent 6 }}
  {{- end }}
  {{- if .Values.args }}
    args:
    {{- toYaml .Values.args | nindent 6 }}
  {{- end}}
  {{- if .Values.command }}
    command:
    {{- toYaml .Values.command | nindent 6 }}
  {{- end }}
    ports:
      - name: http
        containerPort: {{ .Values.metricsPort }}
        protocol: TCP
    {{- if .Values.extraPorts }}
      {{- range .Values.extraPorts }}
      - name: {{ .name }}
        containerPort: {{ .containerPort }}
        protocol: {{ .protocol }}
      {{- end }}
    {{- end }}
  {{- with .Values.lifecycle }}
    lifecycle:
      {{- toYaml . | nindent 6 }}
  {{- end }}
    livenessProbe:
      {{- toYaml .Values.livenessProbe | nindent 6 }}
    readinessProbe:
      {{- toYaml .Values.readinessProbe | nindent 6 }}
  {{- with .Values.resources }}
    resources:
      {{- toYaml . | nindent 6 }}
  {{- end }}
    volumeMounts:
      {{- toYaml .Values.volumeMounts | nindent 6 }}
    {{- range $key, $val := .Values.config.extraFiles }}
      - name: config
        mountPath: /fluent-bit/etc/{{ $key }}
        subPath: {{ $key }}
    {{- end }}
    {{- range $key, $value := .Values.luaScripts }}
      - name: luascripts
        mountPath: /fluent-bit/scripts/{{ $key }}
        subPath: {{ $key }}
    {{- end }}
    {{- if eq .Values.kind "DaemonSet" }}
      {{- toYaml .Values.daemonSetVolumeMounts | nindent 6 }}
    {{- end }}
    {{- if .Values.extraVolumeMounts }}
      {{- toYaml .Values.extraVolumeMounts | nindent 6 }}
    {{- end }}
  {{- if .Values.extraContainers }}
    {{- toYaml .Values.extraContainers | nindent 2 }}
  {{- end }}
volumes:
  - name: config
    configMap:
      name: {{ if .Values.existingConfigMap }}{{ .Values.existingConfigMap }}{{- else }}{{ include "fluent-bit.fullname" . }}{{- end }}
{{- if gt (len .Values.luaScripts) 0 }}
  - name: luascripts
    configMap:
      name: {{ include "fluent-bit.fullname" . }}-luascripts
{{- end }}
{{- if eq .Values.kind "DaemonSet" }}
  {{- toYaml .Values.daemonSetVolumes | nindent 2 }}
{{- end }}
{{- if .Values.extraVolumes }}
  {{- toYaml .Values.extraVolumes | nindent 2 }}
{{- end }}
{{- with .Values.nodeSelector }}
nodeSelector:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- with .Values.affinity }}
affinity:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- with .Values.tolerations }}
tolerations:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- end -}}
