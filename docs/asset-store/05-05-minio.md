---
title: MinIO sub-chart
type: Configuration
---

To configure the MinIO sub-chart, override the default values of its `values.yaml` file.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Sub-charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-sub-chart-overrides)

As MinIO is originally an open-source Helm chart, some of its values in Kyma are already overridden in the main `values.yaml` of the Asset Store chart.

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **secretkey** | Defines the secret key. Add the parameter to set your own **secretkey** credentials. | By default, **secretkey** is automatically generated. |
| **accesskey** | Defines the access key. Add the parameter to set your own **accesskey** credentials.  | By default, **accesskey** is automatically generated. |

See the official [MinIO documentation](https://github.com/helm/charts/tree/master/stable/minio#configuration) for a full list of MinIO configurable parameters.
