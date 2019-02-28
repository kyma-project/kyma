#!/usr/bin/env bash

set -o errexit

# # # # # # # # # # # # # # # # #
# VARs coming from environment: #
#                               #
# EXTERNAL_PUBLIC_IP            #
# INGRESSGATEWAY_SERVICE_NAME   #
# PUBLIC_DOMAIN                 #
# TLS_CERT                      #
# # # # # # # # # # # # # # # # #

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source $CURRENT_DIR/utils.sh

generateXipDomain() {

    if [ -z "${EXTERNAL_PUBLIC_IP}" ]; then

        local namespace="istio-system"

        if [ -z "${INGRESSGATEWAY_SERVICE_NAME}" ]; then
            INGRESSGATEWAY_SERVICE_NAME=istio-ingressgateway
        fi

        echo "Trying to get loadbalancer IP address"

        EXTERNAL_PUBLIC_IP=$(getLoadBalancerIP "${INGRESSGATEWAY_SERVICE_NAME}" "${namespace}")

        if [[ "$?" != 0 ]]; then
            echo "External public IP not found"
            exit 1
        fi

        echo "External public IP address is ${EXTERNAL_PUBLIC_IP}"
    fi

    PUBLIC_DOMAIN="${EXTERNAL_PUBLIC_IP}.xip.io"

    DOMAIN_YAML=$(cat << EOF
---
data:
  global.domainName: "${PUBLIC_DOMAIN}"
EOF
)

    kubectl patch configmap installation-config-overrides --patch "${DOMAIN_YAML}" -n kyma-installer

}

generateCerts() {

    XIP_PATCH_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
    KEY_PATH="${XIP_PATCH_DIR}/key.pem"
    CERT_PATH="${XIP_PATCH_DIR}/cert.pem"

    generateCertificatesForDomain "${PUBLIC_DOMAIN}" "${KEY_PATH}" "${CERT_PATH}"

    TLS_CERT=$(base64 "${CERT_PATH}" | tr -d '\n')
    TLS_KEY=$(base64 "${KEY_PATH}" | tr -d '\n')

    rm "${CERT_PATH}"
    rm "${KEY_PATH}"

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

    kubectl patch configmap cluster-certificate-overrides --patch "${TLS_CERT_AND_KEY_YAML}" -n kyma-installer
    kubectl patch secret ingress-tls-cert --patch "${TLS_CERT_YAML}" -n kyma-system

}


if [ -z "${TLS_CERT}" ]; then

    if [ -z "${PUBLIC_DOMAIN}" ] ; then
        generateXipDomain
    fi

    generateCerts
    exit 0
fi

if [ -z "${PUBLIC_DOMAIN}" ]; then
    echo "Invalid setup - no domain for provided certs"
    exit 1
fi
