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

>**NOTE:** You can define all **envs** either by providing them as inline values or using the **valueFrom** object. See the [example](https://github.com/kyma-project/rafter/tree/master/charts/rafter-controller-manager#change-values-for-envs-parameters) for reference.

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **controller-manager.minio.persistence.enabled** | Parameter that enables MinIO persistence. Deactivate it only if you use [Gateway mode](#tutorials-set-minio-to-gateway-mode). | `true` |
| **controller-manager.minio.environment.MINIO_BROWSER** | Parameter that enables browsing MinIO storage. By default, the MinIO browser is turned off for security reasons. You can change the value to `on` to use the browser. If you enable the browser, it is available at `https://storage.{DOMAIN}/minio/`, for example at `https://storage.kyma.local/minio/`. | `"off"` |
| **controller-manager.minio.resources.requests.memory** | Requests for memory resources. | `32Mi` |
| **controller-manager.minio.resources.requests.cpu** |  Requests for CPU resources. | `10m` |
| **controller-manager.minio.resources.limits.memory** |  Limits for memory resources. | `128Mi` |
| **controller-manager.minio.resources.limits.cpu** | Limits for CPU resources. | `100m` |
| **controller-manager.deployment.replicas** | Number of service replicas. | `1` |
| **controller-manager.pod.resources.limits.cpu** |  Limits for CPU resources. | `150m` |
| **controller-manager.pod.resources.limits.memory** | Limits for memory resources. | `128Mi` |
| **controller-manager.pod.resources.requests.cpu** | Requests for CPU resources. | `10m` |
| **controller-manager.pod.resources.requests.memory** | Requests for memory resources. | `32Mi` |
| **controller-manager.envs.clusterAssetGroup.relistInterval** | Time intervals in which the Rafter Controller Manager verifies the ClusterAssetGroup for changes. | `5m` |
| **controller-manager.envs.assetGroup.relistInterval** | Time intervals in which the Rafter Controller Manager verifies the AssetGroup for changes. | `5m` |
| **controller-manager.envs.clusterBucket.region** | Regional location of the ClusterBucket in a given cloud storage. Use one of the available [regions](https://github.com/kyma-project/kyma/blob/main/resources/cluster-essentials/files/clusterbuckets.rafter.crd.yaml#L53). | `us-east-1` |
| **controller-manager.envs.bucket.region** | Regional location of the bucket in a given cloud storage. Use one of the available [regions](https://github.com/kyma-project/kyma/blob/main/resources/cluster-essentials/files/buckets.rafter.crd.yaml#L53). | `us-east-1` |
| **controller-manager.envs.clusterBucket.maxConcurrentReconciles** | Maximum number of cluster bucket concurrent reconciles which will run. | `1` |
| **controller-manager.envs.bucket.maxConcurrentReconciles** | Maximum number of bucket concurrent reconciles which will run. | `1` |
| **controller-manager.envs.clusterAsset.maxConcurrentReconciles** | Maximum number of cluster asset concurrent reconciles which will run. | `1` |
| **controller-manager.envs.asset.maxConcurrentReconciles** | Maximum number of asset concurrent reconciles which will run. | `1` |
| **controller-manager.minio.secretKey** | Secret key. Add the parameter to set your own **secretkey** credentials. | By default, **secretKey** is automatically generated. |
| **controller-manager.minio.accessKey** | Access key. Add the parameter to set your own **accesskey** credentials. | By default, **accessKey** is automatically generated. |
| **controller-manager.envs.store.uploadWorkers** | Number of workers used in parallel to upload files to the storage bucket. | `10` |
| **controller-manager.envs.webhooks.validation.workers** | Number of workers used in parallel to validate files. | `10` |
| **controller-manager.envs.webhooks.mutation.workers** | Number of workers used in parallel to mutate files. | `10` |
| **upload-service.deployment.replicas** | Number of service replicas. | `1` |
| **upload-service.envs.verbose** | If set to `true`, you enable the extended logging mode that records more information on AsyncAPI Service activities than the usual logging mode which registers only errors and warnings. | `true` |
| **front-matter-service.deployment.replicas** | Number of service replicas. For more details, see the [Kubernetes documentation](https://kubernetes.io/docs/concepts/workloads/controllers/replicaset/).| `1` |
| **front-matter-service.envs.verbose** |  If set to `true`, you enable the extended logging mode that records more information on Front Matter Service activities than the usual logging mode which registers only errors and warnings. | `true` |
| **asyncapi-service.deployment.replicas** | Number of service replicas. | `1` |
| **asyncapi-service.envs.verbose** |  If set to `true`, you enable the extended logging mode that records more information on AsyncAPI Service activities than the usual logging mode which registers only errors and warnings. | `true` |
