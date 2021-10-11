---
title: Application Connector chart
---

To configure the Application Connector (AC) chart, override the default values of its [`values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/application-connector/values.yaml) file. This document describes parameters that you can configure.

>**TIP:** See how to [change Kyma settings](../../04-operation-guides/operations/03-change-kyma-config-values.md).

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **global.disableLegacyConnectivity** | Disables the default standalone (legacy) connectivity components and enables the Compass mode. | `false` |
| **global.isLocalEnv** | Specifies whether the component is run locally or on a cluster. Used in Connector Service and Application Registry. | `false` |