---
title: Service Catalog chart
type: Configuration
---

To configure the Service Catalog chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)


## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **etcd-stateful.replicaCount** | Specifies the number of members in an etcd cluster. | `3` |
| **etcd-stateful.etcd.resources.limits.memory** | Defines limits for memory resources. | `512Mi` |

>**NOTE:** Overriding values in this chart has priority over overriding values in the `etcd-stateful` chart.

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **service-catalog-apiserver.enabled** | Enables Service Catalog with the Aggregated API Server. | `true` |
| **service-catalog-crds.enabled** | Enables Service Catalog with the CRDs implementation. | `false` |

>**CAUTION:** These values are mutually exclusive and they cannot be both enabled or disabled.
