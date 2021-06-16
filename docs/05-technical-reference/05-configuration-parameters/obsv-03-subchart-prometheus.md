---
title: Prometheus sub-chart
---

To configure the Prometheus sub-chart, override the default values of its `values.yaml` file. 
Learn how it works under [Configurable Parameters](./README.md).

You can set the following parameters:

## Configurable parameters

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **prometheusSpec.retention** | Specifies a period for which Prometheus stores the metrics.| `1d` |
| **prometheusSpec.retentionSize** | Specifies the maximum number of bytes that storage blocks can use. The oldest data will be removed first.| `2GB` |
| **prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage** | Specifies the size of a PersistentVolumeClaim (PVC). | `10Gi` |
