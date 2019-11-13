---
title: Front Matter Service sub-chart
type: Configuration
---

To configure the Front Matter Service sub-chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Sub-charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-sub-chart-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **replicaCount** | Defines the number of service replicas. For more details, see the [Kubernetes documentation](https://kubernetes.io/docs/concepts/workloads/controllers/replicaset/).| `1` |
| **virtualservice.enabled** | Enables the use of an external service. If you activate the **virtualservice**, it is available at `https://{VIRTUALSERVICE_NAME}.{DOMAIN}/`, for example at `https://asset-metadata-service.kyma.local/`. | `false` |
| **virtualservice.annotations** | Defines the service annotation. | `{}` |
| **virtualservice.name** | Defines the service name. | `front-matter-service` |
