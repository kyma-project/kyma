#!/usr/bin/env bash

set -o errexit

EXTERNAL_PUBLIC_IP=$(kubectl get service -n istio-system istio-ingressgateway -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

if [ -z ${EXTERNAL_PUBLIC_IP} ]; then
    echo "External public IP not found"
    exit 1
fi

XIP_PATCH_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

DOMAIN="${EXTERNAL_PUBLIC_IP}.xip.io"

CERT_PATH="${XIP_PATCH_DIR}/cert.pem"
KEY_PATH="${XIP_PATCH_DIR}/key.pem"

openssl req -x509 -nodes -days 5 -newkey rsa:4069 \
                 -subj "/CN=${DOMAIN}" \
                 -reqexts SAN -extensions SAN \
                 -config <(cat /etc/ssl/openssl.cnf \
        <(printf "\\n[SAN]\\nsubjectAltName=DNS:*.%s" "${DOMAIN}")) \
                 -keyout "${KEY_PATH}" \
                 -out "${CERT_PATH}"

TLS_CERT=$(base64 "${CERT_PATH}" | tr -d '\n')
TLS_KEY=$(base64 "${KEY_PATH}" | tr -d '\n')

kubectl patch configmap installation-config-overrides -p '{"data": {"global.domainName":"'"${DOMAIN}"'"}}' -n kyma-installer
kubectl patch configmap cluster-certificate-overrides -p '{"data": {"global.tlsCrt":"'"${TLS_CERT}"'"}}' -n kyma-installer
kubectl patch configmap cluster-certificate-overrides -p '{"data": {"global.tlsKey":"'"${TLS_KEY}"'"}}' -n kyma-installer

rm "${CERT_PATH}"
rm "${KEY_PATH}"
