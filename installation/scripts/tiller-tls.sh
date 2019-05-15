#!/usr/bin/env bash

RETRY_TIME=5
MAX_RETRIES=3
SECRET_NAME="helm-secret"
NAMESPACE="kyma-installer"

mkdir -p "$(helm home)"

function findHelmSecret() {
    kubectl get -n "${NAMESPACE}" secret "${SECRET_NAME}" > /dev/null || return 1
}

function defer() {
    if [[ "${i}" -eq "${MAX_RETRIES}" ]]; then return 1; fi
    echo "---> Retrying in ${RETRY_TIME} seconds..."
    sleep "${RETRY_TIME}"
}

function fail() {
    echo "---> Warning! Unable to find Helm secret: timeout."
    exit 1
}

function saveCerts {
    kubectl get -n "${NAMESPACE}" secret "${SECRET_NAME}" -o jsonpath="{.data['global\.helm\.ca\.crt']}" | base64 --decode > "$(helm home)/ca.pem"
    kubectl get -n "${NAMESPACE}" secret "${SECRET_NAME}" -o jsonpath="{.data['global\.helm\.tls\.crt']}" | base64 --decode > "$(helm home)/cert.pem"
    kubectl get -n "${NAMESPACE}" secret "${SECRET_NAME}" -o jsonpath="{.data['global\.helm\.tls\.key']}" | base64 --decode > "$(helm home)/key.pem"
}

echo "---> Finding Helm secret..."
for i in $(seq 1 "${MAX_RETRIES}"); do findHelmSecret && break || defer || fail ; done

echo "---> Helm secret found. Saving Helm certificates under the \"$(helm home)\" directory..."
saveCerts