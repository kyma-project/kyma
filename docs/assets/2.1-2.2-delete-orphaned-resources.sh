#!/usr/bin/env bash

kubectl delete -n kyma-system configmap monitoring-kyma-pods-grafana-dashboard
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-kyma-general.rules
kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-kyma-prometheus-operator.rules
