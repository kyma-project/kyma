#!/usr/bin/env bash

kubectl delete -n kyma-system prometheusrules.monitoring.coreos.com monitoring-kube-apiserver-slos
kubectl delete -n kyma-system serviceaccount kiali
kubectl delete -n kyma-system configmap monitoring-apiserver-per-client-grafana-dashboard
kubectl delete -n kyma-system configmap monitoring-kyma-backends-grafana-dashboard
kubectl delete -n kyma-system configmap service-binding-usage-controller-dashboard
kubectl delete -n kyma-system configmap monitoring-apiserver-grafana-dashboard
kubectl delete -n kyma-system configmap monitoring-kyma-frontends-grafana-dashboard
kubectl delete -n kyma-system configmap dockerfile-python-38
kubectl delete -n kyma-system configmap rafter-controller-manager-dashboard
kubectl delete -n kyma-system servicemonitors.monitoring.coreos.com monitoring-node-exporter
kubectl delete -n kyma-system configmap monitoring-alertmanager-grafana-dashboard