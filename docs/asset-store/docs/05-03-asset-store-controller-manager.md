---
title: Asset Store Controller Manager sub-chart
type: Configuration
---

To configure the Asset Store Controller Manager sub-chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:


| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **resources.limits.cpu** | Defines limits for CPU resources. | `100m` |
| **resources.limits.memory** | Defines limits for memory resources. | `128Mi` |
| **resources.requests.cpu** | Defines requests for CPU resources. | `100m` |
| **resources.requests.memory** | Defines requests for memory resources. | `64Mi` |
