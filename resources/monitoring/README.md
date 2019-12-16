# Monitoring

## Overview

The [Kube-Prometheus](https://github.com/coreos/kube-prometheus) implementation provides end-to-end Kubernetes cluster monitoring in [Kyma](https://github.com/kyma-project/kyma) using the [Prometheus operator](https://github.com/coreos/prometheus-operator).

This chart installs the Prometheus operator along with [Grafana](https://grafana.com/), [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics), and Prometheus [node exporter](https://github.com/prometheus/node_exporter). Once deployed, the Prometheus operator installs [Prometheus](https://prometheus.io/) and [Alertmanager](https://github.com/prometheus/alertmanager) based on the configuration specified in this chart.

All the monitoring components run in the `kyma-system` Namespace.

## Custom Resource Definitions

Custom resource definitions for `crd-alertmanager.yaml`, `crd-podmonitor.yaml`, and `crd-servicemonitor.yaml` have been moved to the `resources/cluster-essentials/templates` folder and will be maintained there. Along with the next update of the Prometheus Operator, update the listed files in the [Cluster Essentials](https://github.com/kyma-project/kyma/tree/master/resources/cluster-essentials) chart.

## Kyma customizations

All the extra files added to this chart for the usage in Kyma are collected under `kyma-additions` folders in this chart or the sub-charts.

All the in-line customizations are commented on in the respective files. Search for `Customization` to see them all.

## Kyma Dashboards

Kyma comes with some extra dashboards to enable monitoring its components. All the Kyma dashboards can be found under [kyma-dashboards](templates/grafana/kyma-dashboards). For more information on the dashboards, see [this document](charts/grafana/README.md).

## Kyma alerting rules

Kyma comes with some additional alerting rules to enable alerting for its components. You can find all Kyma alerting rules under [kyma-rules](templates/prometheus/kyma-rules).

## Details

For details on Grafana usage in Kyma, see [this document](charts/grafana/README.md).
