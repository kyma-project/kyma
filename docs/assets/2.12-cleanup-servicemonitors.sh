#!/usr/bin/env bash

kubectl delete -n kyma-system --ignore-not-found=true servicemonitors.monitoring.coreos.com eventing-controller
kubectl delete -n kyma-system --ignore-not-found=true servicemonitors.monitoring.coreos.com eventing-nats
kubectl delete -n kyma-system --ignore-not-found=true servicemonitors.monitoring.coreos.com eventing-publisher-proxy
kubectl delete -n kyma-system --ignore-not-found=true servicemonitors.monitoring.coreos.com telemetry-fluent-bit
kubectl delete -n kyma-system --ignore-not-found=true servicemonitors.monitoring.coreos.com telemetry-operator-metrics
