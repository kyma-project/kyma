#!/usr/bin/env bash

getLoadBalancerIP() {

    if [ "$#" -ne 2 ]; then
        echo "usage: getLoadBalancerIP <service_name> <namespace>"
        exit 1
    fi

    SERVICE_NAME="$1"
    NAMESPACE="$2"
    LOAD_BALANCER_IP=""

    SECONDS=0
    END_TIME=$((SECONDS+60))

    while [ ${SECONDS} -lt ${END_TIME} ];do

        LOAD_BALANCER_IP=$(kubectl get service -n "${NAMESPACE}" "${SERVICE_NAME}" -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

        if [ -n "${LOAD_BALANCER_IP}" ]; then
            break
        fi

        sleep 10

    done

    if [ -z "${LOAD_BALANCER_IP}" ]; then
        echo "---> Could not retrive the IP address. Verify if service ${SERVICE_NAME} exists in the namespace ${NAMESPACE}" >&2
        echo "---> Command executed: kubectl get service -n ${NAMESPACE} ${SERVICE_NAME} -o jsonpath='{.status.loadBalancer.ingress[0].ip}'" >&2
        exit 1
    fi

    echo "${LOAD_BALANCER_IP}"
}

getLoadBalancerIPFromLabel() {

    if [ "$#" -ne 2 ]; then
        echo "usage: getLoadBalancerIP <label> <namespace>"
        exit 1
    fi

    LABEL="$1"
    NAMESPACE="$2"
    LOAD_BALANCER_IP=""

    SECONDS=0
    END_TIME=$((SECONDS+60))

    while [ ${SECONDS} -lt ${END_TIME} ];do

        LOAD_BALANCER_IP=$(kubectl get service -l "${LABEL}" -o jsonpath='{.items[0].status.loadBalancer.ingress[0].ip}' -n "${NAMESPACE}")

        if [ -n "${LOAD_BALANCER_IP}" ]; then
            break
        fi

        sleep 10

    done

    if [ -z "${LOAD_BALANCER_IP}" ]; then
        echo "---> Could not retrive the IP address. Verify if any service has the label ${LABEL} in the namespace ${NAMESPACE}" >&2
        echo "---> Command executed: kubectl get service -l ${LABEL} -o jsonpath='{.items[0].status.loadBalancer.ingress[0].ip}' -n ${NAMESPACE}" >&2
        exit 1
    fi

    echo "${LOAD_BALANCER_IP}"
}

generateCertificatesForDomain() {

    if [ "$#" -ne 3 ]; then
        echo "usage: generateCertificatesForDomain <domain> <key_output_file> <cert_output_file>"
        exit 1
    fi

    DOMAIN="$1"
    KEY_PATH="$2"
    CERT_PATH="$3"

    openssl req -x509 -nodes -days 30 -newkey rsa:4069 \
                 -subj "/CN=${DOMAIN}" \
                 -reqexts SAN -extensions SAN \
                 -config <(cat /etc/ssl/openssl.cnf \
        <(printf "\\n[SAN]\\nsubjectAltName=DNS:*.%s" "${DOMAIN}")) \
                 -keyout "${KEY_PATH}" \
                 -out "${CERT_PATH}"
}
