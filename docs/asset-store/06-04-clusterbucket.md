---
title: ClusterBucket
type: Custom Resource
---

The `clusterbuckets.assetstore.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define the name of the cloud storage bucket for storing assets. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd clusterbuckets.assetstore.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource that defines the storage bucket configuration.

```
apiVersion: assetstore.kyma-project.io/v1alpha2
kind: ClusterBucket
metadata:
  name: test-sample
spec:
  region: "us-east-1"
  policy: readonly
status:
  lastHeartbeatTime: "2019-02-04T11:50:26Z"
  message: Bucket policy has been updated
  phase: Ready
  reason: BucketPolicyUpdated
  remoteName: test-sample-1b19rnbuc6ir8
  url: https://minio.kyma.local/test-sample-1b19rnbuc6ir8
  observedGeneration: 1
```

## Custom resource parameters

This table lists all possible parameters of a given resource together with their descriptions:


| Parameter   |      Mandatory      |  Description |
|----------|:-------------:|------|
| **metadata.name** |    **YES**   | Specifies the name of the CR which is also the prefix of the bucket name in the bucket storage. |
| **spec.region** |    **NO**   | Specifies the location of the [region](https://github.com/kyma-project/kyma/blob/master/components/asset-store-controller-manager/config/crds/assetstore_v1alpha2_bucket.yaml#L48) under which the ClusterBucket Controller creates the bucket. If the field is empty, the ClusterBucket Controller creates the bucket under the default location. |
| **spec.policy** | **NO** | Specifies the type of bucket access. Use `none`, `readonly`, `writeonly`, or `readwrite`. |
| **status.lastheartbeattime** |    **Not applicable**    | Provides the last time when the ClusterBucket Controller processed the ClusterBucket CR. |
| **status.message** |    **Not applicable**    | Describes a human-readable message on the CR processing success or failure. |
| **status.phase** |    **Not applicable**    | The ClusterBucket Controller automatically adds it to the ClusterBucket CR. It describes the status of processing the ClusterBucket CR by the ClusterBucket Controller. It can be `Ready` or `Failed`. |
| **status.reason** |    **Not applicable**    | Provides information on the ClusterBucket CR processing success or failure. |
| **status.url** |    **Not applicable**   | Provides the address of the bucket storage under which the asset is available. |
| **status.remoteName** |    **Not applicable**   | Provides the name of the bucket in storage. |
| **status.observedGeneration** |    **Not applicable**   | Specifies the most recent generation that the ClusterBucket Controller observes. |

> **NOTE:** The ClusterBucket Controller automatically adds all parameters marked as **Not applicable** to the ClusterBucket CR.

## Related resources and components

These are the resources related to this CR:

| Custom resource |   Description |
|----------|------|
| ClusterAsset |  Provides the name of the storage bucket which the ClusterAsset CR refers to. |

These components use this CR:

| Component   |   Description |
|----------|------|
| Asset Store |  Uses the ClusterBucket CR for the storage bucket definition. |
