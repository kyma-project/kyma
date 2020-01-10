---
title: ClusterAssetGroup
type: Custom Resource
---

The `clusterassetgroups.rafter.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define an orchestrator that creates ClusterAsset CRs for a specific asset type. To get the up-to-date CRD and show the output in the YAML format, run this command:

```bash
kubectl get crd clusterassetgroups.rafter.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample ClusterAssetGroup custom resource (CR) that provides details of the ClusterAsset CR for the **markdown** source type.

```yaml
apiVersion: rafter.kyma-project.io/v1beta1
kind: ClusterAssetGroup
metadata:
  name: service-mesh
  labels:
    rafter.kyma-project.io/view-context: docs-ui
    rafter.kyma-project.io/group-name: components
    rafter.kyma-project.io/order: "6"
spec:
  displayName: "Service Mesh"
  description: "Overall documentation for Service Mesh"
  bucketRef:
    name: test-bucket
  sources:
    - type: markdown
      name: docs
      mode: package
      parameters:
        disableRelativeLinks: "true"
      url: https://github.com/kyma-project/kyma/archive/master.zip
      filter: /docs/service-mesh/docs/
status:
  lastHeartbeatTime: "2019-03-18T13:42:55Z"
  message: Assets are ready to use
  phase: Ready
  reason: AssetsReady

```

## Custom resource parameters

This table lists all possible parameters of a given resource together with their descriptions:


| Parameter   |      Required      |  Description |
|----------|:-------------:|------|
| **metadata.name** | Yes | Specifies the name of the CR. It also defines the **rafter.kyma-project.io/asset-group** label added to the ClusterAsset CR that the ClusterAssetGroup CR defines. Because of label name limitations, ClusterAssetGroup CR names can have a maximum length of 63 characters. |
| **metadata.labels** | No | Specifies how to filter and group ClusterAsset CRs that the ClusterAssetGroup CR defines. See [this](#details-rafter-in-console) document for more details. |
| **spec.displayname** | Yes | Specifies a human-readable name of the ClusterAssetGroup CR. |
| **spec.description** | Yes | Provides more details on the purpose of the ClusterAssetGroup CR. |
| **spec.bucketRef.name** | No | Specifies the name of the bucket that stores the assets from the ClusterAssetGroup. |
| **spec.sources** | Yes  | Defines the type of the asset and the **rafter.kyma-project.io/type** label added to the ClusterAsset CR.  |
| **spec.sources.type** | Yes | Specifies the type of assets included in the ClusterAssetGroup CR. |
| **spec.sources.name** | Yes | Defines a unique identifier of a given asset. It must be unique if there is more than one asset of a given type in a ClusterAssetGroup CR. |
| **spec.sources.mode** | Yes | Specifies if the asset consists of one file or a set of compressed files in the ZIP or TAR format. Use `single` for one file and `package` for a set of files.  |
| **spec.sources.parameters** | No | Specifies a set of parameters for the ClusterAsset. For example, use it to define what to render, disable, or modify in the UI. Define it in a valid YAML or JSON format. |
| **spec.sources.url** | Yes  | Specifies the location of a single file or a package. |
| **spec.sources.filter** | No | Specifies a set of assets from the package to upload. The regex used in the filter must be [RE2](https://golang.org/s/re2syntax)-compliant. |
| **status.lastheartbeattime** | Not applicable | Provides the last time when the ClusterAssetGroup Controller processed the ClusterAssetGroup CR. |
| **status.message** | Not applicable | Describes a human-readable message on the CR processing progress, success, or failure. |
| **status.phase** | Not applicable | The ClusterAssetGroup Controller adds it to the ClusterAssetGroup CR. It describes the status of processing the ClusterAssetGroup CR by the ClusterAssetGroup Controller. It can be `Ready`, `Pending`, or `Failed`. |
| **status.reason** | Not applicable | Provides the reason why the ClusterAssetGroup CR processing succeeded, is pending, or failed. See the [**Reasons**](#status-reasons) section for the full list of possible status reasons and their descriptions.  |

>**NOTE:** The ClusterAssetGroup Controller automatically adds all parameters marked as **Not applicable** to the ClusterAssetGroup CR.

### Status reasons

Processing of a ClusterAssetGroup CR can succeed, continue, or fail for one of these reasons:

| Reason | Phase | Description |
| --------- | ------------- | ----------- |
| `AssetCreated` | `Pending` | The ClusterAssetGroup Controller created the specified asset. |
| `AssetCreationFailed` | `Failed` | The ClusterAssetGroup Controller couldn't create the specified asset due to an error. |
| `AssetsCreationFailed` | `Failed` | The ClusterAssetGroup Controller couldn't create assets due to an error. |
| `AssetsListingFailed` | `Failed` | The ClusterAssetGroup Controller couldn't list assets due to an error. |
| `AssetDeleted` | `Pending` | The ClusterAssetGroup Controller deleted specified assets. |
| `AssetDeletionFailed` | `Failed`  | The ClusterAssetGroup Controller couldn't delete the specified asset due to an error. |
| `AssetsDeletionFailed` | `Failed` | The ClusterAssetGroup Controller couldn't delete assets due to an error. |
| `AssetUpdated` | `Pending` | The ClusterAssetGroup Controller updated the specified asset. |
| `AssetUpdateFailed` | `Failed` | The ClusterAssetGroup Controller couldn't upload the specified asset due to an error. |
| `AssetsUpdateFailed` | `Failed` | The ClusterAssetGroup Controller couldn't update assets due to an error. |
| `AssetsReady` | `Ready` | Assets are ready to use. |  
| `WaitingForAssets` | `Pending` | Waiting for assets to be in the `Ready` status phase. |
| `BucketError` | `Failed` | Bucket verification failed due to an error. |
| `AssetsWebhookGetFailed` | `Failed` | The ClusterAssetGroup Controller failed to obtain proper webhook configuration. |
| `AssetsSpecValidationFailed` | `Failed` | Asset CR specification is invalid due to an error. |

## Related resources and components

These are the resources related to this CR:

| Custom resource |   Description |
|----------|------|
| ClusterAsset | The ClusterAssetGroup CR orchestrates the creation of the ClusterAsset CR and defines its content. |

These components use this CR:

| Component   |   Description |
|----------|------|
| Rafter |  Manages ClusterAsset CRs created based on the definition in the ClusterAssetGroup CR. |
