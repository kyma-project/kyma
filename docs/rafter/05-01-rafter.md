---
title: Rafter chart
type: Configuration
---

To configure the Rafter chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values.

>**NOTE:** You can define all **envs** either by providing them as inline values or using the **valueFrom** object. See [this](https://github.com/kyma-project/rafter/tree/master/charts/rafter-controller-manager#change-values-for-envs-parameters) document for reference.

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **controller-manager.minio.persistence.enabled** | Enables MinIO persistence. Deactivate it only if you use [Gateway mode](#tutorials-set-minio-to-gateway-mode). | `true` |
| **controller-manager.minio.environment.MINIO_BROWSER** | Enables browsing MinIO storage. By default, the MinIO browser is turned off for security reasons. You can change the value to `on` to use the browser. If you enable the browser, it is available at `https://storage.{DOMAIN}/minio/`, for example at `https://storage.kyma.local/minio/`. | `"off"` |
| **controller-manager.minio.resources.requests.memory** | Defines requests for memory resources. | `32Mi` |
| **controller-manager.minio.resources.requests.cpu** |  Defines requests for CPU resources. | `10m` |
| **controller-manager.minio.resources.limits.memory** |  Defines limits for memory resources. | `128Mi` |
| **controller-manager.minio.resources.limits.cpu** | Defines limits for CPU resources. | `100m` |
| **controller-manager.deployment.replicas** | Defines the number of service replicas. | `1` |
| **controller-manager.pod.resources.limits.cpu** |  Defines limits for CPU resources. | `150m` |
| **controller-manager.pod.resources.limits.memory** | Defines limits for memory resources. | `128Mi` |
| **controller-manager.pod.resources.requests.cpu** | Defines requests for CPU resources. | `10m` |
| **controller-manager.pod.resources.requests.memory** | Defines requests for memory resources. | `32Mi` |
| **controller-manager.envs.clusterAssetGroup.relistInterval** | Determines time intervals in which the Rafter Controller Manager verifies the ClusterAssetGroup for changes. | `5m` |
| **controller-manager.envs.assetGroup.relistInterval** | Determines time intervals in which the Rafter Controller Manager verifies the AssetGroup for changes. | `5m` |
| **controller-manager.envs.clusterBucket.region** | Specifies the regional location of the ClusterBucket in a given cloud storage. Use one of [these](https://github.com/kyma-project/kyma/blob/master/resources/cluster-essentials/templates/rafter.clusterbuckets.crd.yaml#L52) regions. | `us-east-1` |
| **controller-manager.envs.bucket.region** | Specifies the regional location of the bucket in a given cloud storage. Use one of [these](https://github.com/kyma-project/kyma/blob/master/resources/cluster-essentials/templates/rafter.buckets.crd.yaml#L52) regions. | `us-east-1` |
| **controller-manager.envs.clusterBucket.maxConcurrentReconciles** | Defines the maximum number of cluster bucket concurrent reconciles which will run. | `1` |
| **controller-manager.envs.bucket.maxConcurrentReconciles** | Defines the maximum number of bucket concurrent reconciles which will run. | `1` |
| **controller-manager.envs.clusterAsset.maxConcurrentReconciles** | Defines the maximum number of cluster asset concurrent reconciles which will run. | `1` |
| **controller-manager.envs.asset.maxConcurrentReconciles** | Defines the maximum number of asset concurrent reconciles which will run. | `1` |
| **controller-manager.minio.secretKey** | Defines the secret key. Add the parameter to set your own **secretkey** credentials. | By default, **secretKey** is automatically generated. |
| **controller-manager.minio.accessKey** | Defines the access key. Add the parameter to set your own **accesskey** credentials. | By default, **accessKey** is automatically generated. |
| **controller-manager.envs.store.uploadWorkers** | Defines the number of workers used in parallel to upload files to the storage bucket. | `10` |
| **controller-manager.envs.webhooks.validation.workers** | Defines the number of workers used in parallel to validate files. | `10` |
| **controller-manager.envs.webhooks.mutation.workers** | Defines the number of workers used in parallel to mutate files. | `10` |
| **upload-service.deployment.replicas** | Defines the number of service replicas. | `1` |
| **upload-service.envs.verbose** | If set to `true`, you enable the extended logging mode that records more information on AsyncAPI Service activities than the usual logging mode which registers only errors and warnings. | `true` |
| **front-matter-service.deployment.replicas** | Defines the number of service replicas. For more details, see the [Kubernetes documentation](https://kubernetes.io/docs/concepts/workloads/controllers/replicaset/).| `1` |
| **front-matter-service.envs.verbose** |  If set to `true`, you enable the extended logging mode that records more information on Front Matter Service activities than the usual logging mode which registers only errors and warnings. | `true` |
| **asyncapi-service.deployment.replicas** | Defines the number of service replicas. | `1` |
| **asyncapi-service.envs.verbose** |  If set to `true`, you enable the extended logging mode that records more information on AsyncAPI Service activities than the usual logging mode which registers only errors and warnings. | `true` |
