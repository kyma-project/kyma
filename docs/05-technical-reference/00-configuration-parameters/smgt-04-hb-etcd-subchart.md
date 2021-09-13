---
title: Helm Broker - Etcd-stateful sub-chart
---

To configure the Etcd-stateful sub-chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** See how to [change Kyma settings](../../04-operation-guides/operations/03-change-kyma-config-values.md).

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **etcd.resources.limits.cpu** | Defines limits for CPU resources. | `200m` |
| **etcd.resources.limits.memory** | Defines limits for memory resources. | `256Mi` |
| **etcd.resources.requests.cpu** | Defines requests for CPU resources. | `50m` |
| **etcd.resources.requests.memory** | Defines requests for memory resources. | `64Mi` |
| **replicaCount** | Defines the size of the etcd cluster. | `1` |
