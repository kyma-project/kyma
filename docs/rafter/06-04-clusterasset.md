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
status:
  phase: Failed
  reason: ValidationFailed
  message: "The file is not valid against the provided json schema"
  lastHeartbeatTime: "2018-01-03T07:38:24Z"
  observedGeneration: 1
  assetRef:
    assets:
    - README.md
    - directory/subdirectory/file.md
    baseUrl: https://minio.kyma.local/test-sample-1b19rnbuc6ir8/asset-sample

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
| **spec.source.validationwebhookservice** | No | Provides specification of the validation webhook services. |
| **spec.source.validationwebhookservice.name** | Yes | Provides the name of the validation webhook service. |
| **spec.source.validationwebhookservice.namespace** | Yes | Provides the Namespace in which the service is available. |
| **spec.source.validationwebhookservice.endpoint** | No | Specifies the endpoint to which the service sends calls. |
| **spec.source.validationwebhookservice.parameters** | No | Provides detailed parameters specific for a given validation service and its functionality. |
| **spec.source.validationwebhookservice.filter** | No | Specifies the regex pattern used to select files sent to the service. |
| **spec.source.mutationwebhookservice** | No  | Provides specification of the mutation webhook services. |
| **spec.source.mutationwebhookservice.name** | Yes | Provides the name of the mutation webhook service. |
| **spec.source.mutationwebhookservice.namespace** | Yes | Provides the Namespace in which the service is available. |
| **spec.source.mutationwebhookservice.endpoint** | No | Specifies the endpoint to which the service sends calls. |
| **spec.source.mutationwebhookservice.parameters** | No | Provides detailed parameters specific for a given mutation service and its functionality. |
| **spec.source.mutationwebhookservice.filter** | No | Specifies the regex pattern used to select files sent to the service. |
| **spec.source.metadatawebhookservice** | No | Provides specification of the metadata webhook services. |
| **spec.source.metadatawebhookservice.name** | Yes | Provides the name of the metadata webhook service. |
| **spec.source.metadatawebhookservice.namespace** | Yes  | Provides the Namespace in which the service is available. |
| **spec.source.metadatawebhookservice.endpoint** | No | Specifies the endpoint to which the service sends calls. |
| **spec.source.metadatawebhookservice.filter** | No | Specifies the regex pattern used to select files sent to the service. |
| **spec.bucketref.name** | Yes | Provides the name of the bucket for storing the asset. |
| **status.phase** | Not applicable | The ClusterAsset Controller adds it to the ClusterAsset CR. It describes the status of processing the ClusterAsset CR by the ClusterAsset Controller. It can be `Ready`, `Failed`, or `Pending`. |
| **status.reason** | Not applicable | Provides the reason why the ClusterAsset CR processing failed or is pending. See the [**Reasons**](#status-reasons) section for the full list of possible status reasons and their descriptions.  |
| **status.message** | Not applicable | Describes a human-readable message on the CR processing progress, success, or failure. |
| **status.lastheartbeattime** | Not applicable | Provides the last time when the ClusterAsset Controller processed the ClusterAsset CR. |
| **status.observedGeneration** | Not applicable | Specifies the most recent generation that the ClusterAsset Controller observes. |
| **status.assetref** | Not applicable | Provides details on the location of the assets stored in the bucket.   |
| **status.assetref.assets** | Not applicable | Provides the relative path to the given asset in the storage bucket. |
| **status.assetref.baseurl** | Not applicable | Specifies the absolute path to the location of the assets in the storage bucket.   |


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
