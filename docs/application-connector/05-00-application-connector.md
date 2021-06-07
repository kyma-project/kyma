---
title: Application Connector chart
type: Configuration
---

To configure the Application Connector (AC) chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **central_connectivity_validator.enabled** | Specifies whether to use Central Connectivity Validator for Application Operator. If enabled, it removes the existing per-Application Gateways and installs the Central Connectivity Validator chart. | `false` |
| **global.disableLegacyConnectivity** | Disables the default standalone (legacy) connectivity components and enables the Compass mode. | `false` |
| **global.isLocalEnv** | Specifies whether the component is run locally or on a cluster. Used in Connector Service and Application Registry. | `false` |
| **global.log.format** | Specifies the logging format. Used in Application Operator. | `json` |
| **global.log.level** | Specifies the logging level. Used in Application Operator. | `warn` |