---
title: CMS Controller Manager sub-chart
type: Configuration
---

To configure the Content Management System (CMS) Controller Manager sub-chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Sub-charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-sub-chart-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **resources.limits.cpu** |  Defines limits for CPU resources. | `100m` |
| **resources.limits.memory** | Defines limits for memory resources. | `30Mi` |
| **resources.requests.cpu** | Defines requests for CPU resources. | `100m` |
| **resources.requests.memory** | Defines requests for memory resources. | `20Mi` |
| **clusterDocsTopic.relistInterval** | Determines time intervals in which the Controller Manager verifies the ClusterDocsTopic for changes. | `5m` |
| **docsTopic.relistInterval** | Determines time intervals in which the Controller Manager verifies the DocsTopic for changes. | `5m` |
| **clusterBucket.region** | Specifies the regional location of the ClusterBucket in a given cloud storage. | `us-east-1` |
| **bucket.region** | Specifies the regional location of the Bucket in a given cloud storage. | `us-east-1` |
