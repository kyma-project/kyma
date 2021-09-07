---
title: Application Operator sub-chart
---

To configure the Application Operator (AO) sub-chart, override the default values of its [`values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/application-connector/charts/application-operator/values.yaml) file. This document describes parameters that you can configure.

>**TIP:** See how to [change Kyma settings](../../04-operation-guides/operations/03-change-kyma-config-values.md).

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **controller.args.installationTimeout** | Specifies a period of time provided for Application Gateway, Application Connectivity Validator, and Event Publisher installation. The Application requires these services to be operational. The value is provided in seconds. | `240` |
| **controller.args.helmDriver** | Specifies the backend storage driver used by Helm 3 to store release data. Possible values are `configmap`, `secret` and `memory`. | `secret` |
| **global.disableLegacyConnectivity** | Disables the default legacy [AO work mode](../03-architecture/ac-01-application-connector-components.md) and enables the Compass mode. | `false` |