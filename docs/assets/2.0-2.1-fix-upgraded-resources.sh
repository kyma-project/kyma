#!/usr/bin/env bash

echo "
Patching label merge from reconciler"
kubectl -n kyma-system patch service monitoring-alertmanager --type=json -p='[{"op": "remove", "path": "/spec/selector/app"}]'

echo "
Deleting orphaned node-exporter"
kubectl -n kyma-system delete servicemonitors.monitoring.coreos.com monitoring-node-exporter

echo "
Deleting orphaned Grafana dashboards"
kubectl -n kyma-system delete configmap rafter-controller-manager-dashboard
kubectl -n kyma-system delete configmap service-binding-usage-controller-dashboard
