---
title: Asset Store Metadata Service sub-chart
type: Configuration
---

To configure the Asset Store Metadata Service sub-chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents: 
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Sub-charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-sub-chart-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **replicaCount** | Defines the number of replicas of the service. For more details, see the [Kubernetes documentation](https://kubernetes.io/docs/concepts/workloads/controllers/replicaset/).| `1` |
| **virtualservice.enabled** | Enables to use an external service. If you enable the **virtualservice**, it is available at `https://{VIRTUALSERVICE_NAME}.{DOMAIN}/minio/`, for example `https://asset-metadata-service.kyma.local/`. | `false` |
| **virtualservice.annotations** | Defines the service annotation. | `{}` |
| **virtualservice.name** | Defines the service name. | `asset-metadata-service` |
