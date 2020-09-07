#!/usr/bin/env bash

set -o errexit

for var in KYMA_TLS_CERT GLOBAL_DOMAIN APISERVER_PROXY_TLS_CERT KYMA_CA_ISSUER KYMA_TLS_SECRET APISERVER_PROXY_TLS_SECRET; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

echo "Checking if cert-manager CRDs are installed"

kubectl get crd certificates.cert-manager.io
kubectl get crd clusterissuers.cert-manager.io

echo "Checking if cert-manager ClusterIssuer exists"

kubectl get clusterissuers.cert-manager.io ${KYMA_CA_ISSUER} -n istio-system

echo "Checking if cert-manager Certificates exist"

kubectl get certificates.cert-manager.io ${KYMA_TLS_CERT} -n istio-system
kubectl get certificates.cert-manager.io ${APISERVER_PROXY_TLS_CERT} -n kyma-system

echo "Checking if Secrets exist"

kubectl get secret ${KYMA_TLS_SECRET} -n istio-system
kubectl get secret ${APISERVER_PROXY_TLS_SECRET} -n kyma-system