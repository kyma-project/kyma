#!/usr/bin/env bash

set -o errexit

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "Trying to get loadbalancer IP address"
if [[ -z "${INGRESSGATEWAY_SERVICE_NAME}" ]]; then
  INGRESSGATEWAY_SERVICE_NAME=istio-ingressgateway
fi

if [[ -z "${EXTERNAL_PUBLIC_IP}" ]]; then
  EXTERNAL_PUBLIC_IP=$(SERVICE_NAME="$INGRESSGATEWAY_SERVICE_NAME" SERVICE_NAMESPACE="istio-system" get-service-ip.sh)
fi
echo "External public IP address is ${EXTERNAL_PUBLIC_IP}"

DOMAIN="${EXTERNAL_PUBLIC_IP}.xip.io"

OUT_DIR=${DIR} ${DIR}/generate-cert.sh

TLS_CERT=$(base64 "${CERT_PATH}" | tr -d '\n')
TLS_KEY=$(base64 "${KEY_PATH}" | tr -d '\n')

rm "${CERT_PATH}"
rm "${KEY_PATH}"

PUBLIC_DOMAIN_YAML=$(cat << EOF
---
data:
  global.domainName: "${PUBLIC_DOMAIN}"
EOF
)

TLS_CERT_YAML=$(cat << EOF
---
data:
  tls.crt: "${TLS_CERT}"
EOF
)

TLS_CERT_AND_KEY_YAML=$(cat << EOF
---
data:
  global.tlsCrt: "${TLS_CERT}"
  global.tlsKey: "${TLS_KEY}"
EOF
)

kubectl patch configmap installation-config-overrides --patch "${PUBLIC_DOMAIN_YAML}" -n kyma-installer
kubectl patch configmap cluster-certificate-overrides --patch "${TLS_CERT_AND_KEY_YAML}" -n kyma-installer
kubectl patch secret ingress-tls-cert --patch "${TLS_CERT_YAML}" -n kyma-system
