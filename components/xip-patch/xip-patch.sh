#!/usr/bin/env bash

set -o errexit

if [ -z "${INGRESSGATEWAY_SERVICE_NAME}" ]; then
  INGRESSGATEWAY_SERVICE_NAME=istio-ingressgateway
fi

if [ -z "${EXTERNAL_PUBLIC_IP}" ]; then

  SECONDS=0
  END_TIME=$((SECONDS+60))

  while [ ${SECONDS} -lt ${END_TIME} ];do
      echo "Trying to get loadbalancer IP address"

      EXTERNAL_PUBLIC_IP=$(kubectl get service -n istio-system ${INGRESSGATEWAY_SERVICE_NAME} -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

      if [ "${EXTERNAL_PUBLIC_IP}" ]; then
          echo "External public IP address is ${EXTERNAL_PUBLIC_IP}"
          break
      fi
      
      sleep 10
  done
fi

if [ -z "${EXTERNAL_PUBLIC_IP}" ]; then
    echo "External public IP not found"
    exit 1
fi

XIP_PATCH_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

PUBLIC_DOMAIN="${EXTERNAL_PUBLIC_IP}.xip.io"

CERT_PATH="${XIP_PATCH_DIR}/cert.pem"
KEY_PATH="${XIP_PATCH_DIR}/key.pem"

openssl req -x509 -nodes -days 5 -newkey rsa:4069 \
                 -subj "/CN=${PUBLIC_DOMAIN}" \
                 -reqexts SAN -extensions SAN \
                 -config <(cat /etc/ssl/openssl.cnf \
        <(printf "\\n[SAN]\\nsubjectAltName=DNS:*.%s" "${PUBLIC_DOMAIN}")) \
                 -keyout "${KEY_PATH}" \
                 -out "${CERT_PATH}"

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
