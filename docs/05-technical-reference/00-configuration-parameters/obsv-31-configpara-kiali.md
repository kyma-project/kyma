---
title: Kiali Chart
---

To configure the Kiali chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

> **TIP:** See how to [change Kyma settings](https://github.com/kyma-project/kyma/blob/kyma-2.0-docu/docs/04-operation-guides/operations/03-change-kyma-config-values.md).
## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **kiali.spec.server.web_root** | Defines the context root path for Kiali console, API endpoints, and readiness probes. | `/` |
| **kiali.spec.deployment.view_only_mode** | When set to `true`, Kiali is available in view-only mode, allowing you to view and retrieve management data for the Service Mesh. You cannot modify the Service Mesh.  | `true` |
| **kiali.spec.deployment.accessible_namespaces** | Specifies the Namespaces Kiali can access to monitor the Service Mesh components deployed there. You can provide the names using regex expressions. The default value is `**`(two asterisks) meaning Kiali can access any Namespace. | `**` |
| **kiali.spec.deployment.resources.requests.cpu** | Minimum number of CPUs requested by the Kiali Deployment to use. | `10m` |
| **kiali.spec.deployment.resources.requests.memory** | Minimum amount of memory requested by the Kiali Deployment to use. | `20Mi` |
| **kiali.spec.deployment.resources.limits.cpu** | Maximum number of CPUs available for the Kiali Deployment to use. | `100m` |
| **kiali.spec.deployment.resources.limits.memory** | Maximum amount of memory available for the Kiali Deployment to use. | `100Mi` |
| **kiali.spec.kubernetes_config.qps** | Defines the allowed queries per second to adjust the API server's throttling rate. | `50` |

For more details on Kiali configuration and customization, see the [`values.yaml`](https://github.com/kyma-project/kyma/blob/master/resources/kiali/values.yaml) file.
