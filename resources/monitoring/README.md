# Monitoring

## Overview

The [Kube-Prometheus](https://github.com/coreos/kube-prometheus) implementation provides end-to-end Kubernetes cluster monitoring in [Kyma](https://github.com/kyma-project/kyma) using the [Prometheus operator](https://github.com/coreos/prometheus-operator).

This chart installs the Prometheus operator along with [Grafana](https://grafana.com/), [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics), and Prometheus [node exporter](https://github.com/prometheus/node_exporter). Once deployed, the Prometheus operator installs [Prometheus](https://prometheus.io/) and [Alertmanager](https://github.com/prometheus/alertmanager) based on the configuration specified in this chart.

All the monitoring components run in the `kyma-system` Namespace.

## Details

* For details on Grafana implementation in Kyma, see [this document](charts/grafana/README.md).
