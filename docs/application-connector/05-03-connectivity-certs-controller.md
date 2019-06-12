---
title: Connectivity Certs Controller sub-chart
type: Configuration
---

To configure the Connectivity Certs Controller sub-chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents: 
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Sub-charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-sub-chart-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **deployment.args.minimalConnectionSyncPeriod** | Specifies a minimum period of time between particular attempts to synchronize with the Central Connector Service. It is provided in seconds. | `300` |
