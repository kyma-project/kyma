#!/usr/bin/env bash

echo "Deleting logging Fluent Bit resources"

kubectl delete -n kyma-system clusterroles.rbac.authorization.k8s.io logging-fluent-bit
kubectl delete -n kyma-system clusterrolebindings.rbac.authorization.k8s.io logging-fluent-bit
kubectl delete -n kyma-system configmap logging-fluent-bit
kubectl delete -n kyma-system configmap logging-fluent-bit-grafana-dashboard
kubectl delete -n kyma-system daemonsets.apps logging-fluent-bit
kubectl delete -n kyma-system peerauthentications.security.istio.io logging-fluent-bit-metrics
kubectl delete -n kyma-system podsecuritypolicys.policy 000-logging-fluent-bit
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com logging-fluent-bit-kyma-fluent-bit.rules
kubectl delete -n kyma-system secret logging-fluent-bit-es-ca-secret
kubectl delete -n kyma-system service logging-fluent-bit
kubectl delete -n kyma-system serviceaccount logging-fluent-bit
kubectl delete -n kyma-system servicemonitors.monitoring.coreos.com logging-fluent-bit
