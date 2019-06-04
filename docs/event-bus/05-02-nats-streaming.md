---
title: NATS Streaming chart
type: Configuration
---

To configure NATS Streaming chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **global.natsStreaming.persistence.maxAge** | Specifies the time for which the given Event is stored in NATS Streaming. | `24h` |
| **global.natsStreaming.persistence.size** | Specifies the size of the persistence volume in NATS Streaming. | `1Gi` |
| **global.natsStreaming.resources.limits.memory** | Specifies the memory limits for NATS Streaming. | `256M` |
| **global.natsStreaming.channel.maxInactivity** | Specifies the time after which the autocleaner removes all backing resources related to a given Event type from the NATS Streaming database if there is no activity for this Event type. | `48h` |
