#!/usr/bin/env bash

set -o errexit

# # # # # # # # # # # # # # # # #
# VARs coming from environment: #
#                               #
# EXTERNAL_PUBLIC_IP            #
# INGRESSGATEWAY_SERVICE_NAME   #
# GLOBAL_DOMAIN                 #
# GLOBAL_TLS_CERT               #
# GLOBAL_TLS_KEY                #
# INGRESS_DOMAIN                #
# INGRESS_TLS_CERT              #
# INGRESS_TLS_KEY               #
# # # # # # # # # # # # # # # # #

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source $CURRENT_DIR/utils.sh

generateXipDomain() {

    if [ -z "${EXTERNAL_PUBLIC_IP}" ]; then

        local namespace="istio-system"

        if [ -z "${INGRESSGATEWAY_SERVICE_NAME}" ]; then
            INGRESSGATEWAY_SERVICE_NAME=istio-ingressgateway
        fi

        EXTERNAL_PUBLIC_IP=$(getLoadBalancerIP "${INGRESSGATEWAY_SERVICE_NAME}" "${namespace}")

        if [[ "$?" != 0 ]]; then
            echo "External public IP not found"
            exit 1
        fi
    fi

    echo "${EXTERNAL_PUBLIC_IP}.xip.io"

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
}

createOverridesConfigMap() {
    if [ -z "$(kubectl get configmap -n kyma-installer net-global-overrides --ignore-not-found)" ]; then
        kubectl create configmap net-global-overrides \
            --form-literal global.ingress.domainName="$PUBLIC_DOMAIN" \
            --from-literal global.ingress.tlsCrt="$TLS_CERT" \
            --from-literal global.ingress.tlsKey="$TLS_KEY" \
            -n kyma-installer
    fi
    kubectl label configmap net-global-overrides --overwrite installer=overrides
}

patchTlsCrtSecret() {
    TLS_CERT_YAML=$(cat << EOF
---
data:
  tls.crt: "${TLS_CERT}"
EOF
    )
    kubectl patch secret ingress-tls-cert --patch "${TLS_CERT_YAML}" -n kyma-system
}

if [ -z "${INGRESS_TLS_CERT}" ]; then
    INGRESS_TLS_CERT="${GLOBAL_TLS_CERT}"
    INGRESS_TLS_KEY="${GLOBAL_TLS_KEY}"

    if [ -z "${INGRESS_TLS_CERT}" ] ; then
        generateCerts
        INGRESS_TLS_CERT=${TLS_CERT}
        INGRESS_TLS_KEY=${TLS_KEY}
    fi
elif [ -z "${INGRESS_DOMAIN}" ]; then
    echo "Certificate provided, but domain is missing!"
    exit 1
fi

if [ -z "${INGRESS_DOMAIN}" ]; then
    INGRESS_DOMAIN="${GLOBAL_DOMAIN}"

    if [ -z "${INGRESS_DOMAIN}" ] ; then
        INGRESS_DOMAIN=$(generateXipDomain)
    fi
fi

createOverridesConfigMap

patchTlsCrtSecret
