#!/usr/bin/env bash

kubectl delete -n kyma-system peerauthentications.security.istio.io telemetry-fluent-bit-metrics --ignore-not-found
kubectl delete -n kube-public configmaps logparsers.telemetry.kyma-project.io --ignore-not-found
kubectl delete -n kube-public configmaps logpipelines.telemetry.kyma-project.io --ignore-not-found
kubectl delete -n kube-public configmaps tracepipelines.telemetry.kyma-project.io --ignore-not-found