#!/usr/bin/env bash
# EXPECTED ENVS
# - DOMAIN (optional) - Static domain for which to generate certs
# - TLS_CERT (optinal) - Path to current TLS certificate
# - TLS_KEY (optional) - Path to current TLS cert key
# - LB_SERVICE_NAME (required) - Name of LoadBalancer Service
# - LB_SERVICE_NS (required) - Namespace of LoadBalancer Service

set -e

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source $CURRENT_DIR/utils.sh

discoverUnsetVar=false

for var in LB_SERVICE_NAME LB_SERVICE_NS; do
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
	INGRESS_IP=$(getLoadBalancerIP "${LB_SERVICE_NAME}" "${LB_SERVICE_NS}")
	DOMAIN="${INGRESS_IP}.xip.io"
fi

if [[ -z "$(cat $TLS_CERT)" ]] && [[ -z "$(cat $TLS_CERT)" ]]; then
	echo "---> Generating Certs for ${DOMAIN}"
	generateCertificatesForDomain "${DOMAIN}" "${TLS_CERT}" "${TLS_CERT}"
fi
