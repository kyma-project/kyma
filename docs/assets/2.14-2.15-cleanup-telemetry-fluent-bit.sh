#!/usr/bin/env bash

kubectl delete -n kyma-system peerauthentications.security.istio.io telemetry-fluent-bit-metrics --ignore-not-found
kubectl delete -n kyma-system configmaps telemetry-otel-collector-grafana-dashboard --ignore-not-found
kubectl delete -n kyma-system networkpolicies.networking.k8s.io telemetry-operator-pprof-deny-ingress --ignore-not-found