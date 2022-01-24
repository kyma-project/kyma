{{- define "fluent-bit.pod" -}}
{{- with .Values.imagePullSecrets }}
imagePullSecrets:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- if or .Values.priorityClassName .Values.global.highPriorityClassName -}}
priorityClassName: {{ coalesce .Values.priorityClassName .Values.global.highPriorityClassName }}
{{- end }}
serviceAccountName: {{ include "fluent-bit.serviceAccountName" . }}
securityContext:
  {{- toYaml .Values.podSecurityContext | nindent 2 }}
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
{{- if .Values.initContainers }}
initContainers:  {{ range $container := .Values.initContainers}}
  - name: {{ $container.name }}
    image: "{{ include "imageurl" (dict "reg" $.Values.global.containerRegistry "img" $.Values.global.images.busybox) }}"
    command:
      {{- toYaml $container.command | nindent 4 }}
  {{- if $container.volumeMounts }}
    volumeMounts:
      {{- toYaml $container.volumeMounts | nindent 4 }}
  {{- end }}
  {{ end }}
{{- end }}
containers:
  - name: {{ .Chart.Name }}
    securityContext:
      {{- toYaml .Values.securityContext | nindent 6 }}
    image: "{{ include "imageurl" (dict "reg" .Values.global.containerRegistry "img" .Values.global.images.fluent_bit) }}"
    imagePullPolicy: {{ .Values.image.pullPolicy }}
  {{- if .Values.env }}
    env:
      {{- toYaml .Values.env | nindent 6 }}
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
        containerPort: 2020
        protocol: TCP
    {{- if .Values.extraPorts }}
      {{- range .Values.extraPorts }}
      - name: {{ .name }}
        containerPort: {{ .containerPort }}
        protocol: {{ .protocol }}
      {{- end }}
    {{- end }}
    livenessProbe:
      {{- toYaml .Values.livenessProbe | nindent 6 }}
    readinessProbe:
      {{- toYaml .Values.readinessProbe | nindent 6 }}
    resources:
      {{- toYaml .Values.resources | nindent 6 }}
    volumeMounts:
      {{- if .Values.dynamicConfigMap }}
      - name: shared-fluent-bit-config
        mountPath: /fluent-bit/etc
      {{- else }}
      - name: config
        mountPath: /fluent-bit/etc
      {{- end }}
      {{- toYaml .Values.volumeMounts | nindent 6 }}
    {{- range $key, $value := .Values.luaScripts }}
      - name: luascripts
        mountPath: /fluent-bit/scripts/{{ $key }}
        subPath: {{ $key }}
    {{- end }}
    {{- if eq .Values.kind "DaemonSet" }}
      {{- toYaml .Values.daemonSetVolumeMounts | nindent 6 }}
      {{- if .Values.volumes.mountMachineIdFile }}
      - name: etcmachineid
        mountPath: /etc/machine-id
        readOnly: true
      {{- end }}
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
  {{- if .Values.dynamicConfigMap }}
  - name: shared-fluent-bit-config
    emptyDir: {}
  - name: dynamic-config
    configMap:
      name: {{ .Values.dynamicConfigMap }}
      optional: true
  {{- end }}
    {{- if .Values.dynamicParsersConfigMap }}
  - name: dynamic-parsers-config
    configMap:
      name: {{ .Values.dynamicParsersConfigMap }}
      optional: true
  {{- end }}
{{- if gt (len .Values.luaScripts) 0 }}
  - name: luascripts
    configMap:
      name: {{ include "fluent-bit.fullname" . }}-luascripts
{{- end }}
{{- if eq .Values.kind "DaemonSet" }}
  {{- toYaml .Values.daemonSetVolumes | nindent 2 }}
  {{- if .Values.volumes.mountMachineIdFile }}
  - name: etcmachineid
    hostPath:
      path: /etc/machine-id
      type: File
  {{- end }}
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
