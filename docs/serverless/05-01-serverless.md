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

>**CAUTION:** Currently, there is no automated validation for CPU and memory requests and limits. If you update their default values, make sure that the values for CPU and memory requests are not lower than the default ones, and that values for requests are lower than or equal to those for the limits.

>**NOTE:** You can define all **envs** either by providing them as inline values or using the **valueFrom** object. See [this](https://github.com/kyma-project/rafter/tree/master/charts/rafter-controller-manager#change-values-for-envs-parameters) document for reference.

| Parameter | Description | Default value | Minikube override |
|-----------|-------------|---------------|
| **containers.manager.envs.buildRequestsCPU** | Specifies the number of CPUs requested by the image building Pod to operate. | `700m` | `100m`|
| **containers.manager.envs.buildRequestsMemory** | Specifies the amount of memory requested by the image building Pod to operate.  | `700Mi` | `200Mi` |
| **containers.manager.envs.buildLimitsCPU** | Defines the maximum number of CPUs available for the image building Pod to use. | `1100m` | `200m` |
| **containers.manager.envs.buildLimitsMemory** | Defines the maximum amount of memory available for the image building Pod to use. | `1100Mi` | `400Mi` |
