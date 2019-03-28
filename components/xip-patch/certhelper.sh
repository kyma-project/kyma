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
  global.applicationConnector.domainName: "${DOMAIN}"
EOF
)
  echo "---> DOMAIN created: ${DOMAIN}, patching configmap"
  kubectl patch configmap application-connector-overrides --patch "${DOMAIN_YAML}" -n kyma-installer
}

function generateCerts() {
  echo "---> Generating Certs for ${DOMAIN}"
  generateCertificatesForDomain "${DOMAIN}" /root/key.pem /root/cert.pem
  CERT=$(base64 /root/cert.pem | tr -d '\n')
  KEY=$(base64 /root/key.pem | tr -d '\n')
  rm /root/key.pem /root/cert.pem
  TLS_CERT_AND_KEY_YAML=$(cat << EOF
---
data:
  global.applicationConnector.tlsCrt: "${CERT}"
  global.applicationConnector.tlsKey: "${KEY}"
EOF
)
  echo "---> Certs have been created, creating patching configmap"
  kubectl patch configmap application-connector-overrides --patch "${TLS_CERT_AND_KEY_YAML}" -n kyma-installer
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

function rewriteCerts() {
  echo "---> Certs have been given, patching configmap"
  TLS_CERT_AND_KEY_YAML=$(cat << EOF
---
data:
  global.applicationConnector.tlsCrt: "${TLS_CRT}"
  global.applicationConnector.tlsKey: "${TLS_KEY}"
EOF
)
  echo "---> Certs have been created, creating patching configmap"
  set +e
  local msg
  local status
  msg=$(kubectl patch configmap application-connector-overrides --patch "${TLS_CERT_AND_KEY_YAML}" -n kyma-installer 2>&1)
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
# Certa are given, rewrite
if [[ -z "${DOMAIN}" ]]; then
  echo "Invalid setup - no domain for provided certs"
  exit 1
else
  rewriteDomain
  rewriteCerts
fi
