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

kubectl get clusterissuers.cert-manager.io ${KYMA_CA_ISSUER}

echo "Checking if cert-manager Certificates exist"

KYMA_TLS_CERT_SECRET=$(kubectl get certificates.cert-manager.io ${KYMA_TLS_CERT} -n istio-system -o jsonpath='{.spec.secretName}')
APISERVER_PROXY_TLS_CERT_SECRET=$(kubectl get certificates.cert-manager.io ${APISERVER_PROXY_TLS_CERT} -n kyma-system -o jsonpath='{.spec.secretName}')

if [ $KYMA_TLS_CERT_SECRET != $KYMA_TLS_SECRET ]; then
  echo "Wrong secret name in the Certificate $KYMA_TLS_CERT"
  exit 1
fi

if [ $APISERVER_PROXY_TLS_CERT_SECRET != $APISERVER_PROXY_TLS_SECRET ]; then
  echo "Wrong secret name in the Certificate $APISERVER_PROXY_TLS_CERT"
  exit 1
fi

echo "Checking if Secrets exist"

SECONDS=0
END_TIME=$((SECONDS+180))

while [ ${SECONDS} -lt ${END_TIME} ]; do

  KYMA_TLS_SECRET_PRESENT=false
  APISERVER_PROXY_TLS_SECRET_PRESENT=false

  if kubectl get secret ${KYMA_TLS_SECRET} -n istio-system; then
    KYMA_TLS_SECRET_PRESENT=true
  fi

  if kubectl get secret ${APISERVER_PROXY_TLS_SECRET} -n kyma-system; then
    APISERVER_PROXY_TLS_SECRET_PRESENT=true
  fi

  if [ $KYMA_TLS_SECRET_PRESENT == "true" ] && [ $APISERVER_PROXY_TLS_SECRET_PRESENT == "true" ]; then
    echo "Secrets are ready. Exiting..."
    exit 0
  fi

  echo "Secrets are not ready. Sleeping 10 seconds..."
  sleep 10

done

exit 1