#!/usr/bin/env bash

set -o errexit

# prepareSystemNs is responsible for creating system namespace with all required components, such us secret for Docker registry, LimitRange etc.
function prepareSystemNs() {
    nsName=$1
    kubectl create namespace ${nsName}
    kubectl apply -f /kyma/resources/cluster-prerequisites/limit-range.yaml -n ${nsName}
}

prepareSystemNs "istio-system"
prepareSystemNs "kyma-system"
prepareSystemNs "kyma-integration"

kubectl label namespace kyma-system "istio-injection=enabled"
kubectl label namespace kyma-integration "istio-injection=enabled"

kubectl apply -f /kyma/resources/cluster-prerequisites/remote-environments-minio-secret.yaml -n "kyma-integration"

kubectl apply -f /kyma/resources/cluster-prerequisites/resource-quotas.yaml
