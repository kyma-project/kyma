---
title: Serverless chart
type: Configuration
---

To configure the Serverless chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values for both cluster and local installations.

| Parameter | Description | Default value | Minikube override |
|-----------|-------------|---------------|---------------|
| **containers.manager.envs.buildRequestsCPU** | Number of CPUs requested by the image-building Pod to operate. | `700m` | `100m`|
| **containers.manager.envs.buildRequestsMemory** | Amount of memory requested by the image-building Pod to operate.  | `700Mi` | `200Mi` |
| **containers.manager.envs.buildLimitsCPU** | Maximum number of CPUs available for the image-building Pod to use. | `1100m` | `200m` |
| **containers.manager.envs.buildLimitsMemory** | Maximum amount of memory available for the image-building Pod to use. | `1100Mi` | `400Mi` |
