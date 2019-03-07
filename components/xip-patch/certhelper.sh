#!/usr/bin/env bash
# EXPECTED ENVS
# - DOMAIN (optional) - Static domain for which to generate certs
# - TLS_CRT (optinal) - Current TLS certificate
# - TLS_KEY (optional) - Current TLS cert key
# - LB_LABEL (required) - Selector label for the LoadBalancer service
# - LB_NAMESPACE (required) - Namespace for the LoadBalancer service

set -e

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source $CURRENT_DIR/utils.sh

discoverUnsetVar=false

for var in LB_LABEL LB_NAMESPACE; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done

if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

function generateXipDomain() {
  INGRESS_IP=$(getLoadBalancerIPFromLabel "${LB_LABEL}" "${LB_NAMESPACE}")
  DOMAIN="${INGRESS_IP}.xip.io"
  DOMAIN_YAML=$(cat << EOF
---
data:
  global.applicationConnectorDomainName: "${DOMAIN}"
EOF
)
  echo "---> DOMAIN created: ${DOMAIN}, patching configmap"
  kubectl patch configmap installation-config-overrides --patch "${DOMAIN_YAML}" -n kyma-installer
}

function generateCerts() {
  echo "---> Generating Certs for ${DOMAIN}"
  generateCertificatesForDomain "${DOMAIN}" /root/key.pem /root/cert.pem
  TLS_CERT=$(base64 /root/cert.pem | tr -d '\n')
  TLS_KEY=$(base64 /root/key.pem | tr -d '\n')

  TLS_CERT_AND_KEY_YAML=$(cat << EOF
---
data:
  global.applicationConnector.tlsCrt: "${TLS_CERT}"
  global.applicationConnector.tlsKey: "${TLS_KEY}"
EOF
)
  echo "---> Certs have been created, patching configmap"
  kubectl patch configmap cluster-certificate-overrides --patch "${TLS_CERT_AND_KEY_YAML}" -n kyma-installer 
}

if [ -z "${TLS_CRT}" ]] && [[ -z "${TLS_KEY}" ]]; then
    if [[ -z "${DOMAIN}" ]]; then
        generateXipDomain
    fi
    generateCerts
    exit 0
fi
if [ -z "${DOMAIN}" ]; then
    echo "Invalid setup - no domain for provided certs"
    exit 1
fi
