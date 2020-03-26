#!/usr/bin/env bash
#
# This script runs when Helm override global.domainName is set and we're in Gardener environment
#
set -e
# when running on Gardener create Certificate CR
if [ "$DOMAIN" = "" ]; then
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
  DOMAIN="{{ trimPrefix "*." .Values.global.domainName }}" #<----
  kubectl create configmap {{ template "name" . }} --from-literal DOMAIN="$DOMAIN"
fi
if [ "$(cat /etc/apiserver-proxy-tls-cert/tls.key)" = "" ]; then
# when running on Gardener && there is key mouted we skip, do nothing
  echo "Skipping processing existing cert because environment is Gardener which rotates certs by itself"
fi
