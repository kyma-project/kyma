---
title: Prometheus sub-chart
type: Configuration Parameters
---

To configure the Prometheus sub-chart, override the default values of its `values.yaml` file. This document describes parameters that you can set.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Sub-charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-sub-chart-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **prometheusSpec.retention** | Specifies a period for which Prometheus stores the metrics.| `1d` |
| **prometheusSpec.retentionSize** | Specifies the maximum number of bytes that storage blocks can use. The oldest data will be removed first.| `2GB` |
| **prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage** | Specifies the size of a PersistentVolumeClaim (PVC). | `10Gi` |
