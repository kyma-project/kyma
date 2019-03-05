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

if [[ -z "${DOMAIN}" ]]; then
	echo "---> DOMAIN not SET. Creating..."
	INGRESS_IP=$(getLoadBalancerIPFromLabel "${LB_LABEL}" "${LB_NAMESPACE}")
	DOMAIN="${INGRESS_IP}.xip.io"
	DOMAIN_YAML=$(cat << EOF
---
data:
  global.applicationConnectorDomainName: "${DOMAIN}"
EOF
)
	kubectl patch configmap installation-config-overrides --patch "${DOMAIN_YAML}" -n kyma-installer
fi

if [[ -z "${TLS_CRT}" ]] && [[ -z "${TLS_KEY}" ]]; then
	echo "---> Generating Certs for ${DOMAIN}"
	generateCertificatesForDomain "${DOMAIN}" /root/key.pem /root/cert.pem
	kubectl create secret tls application-connector-ingress-tls-cert --cert=/root/cert.pem --key=/root/key.pem --namespace=kyma-integration
fi
