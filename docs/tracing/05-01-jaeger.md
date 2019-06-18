---
title: Jaeger chart
type: Configuration
---

To configure the Jaeger chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **jaeger.memory.maxTraces** | Defines the maximum number of traces that Jaeger can store. | `40000` |
| **resources.limits.memory** | Defines the maximum amount of memory that is available for storing traces in Jaeger. | `512M` |
