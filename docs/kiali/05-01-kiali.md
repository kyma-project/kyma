---
title: Kiali chart
type: Configuration
---

To configure the Kiali chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **server.metrics.enabled** | Specifies if the metrics endpoint is available for Prometheus to scrape. | `true` |
| **server.webRoot** | Defines the context root path for Kiali console, API endpoints, and readiness probes. | `/` |
| **deployment.viewOnlyMode** | When set to `true`, Kiali is available in view-only mode, allowing you to view and retrieve management data for the Service Mesh. You cannot modify the Service Mesh.  | `true` |
| **deployment.accessibleNamespaces** | Specifies the Namespaces Kiali can access to monitor the Service Mesh components deployed there. You can provide the names using regex expressions. The default value is `**`(two asterisks) meaning Kiali can access any Namespace. | `**` |


For details on Kiali configuration and customization, see [this](https://www.kiali.io/documentation/getting-started/) documentation.
