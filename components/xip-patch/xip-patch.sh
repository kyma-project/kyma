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
# TLS_SECRET_NAME               #
# # # # # # # # # # # # # # # # #

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source $CURRENT_DIR/utils.sh

TLS_SECRET_NAME="${TLS_SECRET_NAME:-kyma-gateway-certs}"
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

    generateCertificatesForDomain "${INGRESS_DOMAIN}" "${KEY_PATH}" "${CERT_PATH}"

    TLS_CERT=$(base64 "${CERT_PATH}" | tr -d '\n')
    TLS_KEY=$(base64 "${KEY_PATH}" | tr -d '\n')

    rm "${CERT_PATH}"
    rm "${KEY_PATH}"
}

createOverridesConfigMap() {
COMMON_PARAMS=$(echo --from-literal global.ingress.domainName="$INGRESS_DOMAIN" \
                     --from-literal global.ingress.tlsCrt="$INGRESS_TLS_CERT" \
                     --from-literal global.ingress.tlsKey="$INGRESS_TLS_KEY" \
                     --from-literal global.environment.gardener="$GARDENER_ENVIRONMENT") \

    if [[ "$GARDENER_ENVIRONMENT" == true ]]; then
        kubectl create configmap net-global-overrides ${COMMON_PARAMS} \
        --from-literal global.domainName="$INGRESS_DOMAIN" \
        --from-literal global.tlsCrt="$INGRESS_TLS_CERT" \
        --from-literal global.tlsKey="$INGRESS_TLS_KEY" \
        -n kyma-installer -o yaml --dry-run | kubectl apply -f -
    else
        kubectl create configmap net-global-overrides ${COMMON_PARAMS} \
        -n kyma-installer -o yaml --dry-run | kubectl apply -f -
    fi
    kubectl label configmap net-global-overrides --overwrite installer=overrides -n kyma-installer
    kubectl label configmap net-global-overrides --overwrite kyma-project.io/installation="" -n kyma-installer
}

patchTlsCrtSecret() {
    TLS_CERT_YAML=$(cat << EOF
---
data:
  tls.crt: "${INGRESS_TLS_CERT}"
EOF
    )
    set +e
    local msg
    local status
    msg=$(kubectl patch secret ingress-tls-cert --patch "${TLS_CERT_YAML}" -n kyma-system 2>&1)
    status=$?
    set -e
    if [[ $status -ne 0 ]] && [[ ! "$msg" == *"not patched"* ]]; then
        echo "$msg"
        exit $status
    fi
}

requestGardenerCerts() {
    local subdomain
    local shoot_domain
    subdomain="kyma"

    echo "Getting Shoot Domain"
    shoot_domain="$(kubectl -n kube-system get configmap shoot-info -o jsonpath='{.data.domain}')"
    DOMAIN="${subdomain}.${shoot_domain}"

    echo "Requesting certificate for domain ${DOMAIN}"
cat <<EOF | kubectl apply -f -
---
apiVersion: cert.gardener.cloud/v1alpha1
kind: Certificate
metadata:
  name: kyma-tls-cert
  namespace: istio-system
spec:
  commonName: "*.${DOMAIN}"
  secretName: "$TLS_SECRET_NAME"
EOF

    SECONDS=0
    END_TIME=$((SECONDS+600)) #600 seconds = 10 minutes
    
    while [ ${SECONDS} -lt ${END_TIME} ];do
        STATUS="$(kubectl get -n istio-system certificate.cert.gardener.cloud kyma-tls-cert -o jsonpath='{.status.state}')"
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

    echo "Getting certificate from secret"
    TLS_CERT=$(kubectl get -n istio-system secret  "${TLS_SECRET_NAME}" -o jsonpath="{.data['tls\.crt']}" | sed 's/ /\\ /g' | tr -d '\n')
    TLS_KEY=$(kubectl get -n istio-system secret  "${TLS_SECRET_NAME}" -o jsonpath="{.data['tls\.key']}" | sed 's/ /\\ /g' | tr -d '\n')

    echo "Annotating Istio Ingress Gateway with Gardener DNS"
    kubectl -n istio-system annotate service istio-ingressgateway dns.gardener.cloud/class='garden' dns.gardener.cloud/dnsnames='*.'"${DOMAIN}"'' --overwrite
}

echo "Checking if running on Gardener"

GARDENER_ENVIRONMENT=false

if [ -n "$(kubectl -n kube-system get configmap shoot-info --ignore-not-found)" ] &&
   [ -z "$(kubectl get configmap -n kyma-installer net-global-overrides --ignore-not-found)" ]; then
  requestGardenerCerts
  INGRESS_DOMAIN=${DOMAIN}
  INGRESS_TLS_CERT=${TLS_CERT}
  INGRESS_TLS_KEY=${TLS_KEY}
  GARDENER_ENVIRONMENT=true
fi

INGRESS_TLS_CERT="${INGRESS_TLS_CERT:-$GLOBAL_TLS_CERT}"
INGRESS_TLS_KEY="${INGRESS_TLS_KEY:-$GLOBAL_TLS_KEY}"
INGRESS_DOMAIN="${INGRESS_DOMAIN:-$GLOBAL_DOMAIN}"

if [ -n "${INGRESS_TLS_CERT}" ] && [ -z "${INGRESS_DOMAIN}" ]; then
    echo "Certificate provided, but domain is missing!"
    exit 1
fi

if [ -z "${INGRESS_DOMAIN}" ] ; then
    INGRESS_DOMAIN=$(generateXipDomain)
fi

if [ -z "${INGRESS_TLS_CERT}" ] ; then
    generateCerts
    INGRESS_TLS_CERT=${TLS_CERT}
    INGRESS_TLS_KEY=${TLS_KEY}
fi

createOverridesConfigMap

patchTlsCrtSecret
