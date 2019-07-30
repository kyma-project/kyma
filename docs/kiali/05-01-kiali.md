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
| **server.metrics.enabled** | Specifies whether the metrics endpoint will be available for Prometheus to scrape. | `true` |
| **server.webRoot** | Defines the context root path for Kiali console, API endpoints, and readiness probes. | `/` |
| **deployment.viewOnlyMode** | When true, Kiali will be in "view only" mode, allowind the user to view and retrieve management data for the service mesh, but not allowed the user to modify the service mesh.  | `true` |
| **deployment.accessibleNamespaces** | A list of namespaces, Kiali allowed access to service mesh components deployed on those namespaces. You can provide names using regex expressions. Default value is a special value of "**" (two asterisks), which mean Kiali allowed access any namespace. | `**` |


For more about Kiali configuration and customization you can find in [Kiali getting started](https://www.kiali.io/documentation/getting-started/) documentation.