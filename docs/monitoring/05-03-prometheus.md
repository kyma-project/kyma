---
title: Prometheus sub-chart
type: Configuration
---

To configure the Prometheus sub-chart, override the default values of its `values.yaml` file. This document describes parameters that you can set.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Sub-charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-sub-chart-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **retention** | Specifies a period for which Prometheus stores the metrics in-memory. This retention time applies to in-memory storage only. Prometheus stores the recent data in-memory for the specified amount of time to avoid reading the entire data from disk.| `2h` |
| **storageSpec.volumeClaimTemplate.spec.resources.requests.storage** | Specifies the size of a Persistent Volume Claim (PVC). | `4Gi` |
