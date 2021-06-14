---
title: Serverless chart
---

To configure the Serverless chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>
> - [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
> - [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values for both cluster and local installations.

>**NOTE:** Limited memory and CPU resources on Minikube directly affect the Serverless functionality as you can process only a limited number of Functions at the same time. Also, their processing time is significantly longer. To process large workloads, we recommend using the cluster setup.

| Parameter                                       | Description                                                           | Default value | Minikube override |
| ----------------------------------------------- | --------------------------------------------------------------------- | ------------- | ----------------- |
| **webhook.values.buildJob.resources.minRequestCpu**    | Minimum number of CPUs requested by the image-building Pod to operate.        | `200m`        | `200m`   |
| **webhook.values.buildJob.resources.minRequestMemory** | Minimum amount of memory requested by the image-building Pod to operate.      | `200Mi`       | `200Mi`  |
| **webhook.values.buildJob.resources.defaultPreset**    | Default preset for image-building Pod's resources.      | `normal`        | `local-dev`   |
| **webhook.values.function.replicas.minValue**      | Minimum number of replicas of a single Function.   | `1`       | `1`            |
| **webhook.values.function.replicas.defaultPreset**      | Default preset for Function's replicas.   | `S`       | `S`            |
| **webhook.values.function.resources.minRequestCpu**      | Maximum number of CPUs available for the image-building Pod to use.   | `10m`       | `10m`            |
| **webhook.values.function.resources.minRequestMemory**   | Maximum amount of memory available for the image-building Pod to use. | `16Mi`      | `16Mi`           |
| **webhook.values.function.resources.defaultPreset**      | Default preset for Function's resources.   | `M`       | `M`            |
| **webhook.values.deployment.resources.requests.cpu**      | Value defining CPU requests for a Function's Deployment.   | `30m`       | `30m`            |
| **webhook.values.deployment.resources.requests.memory**      | Value defining memory requests for a Function's Deployment.   | `50Mi`       | `50Mi`            |
| **webhook.values.deployment.resources.limits.cpu**      | Value defining CPU limits for a Function's Deployment.   | `300m`       | `300m`            |
| **webhook.values.deployment.resources.limits.memory**      | Value defining memory limits for a Function's Deployment.   | `300Mi`       | `300Mi`            |
| **containers.manager.envs.functionBuildMaxSimultaneousJobs.value**      | Maximum number of build jobs running simultaneously.   | ` "5"`       | ` "5"`            |

>**TIP:** To learn more, read the official documentation on [resource units in Kubernetes](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-units-in-kubernetes).
