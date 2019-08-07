---
title: Jaeger chart
type: Configuration
---

To configure the Jaeger chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **resources.limits.memory** | Defines the maximum amount of memory that is available for storing traces in Jaeger. | `128M` |
| **jaeger.persistence.storageType** | Defines storage type for span data. | `badger` |
| **jaeger.persistence.dataPath** | Directory path where span data will be stored. | `/badger/data` |
| **jaeger.persistence.keyPath** | Directory path where data keys will be stored. | `/badger/key` |
| **jaeger.persistence.ephemeral** | Defines whether storage using temporary file system or not. | `false` |
| **jaeger.persistence.accessModes** | Access mode settings for persistence volume claim (PVC). | `ReadWriteOnce` |
| **jaeger.persistence.size** | Defines disk size will be used from persistence volume claim. | `1Gi` |
| **jaeger.persistence.storageClassName** | Defines persistence volume claim storage class name. | `` |


