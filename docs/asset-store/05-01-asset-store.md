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
| **minio.persistence.enabled** | Enables MinIO persistence. Deactivate it only if you use the Gateway mode. For more details about how to switch to the MinIO Gateway mode, see [this document](#tutorials-set-minio-to-the-google-cloud-storage-gateway-mode). | `true` |
| **minio.environment.MINIO_BROWSER** | Enables browsing MinIO storage. By default, the MinIO browser is turned off for security reasons. You can change the value to `on` to use the browser. If you enable the browser, it is available at `https://minio.{DOMAIN}/minio/`, for example at `https://minio.kyma.local/minio/`. | `"off"` |
| **minio.resources.requests.memory** | Defines requests for memory resources. | `32Mi` |
| **minio.resources.requests.cpu** |  Defines requests for CPU resources. | `10m` |
| **minio.resources.limits.memory** |  Defines limits for memory resources. | `128Mi` |
| **minio.resources.limits.cpu** | Defines limits for CPU resources. | `100m` |
