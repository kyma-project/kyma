---
title: Application Operator sub-chart
type: Configuration
---

To configure the Application Operator (AO) sub-chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents: 
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Sub-charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-sub-chart-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **controller.args.installationTimeout** | Specifies a period of time provided for the Application Gateway, Application Connectivity Validator, and Event Service installation. The Application requires these services to be operational. The value is provided in seconds.| `240` |
| **controller.args.tillerTLSSkipVerify** | Disables TLS verification in Tiller. | `true` 
| **global.enableAPIPackages** | Enables the `gatewayOncePerNamespace` [AO work mode](#architecture-application-connector-components-application-operator). | `false` 