#!/usr/bin/env bash
set -e
# if running on Gardener create Certificate CR
# else proceed 'old' way
if [ "$DOMAIN" = "" ]; then
{{ if .Values.global.domainName }}
{{ if .Values.global.environment.gardener }}
cat <<EOF | kubectl apply -f -
apiVersion: cert.gardener.cloud/v1alpha1
kind: Certificate
metadata:
  name: apiserver-proxy-tls-cert
  namespace: kyma-system
spec:
  commonName: "apiserver.{{ trimPrefix "*." .Values.global.domainName }}"
  secretName: "{{ template "name" . }}-tls-cert"
EOF

  SECONDS=0
  END_TIME=$((SECONDS+600)) #600 seconds = 10 minutes
  while [ ${SECONDS} -lt ${END_TIME} ];do
    STATUS="$(kubectl get -n {{ .Release.Namespace }} certificate.cert.gardener.cloud {{ template "name" . }}-tls-cert -o jsonpath='{.status.state}')"
    if [ "${STATUS}" = "Ready" ]; then
      break
    fi
    echo "Waiting for Certicate generation, status is ${STATUS}"
    sleep 10
  done
  if [ "${STATUS}" != "Ready" ]; then
    echo "Certificate is still not ready, status is ${STATUS}. Exiting.."
    exit 1
  fi
  DOMAIN="{{ trimPrefix "*." .Values.global.domainName }}"
{{ else }}
  DOMAIN={{ .Values.global.domainName }}
{{ end }}
{{ else }}
  source /app/utils.sh
  INGRESS_IP=$(getLoadBalancerIP {{ template "name" . }}-ssl {{ .Release.Namespace }})
  DOMAIN="$INGRESS_IP.xip.io"
{{ end }}
  kubectl create configmap {{ template "name" . }} --from-literal DOMAIN="$DOMAIN"
fi
if [ "$(cat /etc/apiserver-proxy-tls-cert/tls.key)" = "" ]; then
# if running on Gardener && there is key mouted we skip, do nothing
# if running on given by user domain create secret with key and cert
# else generate domain and create secret
{{ if .Values.global.environment.gardener }}
  echo "Skipping processing existing cert because environment is Gardener which rotates certs by itself"
{{ else if .Values.global.tlsKey }}
  echo "{{ .Values.global.tlsKey }}" | base64 -d > ${HOME}/key.pem
  echo "{{ .Values.global.tlsCrt }}" | base64 -d > ${HOME}/cert.pem
  kubectl create secret tls {{ template "name" . }}-tls-cert  --key ${HOME}/key.pem --cert ${HOME}/cert.pem
{{ else }}
  source /app/utils.sh
  generateCertificatesForDomain "$DOMAIN" ${HOME}/key.pem ${HOME}/cert.pem
  kubectl create secret tls {{ template "name" . }}-tls-cert  --key ${HOME}/key.pem --cert ${HOME}/cert.pem
{{ end }}
fi