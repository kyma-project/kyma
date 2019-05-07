---
title: Asset Store chart
type: Configuration
---

To configure the Asset Store chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents: 
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **minio.mode** | Specifies Minio mode. For details about switching to Minio Gateway mode, see [this document](#tutorials-set-minio-to-the-google-cloud-storage-gateway-mode) | `standalone` |
| **minio.environment.MINIO_BROWSER** | Enables to browse Minio storage. By deafult, the Minio browser is turned off for security reasons. You can change the value to `on` to use the browser. If you enable the browser, it is available at `https://minio.{DOMAIN}/minio/`, for example `https://minio.kyma.local/minio/`. | `"off"` |
| **resources.requests.memory** | Defines requests for memory resources. | `32Mi` |
| **resources.requests.cpu** |   Defines requests for CPU resources. | `10m` |
| **resources.limits.memory** |  Defines limits for memory resources. | `128Mi` |
| **resources.limits.cpu** | Defines limits for CPU resources. | `100m` |
