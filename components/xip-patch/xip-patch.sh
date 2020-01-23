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

    generateCertificatesForDomain "${INGRESS_DOMAIN}" "${KEY_PATH}" "${CERT_PATH}"

    TLS_CERT=$(base64 "${CERT_PATH}" | tr -d '\n')
    TLS_KEY=$(base64 "${KEY_PATH}" | tr -d '\n')

    rm "${CERT_PATH}"
    rm "${KEY_PATH}"
}

createOverridesConfigMap() {
    if [ -z "$(kubectl get configmap -n kyma-installer net-global-overrides --ignore-not-found)" ]; then
        kubectl create configmap net-global-overrides \
            --from-literal global.ingress.domainName="$INGRESS_DOMAIN" \
            --from-literal global.ingress.tlsCrt="$INGRESS_TLS_CERT" \
            --from-literal global.ingress.tlsKey="$INGRESS_TLS_KEY" \
            -n kyma-installer
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
    local delay
    subdomain="kyma"
    delay=10

    echo "Getting Shoot Domain"
    shoot_domain="$(kubectl -n kube-system get configmap shoot-info -o jsonpath='{.data.domain}')"
    DOMAIN="${subdomain}.${shoot_domain}"

    echo "Requesting certificate for domain ${DOMAIN}"
cat <<EOF | kubectl apply -f -
---
apiVersion: cert.gardener.cloud/v1alpha1
kind: Certificate
metadata:
  name: kyma-cert
  namespace: kyma-installer
spec:
  commonName: "*.${DOMAIN}"
EOF

    while :
    do
    local status
    status="$(kubectl get -n kyma-installer certificate.cert.gardener.cloud kyma-cert -o jsonpath='{.status.state}')"
    if [ "${status}" = "Ready" ]; then
        break
    else
        echo "Waiting for Certicate generation, status is ${status}"
        sleep ${delay}
    fi
    done

    local secret_name
    secret_name=$(kubectl get -n kyma-installer certificate kyma-cert -o jsonpath="{.spec.secretRef.name}")
    echo "Getting certificate from secret"
    TLS_CERT=$(kubectl get -n kyma-installer secret  "${secret_name}" -o jsonpath="{.data['tls\.crt']}" | sed 's/ /\\ /g' | tr -d '\n')
    TLS_KEY=$(kubectl get -n kyma-installer secret  "${secret_name}" -o jsonpath="{.data['tls\.key']}" | sed 's/ /\\ /g' | tr -d '\n')

    echo "Annotating Istio Ingress Gateway with Gardener DNS"
    kubectl -n istio-system annotate service istio-ingressgateway dns.gardener.cloud/class='garden' dns.gardener.cloud/dnsnames='*.'"${DOMAIN}"'' --overwrite
}

echo "Checking if running on Gardener"
if [ -n "$(kubectl -n kube-system get configmap shoot-info --ignore-not-found)" ]; then
  requestGardenerCerts
  INGRESS_DOMAIN=${DOMAIN}
  INGRESS_TLS_CERT=${TLS_CERT}
  INGRESS_TLS_KEY=${TLS_KEY}
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
