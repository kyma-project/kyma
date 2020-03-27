#!/usr/bin/env bash
#
#This script runs when Helm override global.domainName is set and we're NOT in Gardener environment
#
set -e
if [ "$DOMAIN" = "" ]; then
  DOMAIN={{ .Values.global.domainName }}
  kubectl create configmap {{ template "name" . }} --from-literal DOMAIN="$DOMAIN"
fi
if [ "$(cat /etc/apiserver-proxy-tls-cert/tls.key)" = "" ]; then
# if running on given by user domain create secret with key and cert
# else generate domain and create secret
{{ if .Values.global.tlsKey }}
  echo "{{ .Values.global.tlsKey }}" | base64 -d > ${HOME}/key.pem
  echo "{{ .Values.global.tlsCrt }}" | base64 -d > ${HOME}/cert.pem
  kubectl create secret tls {{ template "name" . }}-tls-cert  --key ${HOME}/key.pem --cert ${HOME}/cert.pem
{{ else }}
  source /app/utils.sh
  generateCertificatesForDomain "$DOMAIN" ${HOME}/key.pem ${HOME}/cert.pem
  kubectl create secret tls {{ template "name" . }}-tls-cert  --key ${HOME}/key.pem --cert ${HOME}/cert.pem
{{ end }}
fi
