---
title: Application Registry sub-chart
type: Configuration
---

To configure the Application Registry sub-chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents: 
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Sub-charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-sub-chart-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **deployment.args.requestTimeout** | Specifies the period of time in seconds after which a request to the Application Registry fails to be sent. | `10` |
| **deployment.args.requestLogging** | Enables logging incominng requests. By deafult the logging is disabled. | `false` |
| **deployment.args.specRequestTimeout** | Specifies the period of time in seconds after a request fetching specifications provided by the user fails to be sent. | `5` |
| **deployment.args.assetstoreRequestTimeout** | Specifies the period of time in seconds after a request fetching specifications from the Asset Store fails to be sent. | `5` |
| **deployment.args.insecureAssetDownload** | Disables verifying certificates during downloading data from the Asset Store. | `true` | 
