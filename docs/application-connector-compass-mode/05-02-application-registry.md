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
| **deployment.args.requestTimeout** | Specifies an overall time-out after which requests to the Application Registry fail to be sent. It is provided in seconds. | `10` |
| **deployment.args.requestLogging** | Enables logging incoming requests. By default, logging is disabled. | `false` |
| **deployment.args.specRequestTimeout** | Specifies a time-out after which a request fetching specifications provided by the user fails to be sent. It is provided in seconds. | `20` |
| **deployment.args.rafterRequestTimeout** | Specifies a time-out after which a request fetching specifications from Rafter fails to be sent. It is provided in seconds. | `20` |
| **deployment.args.insecureAssetDownload** | Disables verifying certificates when downloading data from Rafter. | `true` | 
| **deployment.args.insecureSpecDownload** | Disables verifying certificates when fetching API specification from specification URL. | `true` |
| **deployment.args.detailedErrorResponse** | Enables showing full messages for internal error responses. | `false` |
