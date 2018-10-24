#!/usr/bin/env bash

# set -o errexit

kubectl delete installation/kyma-installation
kubectl delete ns kyma-installer

helm del --purge ec-default
helm del --purge hmc-default
helm del --purge dex
helm del --purge core
helm del --purge istio
helm del --purge cluster-essentials
helm del --purge prometheus-operator
helm del --purge logging

kubectl delete ns kyma-system
kubectl delete ns istio-system
kubectl delete ns kyma-integration