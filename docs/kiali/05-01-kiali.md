---
title: Kiali Chart
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
| **server.webRoot** | Defines the context root path for Kiali console, API endpoints, and readiness probes. | `/` |
| **deployment.viewOnlyMode** | When set to `true`, Kiali is available in view-only mode, allowing you to view and retrieve management data for the Service Mesh. You cannot modify the Service Mesh.  | `true` |
| **deployment.accessibleNamespaces** | Specifies the Namespaces Kiali can access to monitor the Service Mesh components deployed there. You can provide the names using regex expressions. The default value is `**`(two asterisks) meaning Kiali can access any Namespace. | `**` |
| **deployment.resources.requests.cpu** | Defines the minimum CPU that the kiali operator will use. | `10m` |
| **deployment.resources.requests.memory** | Defines the minimum amount of memory that the kiali operator will use. | `20Mi` |
| **deployment.resources.limits.cpu** | Defines the maximum CPU that can be used by the kiali operator. | `100m` |
| **deployment.resources.limits.memory** | Defines the maximum amount of memory that can be used by the kiali operator. | `100Mi` |
| **deployment.kubernetes_config.qps** | Defines the allowed queries per second to adjust the API server throttling rate. | `50` |


For details on Kiali configuration and customization, see the [Kiali CRD](https://github.com/kiali/kiali-operator/blob/master/deploy/kiali/kiali_cr.yaml) and the [values file](https://github.com/kyma-project/kyma/blob/master/resources/kiali/values.yaml).