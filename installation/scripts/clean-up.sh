#!/usr/bin/env bash

# set -o errexit

echo "The clean-up.sh script is deprecated and will be removed. Use Kyma CLI instead."

kubectl delete installation/kyma-installation
kubectl delete ns kyma-installer

helm del --purge dex
helm del --purge core
helm del --purge istio
helm del --purge cluster-essentials
helm del --purge logging
helm del --purge jaeger

kubectl delete ns kyma-system
kubectl delete ns istio-system
kubectl delete ns kyma-integration
