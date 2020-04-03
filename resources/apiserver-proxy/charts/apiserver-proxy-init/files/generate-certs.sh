#!/usr/bin/env bash
set -e
# if running on Gardener create Certificate CR
# else proceed 'old' way
{{- if .Values.global.domainName }}
{{- if .Values.global.environment.gardener }}
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
{{- else }}
  DOMAIN={{ .Values.global.domainName }}
{{- end }}
{{- else }}
  source /app/utils.sh
  INGRESS_IP=$(getLoadBalancerIP {{ template "name" . }}-ssl {{ .Release.Namespace }})
  DOMAIN="$INGRESS_IP.xip.io"
{{- end }}
  kubectl create configmap {{ template "name" . }} --from-literal DOMAIN="$DOMAIN"  -o yaml --dry-run | kubectl apply -f -
if [ "$(cat /etc/apiserver-proxy-tls-cert/tls.key)" = "" ]; then
# if running on Gardener do nothing
# if user provided key and cert create secret
# else generate key&cert and create secret
{{- if .Values.global.environment.gardener }}
  echo "Running on Gardener. Skipping processing secret with cert and key because Gardener will provide"
{{- else if .Values.global.tlsKey }}
  echo "Running on envrionment with user provided cert and key, creating secret with it"
  echo "{{ .Values.global.tlsKey }}" | base64 -d > ${HOME}/key.pem
  echo "{{ .Values.global.tlsCrt }}" | base64 -d > ${HOME}/cert.pem
  kubectl create secret tls {{ template "name" . }}-tls-cert  --key ${HOME}/key.pem --cert ${HOME}/cert.pem
{{- else }}
  echo "Running on xip.io enabled cluster, creating certificate for the domain"
  source /app/utils.sh
  generateCertificatesForDomain "$DOMAIN" ${HOME}/key.pem ${HOME}/cert.pem
  kubectl create secret tls {{ template "name" . }}-tls-cert  --key ${HOME}/key.pem --cert ${HOME}/cert.pem -o yaml --dry-run | kubectl apply -f -
{{- end }}
else
# xip.io case, geenerate cert for domain
# This is always done! After 1.11 we need to revert it back to old logic as in 1.11 apiserver-proxy-ssl loadbalancer
# will get new external IP and we do not have information about old IP
# Previously in order to rotate xip cert you need to manually delete secret and then trigger update
# Once 1.11 will be released we can remove it. Currently upgrade on xip cluster will not generate new cert for new IP
# leaving kubeconfig broken. The best approach would be to decode exisiting cert and compare IP from domain with IP of current LB.
{{- if not (and (.Values.global.environment.gardener) (.Values.global.tlsKey)) }}
  echo "Running on xip.io enabled cluster, creating certificate for the domain"
  source /app/utils.sh
  generateCertificatesForDomain "$DOMAIN" ${HOME}/key.pem ${HOME}/cert.pem
  kubectl create secret tls {{ template "name" . }}-tls-cert  --key ${HOME}/key.pem --cert ${HOME}/cert.pem -o yaml --dry-run | kubectl apply -f -
{{- end }}
  echo "Done"
fi