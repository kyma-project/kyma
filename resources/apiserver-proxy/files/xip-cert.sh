#!/usr/bin/env bash

set -o errexit

# # # # # # # # # # # # # # # # # # #
# VARs coming from environment:     #
#                                   #
# TYPE                              #
# APISERVER_SERVICE_NAME            #
# APISERVER_SERVICE_NAMESPACE       #
# APISERVER_SECRET_NAME             #
# APISERVER_SECRET_NAMESPACE        #
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


if [[ $TYPE == "legacy" ]]; then
    echo "{{ .Values.global.tlsCrt }}" | base64 --decode > /etc/ca-tls-cert/tls.crt
elif [[ $TYPE == "xip" ]]; then
    kubectl get secret -n istio-system kyma-ca-key-pair -o jsonpath='{.data.tls\.crt}' | base64 --decode > /etc/ca-tls-cert/tls.crt

    echo "Finding XIP domain for api-server LoadBalancer..."
    XIP_DOMAIN=$(generateXipDomain "${APIGATEWAY_SERVICE_NAME}" "${APIGATEWAY_SERVICE_NAMESPACE}")
    echo "XIP domain for api-server LoadBalancer: ${XIP_DOMAIN}"

    echo "Generating Certificate for ApiServer Ingress Gateway"

cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: ${APISERVER_SECRET_NAME}-cert
  namespace: ${APISERVER_SECRET_NAMESPACE}
spec:
  duration: 720h
  renewBefore: 10m
  keySize: 4096
  organization:
  - kyma
  commonName: ${XIP_DOMAIN}
  dnsNames:
  - "*.${XIP_DOMAIN}"
  secretName: ${APISERVER_SECRET_NAME}
  issuerRef:
    name: kyma-ca-issuer
    kind: ClusterIssuer
    group: cert-manager.io
EOF

echo "Success."

fi
