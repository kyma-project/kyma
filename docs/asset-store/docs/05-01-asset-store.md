---
title: Asset Store chart
type: Configuration
---

To configure the Asset Store chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

To learn more about how to use overrides in Kyma, read [this document](/kyma/#configuration-helm-overrides-for-kyma-installation). See also examples of [top-level charts overrides](/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides) and [sub-charts overrides](https://kyma-project.io/docs/master/root/kyma/#configuration-helm-overrides-for-kyma-installation-sub-chart-overrides).

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **minio.mode** | Specifies Minio mode. | `standalone` |
| **minio.environment.MINIO_BROWSER** | Enables browsing Minio storage. By deafult, the Minio browser is turned off for security reasons. You can change the value to `on` to use the browser. | `"off"` |
| **resources.requests.memory** | Defines requests for memory resources. | `32Mi` |
| **resources.requests.cpu** |   Defines requests for CPU resources. | `10m` |
| **resources.limits.memory** |  Defines limits for memory resources. | `128Mi` |
| **resources.limits.cpu** | Defines limits for CPU resources. | `100m` |
