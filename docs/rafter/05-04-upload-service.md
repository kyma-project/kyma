---
title: Upload Service sub-chart
type: Configuration
---

To configure the Upload Service sub-chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Sub-charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-sub-chart-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **replicaCount** | Defines the number of service replicas. | `1` |
| **virtualservice.enabled** |  Enables the use of an external service. If you activate the **virtualservice**, it is available at `https://{VIRTUALSERVICE_NAME}.{DOMAIN}/`, for example at `https://upload-service.kyma.local/`. | `false` |
| **virtualservice.annotations** | Defines the service annotation. | `{}` |
| **virtualservice.name** |  Defines the service name. | `upload-service` |
