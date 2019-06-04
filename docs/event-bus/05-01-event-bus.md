---
title: Event Bus chart
type: Configuration
---

To configure the Event Bus chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **global.publishKnative.maxRequests** | Specifies the maximum number of parallel Event requests that **publishKnative** can process. If you raise this value, you may also have to increase memory resources for the Event Bus to handle the higher load. | `16` |
| **global.publishKnative.maxRequestSize** | Specifies the maximum size of one Event. If you raise this value, you may also have to increase memory resources for the Event Bus to handle the higher load. | `65536` |
| **global.publishKnative.resources.limits.memory** | Specifies memory limits set for **publishKnative**. | `32M` |
| **global.subscriptionControllerKnative.resources.limits.memory** | Specifies memory limits set for **subscriptionControllerKnative**. | `32M` |
