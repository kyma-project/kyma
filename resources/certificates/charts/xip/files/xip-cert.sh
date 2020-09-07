#!/usr/bin/env bash

set -o errexit

# # # # # # # # # # # # # # # # # # #
# VARs coming from environment:     #
#                                   #
# INGRESSGATEWAY_SERVICE_NAME       #
# INGRESSGATEWAY_SERVICE_NAMESPACE  #
# ROOTCA_SECRET_NAME                #
# ROOTCA_SECRET_NAMESPACE           #
# CLUSTER_INFO_CM_NAME              #
# CLUSTER_INFO_CM_NAMESPACE         #
# # # # # # # # # # # # # # # # # # #


getLoadBalancerIP() {

    if [ "$#" -ne 2 ]; then
        echo "usage: getLoadBalancerIP <service_name> <namespace>"
        exit 1
    fi

    local SERVICE_NAME="$1"
    local SERVICE_NAMESPACE="$2"
    local LOAD_BALANCER_IP=""

    SECONDS=0
    END_TIME=$((SECONDS+60))

    while [ ${SECONDS} -lt ${END_TIME} ];do

        LOAD_BALANCER_IP=$(kubectl get service -n "${SERVICE_NAMESPACE}" "${SERVICE_NAME}" -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

        if [ -n "${LOAD_BALANCER_IP}" ]; then
            break
        fi

        sleep 10

    done

    if [ -z "${LOAD_BALANCER_IP}" ]; then
        echo "---> Could not retrive the IP address. Verify if service ${SERVICE_NAME} exists in the namespace ${SERVICE_NAMESPACE}" >&2
        echo "---> Command executed: kubectl get service -n ${SERVICE_NAMESPACE} ${SERVICE_NAME} -o jsonpath='{.status.loadBalancer.ingress[0].ip}'" >&2
        exit 1
    fi

    echo "${LOAD_BALANCER_IP}"
}

generateXipDomain() {

    if [ "$#" -ne 2 ]; then
        echo "usage: generateXipDomain <service_name> <namespace>"
        exit 1
    fi

    local SERVICE_NAME="$1"
    local SERVICE_NAMESPACE="$2"
    local EXTERNAL_PUBLIC_IP
    EXTERNAL_PUBLIC_IP=$(getLoadBalancerIP "${SERVICE_NAME}" "${SERVICE_NAMESPACE}")

    if [[ "$?" != 0 ]]; then
        echo "External public IP not found"
        exit 1
    fi

    echo "${EXTERNAL_PUBLIC_IP}.xip.io"
}

generateRootCACerts() {

    if [ "$#" -ne 3 ]; then
        echo "usage: generateRootCACerts <ingress_domain> <rootca_secret_name> <rootca_secret_namespace>"
        exit 1
    fi

    local INGRESS_DOMAIN="$1"
    local ROOTCA_SECRET_NAME="$2"
    local ROOTCA_SECRET_NAMESPACE="$3"

    # Just to be sure nothing's there
    rm -f /tmp/ca.key
    rm -f /tmp/ca.crt

    # Generate a Root CA private key
    openssl genrsa -out /tmp/ca.key 4096

    # Create a Root CA: self signed Certificate, valid for 10yrs with the 'signing' option set
    openssl req -x509 -new -nodes -key /tmp/ca.key -subj "/CN=$INGRESS_DOMAIN" -days 3650 -reqexts v3_req -extensions v3_ca -out /tmp/ca.crt

    # Store Root CA key pair as secret (necessary for cert-manager to issue certificates based on the Root CA)
    kubectl create secret tls "${ROOTCA_SECRET_NAME}" \
      --cert=/tmp/ca.crt \
      --key=/tmp/ca.key \
      --namespace="${ROOTCA_SECRET_NAMESPACE}"

    # Cleanup
    rm -f /tmp/ca.key
    rm -f /tmp/ca.crt
}

echo "Finding XIP domain..."
XIP_DOMAIN=$(generateXipDomain "${INGRESSGATEWAY_SERVICE_NAME}" "${INGRESSGATEWAY_SERVICE_NAMESPACE}")
echo "XIP domain: ${XIP_DOMAIN}"

echo "Generating Root CA for XIP domain..."
generateRootCACerts "${XIP_DOMAIN}" "${ROOTCA_SECRET_NAME}" "${ROOTCA_SECRET_NAMESPACE}"

echo "Generating ClusterIssuer"

cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1alpha2
kind: ClusterIssuer
metadata:
  name: kyma-ca-issuer
  namespace: ${ROOTCA_SECRET_NAMESPACE}
spec:
  ca:
    secretName: ${ROOTCA_SECRET_NAME}
EOF

echo "Generating Certificate for Istio Ingress Gateway"

cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  name: kyma-gateway-certs
  namespace: ${ROOTCA_SECRET_NAMESPACE}
spec:
  duration: 720h
  renewBefore: 10m
  keySize: 4096
  organization:
  - kyma
  commonName: ${XIP_DOMAIN}
  dnsNames:
  - "*.${XIP_DOMAIN}"
  secretName: kyma-gateway-certs
  issuerRef:
    name: kyma-ca-issuer
    kind: ClusterIssuer
    group: cert-manager.io
EOF


echo "Update global.ingress.domainName override in: ${CLUSTER_INFO_CM_NAMESPACE}/${CLUSTER_INFO_CM_NAME}"

PATCH_YAML=$(cat << EOF
---
data:
  global.ingress.domainName: ${XIP_DOMAIN}
EOF
)

echo "---> Patching cm ${CLUSTER_INFO_CM_NAMESPACE}/${CLUSTER_INFO_CM_NAME}"
set +e
msg=$(kubectl patch cm ${CLUSTER_INFO_CM_NAME} --patch "${PATCH_YAML}" -n ${CLUSTER_INFO_CM_NAMESPACE} 2>&1)
status=$?
set -e

if [[ $status -ne 0 ]] && [[ ! "$msg" == *"not patched"* ]]; then
    echo "error patching ConfigMap ${CLUSTER_INFO_CM_NAME}"
    echo "$msg"
    exit $status
fi

echo "Success."
