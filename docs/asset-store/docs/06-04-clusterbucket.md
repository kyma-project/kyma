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
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR which is also the prefix of the bucket name in the bucket storage. |
| **spec.region** |    **NO**   | Specifies the location of the [region](https://github.com/kyma-project/kyma/blob/master/components/assetstore-controller-manager/config/crds/assetstore_v1alpha1_bucket.yaml#L34) under which the Bucket Controller creates the bucket. If the field is empty, the Bucket Controller creates the bucket under the default location. |
| **spec.policy** | **NO** | Specifies te access mode to the bucket. Us one of `none`, `readonly`, `writeonly`, `readwrite` |
| **status.lastheartbeattime** |    **Not applicable**    | Provides the last time when the Bucket Controller processed the Bucket CR. |
| **status.message** |    **Not applicable**    | Describes a human-readable message on the CR processing success or failure. |
| **status.phase** |    **Not applicable**    | The Bucket Controller automatically adds it to the Bucket CR. It describes the status of processing the Bucket CR by the Bucket Controller. It can be `Ready` or `Failed`. |
| **status.reason** |    **Not applicable**    | Provides information on the Bucket CR processing success or failure. |
| **status.url** |    **Not applicable**   | Provides the address of the bucket storage under which the asset is available. |
| **status.remoteName** |    **Not applicable**   | Provides the name of the bucket in storage. |
| **status.observedGeneration** |    **Not applicable**   | The generation observed by the ClusterBucket Controller. |

> **NOTE:** The ClusterBucket Controller automatically adds all parameters marked as **Not applicable** to the Bucket CR.

## Related resources and components

These are the resources related to this CR:

| Custom resource |   Description |
|:----------:|:------|
| ClusterAsset |  Provides the name of the storage bucket which the ClusterAsset CR refers to. |

These components use this CR:

| Component   |   Description |
|:----------:|:------|
| Asset Store |  Uses the Bucket CR for the storage bucket definition. |
