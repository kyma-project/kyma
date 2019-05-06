---
title: Asset Store chart
type: Configuration
---

To configure the Asset Store chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:


| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **minio.mode** |  | `standalone` |
| **minio.accessKey** |       | `admin` |
| **minio.secretKey** |       | `topSecretKey` |
| **minio.environment.MINIO_BROWSER** |     | `"off"` |
| **resources.requests.memory** |      | `32Mi` |
| **resources.requests.cpu** |      | `10m` |
| **resources.limits.memory** |      | `128Mi` |
| **resources.limits.cpu** |      | `100m` |
