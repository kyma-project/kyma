#!/usr/bin/env bash
# EXPECTED ENVS
# - DOMAIN (optional) - Static domain for which to generate certs
# - TLS_CRT (optional) - Current TLS certificate
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
  global.applicationConnector.domainName: "${DOMAIN}"
EOF
)
  echo "---> DOMAIN created: ${DOMAIN}, patching configmap"
  set +e
  local msg
  local status
  msg=$(kubectl patch configmap application-connector-overrides --patch "${DOMAIN_YAML}" -n kyma-installer 2>&1)
  status=$?
  set -e
  if [[ $status -ne 0 ]] && [[ ! "$msg" == *"not patched"* ]]; then
    echo "$msg"
    exit $status
  fi
}

function generateCerts() {
  echo "---> Generating Certs for ${DOMAIN}"
  generateCertificatesForDomain "${DOMAIN}" ${HOME}/key.pem ${HOME}/cert.pem
  CERT=$(base64 ${HOME}/cert.pem | tr -d '\n')
  KEY=$(base64 ${HOME}/key.pem | tr -d '\n')
  rm ${HOME}/key.pem ${HOME}/cert.pem
}

function createOverrideCM() {
  if [ -z "$(kubectl get configmap -n kyma-installer application-connector-overrides --ignore-not-found)" ]; then
    echo "---> ConfigMap application-connector-overrides not found! Creating."
    kubectl create cm application-connector-overrides -n kyma-installer --from-literal="foo=bar"
    kubectl label configmap application-connector-overrides --overwrite installer=overrides -n kyma-installer
    kubectl label configmap application-connector-overrides --overwrite kyma-project.io/installation="" -n kyma-installer
  fi

}

function rewriteDomain() {
  DOMAIN_YAML=$(cat << EOF
---
data:
  global.applicationConnector.domainName: "${DOMAIN}"
EOF
)
  echo "---> DOMAIN used: ${DOMAIN}, patching configmap"
  set +e
  local msg
  local status
  msg=$(kubectl patch configmap application-connector-overrides --patch "${DOMAIN_YAML}" -n kyma-installer 2>&1)
  status=$?
  set -e
  if [[ $status -ne 0 ]] && [[ ! "$msg" == *"not patched"* ]]; then
    echo "$msg"
    exit $status
  fi
}

createOverrideCM
# Certs are not given, create
if [[ -z "${TLS_CRT}" ]] && [[ -z "${TLS_KEY}" ]]; then
    if [[ -z "${DOMAIN}" ]]; then
      generateXipDomain
    else
      rewriteDomain
    fi
    generateCerts
    exit 0
fi
# Certs are given, rewrite
if [[ -z "${DOMAIN}" ]]; then
  echo "Invalid setup - no domain for provided certs"
  exit 1
else
  rewriteDomain
  rewriteCerts
fi
