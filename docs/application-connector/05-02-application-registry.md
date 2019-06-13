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
| **deployment.args.requestTimeout** | Specifies an overall time-out for requests sent to the Application Registry. It is provided in seconds. | `10` |
| **deployment.args.requestLogging** | Enables logging incoming requests. By deafult, the logging is disabled. | `false` |
| **deployment.args.specRequestTimeout** | Specifies a period of time after which a request fetching specifications provided by the user fails to be sent. It is provided in seconds. | `5` |
| **deployment.args.assetstoreRequestTimeout** | Specifies a period of time after which a request fetching specifications from the Asset Store fails to be sent. It is provided in seconds. | `5` |
| **deployment.args.insecureAssetDownload** | Disables verifying certificates during downloading data from the Asset Store. | `true` | 
