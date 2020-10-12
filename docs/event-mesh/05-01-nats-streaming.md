---
title: NATS Streaming chart
type: Configuration
---

To configure NATS Streaming chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)
>
> To learn how to configure a Kafka channel instead of the default NATS one, see this tutorial:
>* [Configure the Kafka Channel](#tutorials-configure-the-kafka-channel).

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **global.natsStreaming.persistence.enabled** | Enables / disables saving events to a persistent volume | `true` |
| **global.natsStreaming.persistence.maxAge** | Specifies the time for which the given Event is stored in NATS Streaming. | `24h` |
| **global.natsStreaming.persistence.size** | Specifies the size of the persistent volume in NATS Streaming. | `1Gi` |
| **global.natsStreaming.resources.limits.memory** | Specifies the memory limits for NATS Streaming. | `256M` |
| **global.natsStreaming.channel.maxInactivity** | Specifies the time after which the autocleaner removes all backing resources related to a given Event type from the NATS Streaming database if there is no activity for this Event type. | `48h` |

>**CAUTION:** If persistence is disabled, `nats` will store undelivered messages in memory. All restarts of `nats` will lead to the loss of undelivered messages. Do **not** use this in a production setup.
