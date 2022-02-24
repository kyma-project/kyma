#!/usr/bin/env bash

echo "
Patching label merge from reconciler"
kubectl -n kyma-system patch service monitoring-alertmanager --type=json -p='[{"op": "remove", "path": "/spec/selector/app"}]'

echo "
Deleting orphaned node-exporter"
kubectl -n kyma-system delete servicemonitors.monitoring.coreos.com monitoring-node-exporter