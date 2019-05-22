#!/usr/bin/env bash

# set -o errexit

echo "The script clean-up.sh is deprecated and will be removed with Kyma release 1.4, please use Kyma CLI instead"

kubectl delete installation/kyma-installation
kubectl delete ns kyma-installer

helm del --purge dex
helm del --purge core
helm del --purge istio
helm del --purge cluster-essentials
helm del --purge prometheus-operator
helm del --purge logging
helm del --purge jaeger

kubectl delete ns kyma-system
kubectl delete ns istio-system
kubectl delete ns kyma-integration
