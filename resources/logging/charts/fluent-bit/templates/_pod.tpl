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
{{- with .Values.dnsConfig }}
dnsConfig:
  {{- toYaml . | nindent 2 }}
{{- end }}
containers:
  - name: {{ .Chart.Name }}
    securityContext:
      {{- toYaml .Values.securityContext | nindent 6 }}
    image: {{ include "imageurl" (dict "reg" .Values.global.containerRegistry "img" .Values.global.images.fluent_bit) }}
    imagePullPolicy: {{ .Values.image.pullPolicy }}
  {{- if .Values.env }}
    env:
    {{- toYaml .Values.env | nindent 4 }}
  {{- end }}
  {{- if or .Values.envFrom .Values.config.secrets }}
    envFrom:
    {{- if .Values.envFrom }}
    {{- toYaml .Values.envFrom | nindent 4 }}
    {{- end }}
    {{- if .Values.config.secrets }}
    - secretRef:
        name: "{{ template "fluent-bit.fullname" . }}-env-secret"
    {{- end }}
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
    {{- if .Values.livenessProbe }}
    livenessProbe:
      {{- toYaml .Values.livenessProbe | nindent 6 }}
    {{- end }}
    {{- if .Values.readinessProbe }}
    readinessProbe:
      {{- toYaml .Values.readinessProbe | nindent 6 }}
    {{- end }}
    resources:
      {{- toYaml .Values.resources | nindent 6 }}
    volumeMounts:
      - name: config
        mountPath: /fluent-bit/etc/
    {{- if eq .Values.kind "DaemonSet" }}
      - name: varfluentbit
        mountPath: /data
        readOnly: false
      - name: varlogpods
        mountPath: /var/log/pods
        readOnly: true
      - name: varlogcontainers
        mountPath: /var/log/containers
        readOnly: true
      - name: varlibdockercontainers
        mountPath: /var/lib/docker/containers
        readOnly: true
      {{- if .Values.volumes.mountMachineIdFile }}
      - name: etcmachineid
        mountPath: /etc/machine-id
        readOnly: true
      {{- end }}
    {{- end }}
    {{- if and (.Values.config.outputs.es.tls.cert) (.Values.config.outputs.es.tls.key) }}
      - name: es-tls-secret
        mountPath: /secure/es-tls.crt
        subPath: tls.crt
      - name: es-tls-secret
        mountPath: /secure/es-tls.key
        subPath: tls.key
    {{- end }}
    {{- if .Values.config.outputs.es.tls.ca }}
      - name: es-ca-secret
        mountPath: /secure/es-tls-ca.crt
        subPath: es-tls-ca.crt
    {{- end }}
    {{- if and (.Values.config.outputs.forward.tls.cert) (.Values.config.outputs.forward.tls.key) }}
      - name: forward-tls-secret
        mountPath: /secure/forward-tls.crt
        subPath: tls.crt
      - name: forward-tls-secret
        mountPath: /secure/forward-tls.key
        subPath: tls.key
    {{- end }}
    {{- if and (.Values.config.outputs.forward.tls.verify) (.Values.config.outputs.forward.tls.ca) }}
      - name: forward-ca-secret
        mountPath: /secure/forward-tls-ca.crt
        subPath: forward-tls-ca.crt
    {{- end }}
    {{- if and (.Values.config.outputs.http.tls.cert) (.Values.config.outputs.http.tls.key) }}
      - name: http-tls-secret
        mountPath: /secure/http-tls.crt
        subPath: tls.crt
      - name: http-tls-secret
        mountPath: /secure/http-tls.key
        subPath: tls.key
    {{- end }}
    {{- if and (.Values.config.outputs.http.tls.verify) (.Values.config.outputs.http.tls.ca) }}
      - name: http-ca-secret
        mountPath: /secure/http-tls-ca.crt
        subPath: http-tls-ca.crt
    {{- end }}
    {{- with .Values.extraVolumeMounts }}
      {{- tpl . $ | nindent 6}}
    {{- end }}
volumes:
  - name: config
    configMap:
      name: {{ if .Values.existingConfigMap }}{{ .Values.existingConfigMap }}{{- else }}{{ include "fluent-bit.fullname" . }}{{- end }}
{{- if eq .Values.kind "DaemonSet" }}
  - name: varfluentbit
    hostPath:
      path: /var/fluent-bit
  - name: varlogcontainers
    hostPath:
      path: /var/log/containers
  - name: varlogpods
    hostPath:
      path: /var/log/pods
  - name: varlibdockercontainers
    hostPath:
      path: /var/lib/docker/containers
  {{- if .Values.volumes.mountMachineIdFile }}
  - name: etcmachineid
    hostPath:
      path: /etc/machine-id
      type: File
  {{- end }}
{{- end }}
{{- if .Values.config.outputs.es.tls.ca }}
  - name: es-ca-secret
    secret:
      secretName: "{{ template "fluent-bit.fullname" . }}-es-ca-secret"
{{- end }}
{{- if and (.Values.config.outputs.es.tls.crt) (.Values.config.outputs.es.tls.key) }}
  - name: es-tls-secret
    secret:
      secretName: "{{ template "fluent-bit.fullname" . }}-es-tls-secret"
{{- end }}
{{- if and (.Values.config.outputs.forward.tls.verify) (.Values.config.outputs.forward.tls.ca) }}
  - name: forward-ca-secret
    secret:
      secretName: "{{ template "fluent-bit.fullname" . }}-forward-ca-secret"
{{- end }}
{{- if (.Values.config.outputs.forward.tls.enabled) }}
  - name: forward-tls-secret
    secret:
      secretName: "{{ template "fluent-bit.fullname" . }}-forward-tls-secret"
{{- end }}
{{- if and (.Values.config.outputs.http.tls.verify) (.Values.config.outputs.http.tls.ca) }}
  - name: http-ca-secret
    secret:
      secretName: "{{ template "fluent-bit.fullname" . }}-http-ca-secret"
{{- end }}
{{- if .Values.config.outputs.http.tls.enabled }}
  - name: http-tls-secret
    secret:
      secretName: "{{ template "fluent-bit.fullname" . }}-http-tls-secret"
{{- end }}
{{- with .Values.extraVolumes }}
  {{- tpl . $ | nindent 2}}
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
