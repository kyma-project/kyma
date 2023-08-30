#!/usr/bin/env bash

kubectl delete -n kyma-system authorizationpolicies.security.istio.io logging-loki --ignore-not-found
kubectl delete -n kyma-system configmaps logging-loki-grafana-dashboard --ignore-not-found
kubectl delete -n kyma-system configmaps logging-loki-grafana-datasource --ignore-not-found
kubectl delete -n kyma-system logpipelines.telemetry.kyma-project.io loki --ignore-not-found
kubectl delete -n kyma-system peerauthentications.security.istio.io logging-loki --ignore-not-found
kubectl delete -n kyma-system roles.rbac.authorization.k8s.io logging-loki --ignore-not-found
kubectl delete -n kyma-system rolebindings.rbac.authorization.k8s.io logging-loki --ignore-not-found
kubectl delete -n kyma-system secrets logging-loki --ignore-not-found
kubectl delete -n kyma-system services logging-loki --ignore-not-found
kubectl delete -n kyma-system services logging-loki-headless --ignore-not-found
kubectl delete -n kyma-system serviceaccounts logging-loki --ignore-not-found
kubectl delete -n kyma-system statefulsets.apps logging-loki --ignore-not-found
kubectl delete -n kyma-system persistentvolumeclaims storage-logging-loki-0 --ignore-not-found
