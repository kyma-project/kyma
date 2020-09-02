#!/usr/bin/env bash

set -o errexit

echo "Checking if cert-manager CRDs are installed"

kubectl get crd certificates.cert-manager.io
kubectl get crd clusterissuers.cert-manager.io

echo "Checking if cert-manager Certificates exist"

kubectl get certificates.cert-manager.io ${KYMA_TLS_CERT} -n istio-system
kubectl get certificates.cert-manager.io ${APISERVER_PROXY_TLS_CERT} -n kyma-system

echo "Checking if cert-manager ClusterIssuer exists"

kubectl get clusterissuers.cert-manager.io ${KYMA-CA-ISSUER} -n istio-system

echo "Checking if Secrets exist"

kubectl get secret ${KYMA_TLS_SECRET} -n istio-system
kubectl get secret ${APISERVER_PROXY_TLS_SECRET} -n kyma-system