#!/usr/bin/env bash

function getLoadBalancerIP() {

    if [ "$#" -ne 2 ]; then
        echo "usage: getLoadBalancerIP <service_name> <namespace>"
        exit 1
    fi

    SERVICE_NAME="$1"
    NAMESPACE="$2"
    LOAD_BALANCER_IP=""

    SECONDS=0
    END_TIME=$((SECONDS+5))

    while [ ${SECONDS} -lt ${END_TIME} ];do

        LOAD_BALANCER_IP=$(kubectl get service -n "${NAMESPACE}" "${SERVICE_NAME}" -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

        if [ -n "${LOAD_BALANCER_IP}" ]; then
            break
        fi

        sleep 10

    done

    if [ -z "${LOAD_BALANCER_IP}" ]; then
        exit 1
    fi

    echo "${LOAD_BALANCER_IP}"
}

function generateCertificatesForDomain() {

    if [ "$#" -ne 3 ]; then
        echo "usage: generateCertificatesForDomain <domain> <key_output_file> <cert_output_file>"
        exit 1
    fi

    DOMAIN="$1"
    KEY_PATH="$2"
    CERT_PATH="$3"

    openssl req -x509 -nodes -days 5 -newkey rsa:4069 \
                 -subj "/CN=${DOMAIN}" \
                 -reqexts SAN -extensions SAN \
                 -config <(cat /etc/ssl/openssl.cnf \
        <(printf "\\n[SAN]\\nsubjectAltName=DNS:*.%s" "${DOMAIN}")) \
                 -keyout "${KEY_PATH}" \
                 -out "${CERT_PATH}"
}