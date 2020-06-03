---
title: ClusterAsset
type: Custom Resource
---

The `clusterassets.rafter.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define an asset to store in a cloud storage bucket. To get the up-to-date CRD and show the output in the YAML format, run this command:

```bash
kubectl get crd clusterassets.rafter.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample ClusterAsset CR configuration that contains mutation, validation, and metadata services:

```yaml
apiVersion: rafter.kyma-project.io/v1beta1
kind: ClusterAsset
metadata:
  name: my-package-assets
spec:
  source:
    mode: single
    parameters:
      disableRelativeLinks: "true"
    url: https://some.domain.com/main.js
    mutationWebhookService:
    - name: swagger-operations-svc
      namespace: default
      endpoint: "/mutate"
      filter: \.js$
      parameters:
        rewrite: keyvalue
        pattern: \json|yaml
        data:
          basePath: /test/v2
    validationWebhookService:
    - name: swagger-operations-svc
      namespace: default
      endpoint: "/validate"
      filter: \.js$
    metadataWebhookService:
    - name: swagger-operations-svc
      namespace: default
      endpoint: "/extract"
      filter: \.js$
  bucketRef:
    name: my-bucket
  displayName: "Operations svc"
status:
  phase: Ready
  reason: Uploaded
  message: Asset content has been uploaded
  lastHeartbeatTime: "2018-01-03T07:38:24Z"
  observedGeneration: 1
  assetRef:
    baseUrl: https://{STORAGE_ADDRESS}/my-bucket-1b19rnbuc6ir8/my-package-assets
    files:
    - metadata:
        title: Overview
      name: README.md
    - metadata:
        title: Benefits of distributed storage
        type: Details
      name: directory/subdirectory/file.md
```

## Custom resource parameters

This table lists all possible parameters of a given resource together with their descriptions:


| Parameter   |      Required      |  Description |
|----------|:-------------:|------|
| **metadata.name** | Yes | Specifies the name of the CR. |
| **spec.source.mode** | Yes | Specifies if the asset consists of one file or a set of compressed files in the ZIP or TAR formats. Use `single` for one file and `package` for a set of files. |
| **spec.source.parameters** | No | Specifies a set of parameters for the ClusterAsset. For example, use it to define what to render, disable, or modify in the UI. Define it in a valid YAML or JSON format. |
| **spec.source.url** | Yes | Specifies the location of the file. |
| **spec.source.filter** | No | Specifies the regex pattern used to select files to store from the package. |
| **spec.source.validationWebhookService** | No | Provides specification of the validation webhook services. |
| **spec.source.validationWebhookService.name** | Yes | Provides the name of the validation webhook service. |
| **spec.source.validationWebhookService.namespace** | Yes | Provides the Namespace in which the service is available. |
| **spec.source.validationWebhookService.endpoint** | No | Specifies the endpoint to which the service sends calls. |
| **spec.source.validationWebhookService.parameters** | No | Provides detailed parameters specific for a given validation service and its functionality. |
| **spec.source.validationWebhookService.filter** | No | Specifies the regex pattern used to select files sent to the service. |
| **spec.source.mutationWebhookService** | No  | Provides specification of the mutation webhook services. |
| **spec.source.mutationWebhookService.name** | Yes | Provides the name of the mutation webhook service. |
| **spec.source.mutationWebhookService.namespace** | Yes | Provides the Namespace in which the service is available. |
| **spec.source.mutationWebhookService.endpoint** | No | Specifies the endpoint to which the service sends calls. |
| **spec.source.mutationWebhookService.parameters** | No | Provides detailed parameters specific for a given mutation service and its functionality. |
| **spec.source.mutationWebhookService.filter** | No | Specifies the regex pattern used to select files sent to the service. |
| **spec.source.metadataWebhookService** | No | Provides specification of the metadata webhook services. |
| **spec.source.metadataWebhookService.name** | Yes | Provides the name of the metadata webhook service. |
| **spec.source.metadataWebhookService.namespace** | Yes  | Provides the Namespace in which the service is available. |
| **spec.source.metadataWebhookService.endpoint** | No | Specifies the endpoint to which the service sends calls. |
| **spec.source.metadataWebhookService.filter** | No | Specifies the regex pattern used to select files sent to the service. |
| **spec.bucketRef.name** | Yes | Provides the name of the bucket for storing the asset. |
| **spec.displayName** | No | Specifies a human-readable name of the asset. |
| **status.phase** | Not applicable | The ClusterAsset Controller adds it to the ClusterAsset CR. It describes the status of processing the ClusterAsset CR by the ClusterAsset Controller. It can be `Ready`, `Failed`, or `Pending`. |
| **status.reason** | Not applicable | Provides the reason why the ClusterAsset CR processing failed or is pending. See the [**Reasons**](#status-reasons) section for the full list of possible status reasons and their descriptions.  |
| **status.message** | Not applicable | Describes a human-readable message on the CR processing progress, success, or failure. |
| **status.lastHeartbeatTime** | Not applicable | Specifies when was the last time when the ClusterAsset Controller processed the ClusterAsset CR. |
| **status.observedGeneration** | Not applicable | Specifies the most recent ClusterAsset CR generation that the ClusterAsset Controller observed. |
| **status.assetRef** | Not applicable | Provides details on the location of the assets stored in the bucket.   |
| **status.assetRef.files** | Not applicable | Provides asset metadata and the relative path to the given asset in the storage bucket with metadata. |
| **status.assetRef.files.metadata** | Not applicable | Lists metadata extracted from the asset. |
| **status.assetRef.files.name** | Not applicable | Specifies the relative path to the given asset in the storage bucket. |
| **status.assetRef.baseUrl** | Not applicable | Specifies the absolute path to the location of the assets in the storage bucket.   |


> **NOTE:** The ClusterAsset Controller automatically adds all parameters marked as **Not applicable** to the ClusterAsset CR.

### Status reasons

Processing of a ClusterAsset CR can succeed, continue, or fail for one of these reasons:

| Reason | Phase | Description |
| --------- | ------------- | ----------- |
| `Pulled` | `Pending` | The ClusterAsset Controller pulled the asset content for processing. |
| `PullingFailed` | `Failed` | Asset content pulling failed due to an error. |
| `Uploaded` | `Ready` | The ClusterAsset Controller uploaded the asset content to MinIO. |
| `UploadFailed` | `Failed` | Asset content uploading failed due to an error. |
| `BucketNotReady` | `Pending` | The referenced bucket is not ready. |
| `BucketError` | `Failed` | Reading the bucket status failed due to an error. |
| `Mutated` | `Pending` | Mutation services changed the asset content. |
| `MutationFailed` | `Failed` | Asset mutation failed for one of the provided reasons. |
| `MutationError` | `Failed` | Asset mutation failed due to an error. |
| `MetadataExtracted` | `Pending` | Metadata services extracted metadata from the asset content. |
| `MetadataExtractionFailed` | `Failed` | Metadata extraction failed due to an error. |
| `Validated` | `Pending` | Validation services validated the asset content. |
| `ValidationFailed` | `Failed` | Asset validation failed for one of the provided reasons. |
| `ValidationError` | `Failed` | Asset validation failed due to an error. |
| `MissingContent` | `Failed` | There is missing asset content in the cloud storage bucket. |
| `RemoteContentVerificationError` | `Failed` | Asset content verification in the cloud storage bucket failed due to an error. |
| `CleanupError` | `Failed` | The ClusterAsset Controller failed to remove the old asset content due to an error. |
| `Cleaned` | `Pending` | The ClusterAsset Controller removed the old asset content that was modified. |
| `Scheduled` | `Pending` | The asset you added is scheduled for processing. |

## Related resources and components

These are the resources related to this CR:

| Custom resource |   Description |
|----------|------|
| ClusterBucket |  The ClusterAsset CR uses the name of the bucket specified in the definition of the ClusterBucket CR. |

These components use this CR:

| Component   |   Description |
|----------|------|
| Rafter |  Uses the ClusterAsset CR for the detailed asset definition, including its location and the name of the bucket in which it is stored. |
