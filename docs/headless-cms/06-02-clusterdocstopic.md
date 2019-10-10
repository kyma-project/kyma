---
title: ClusterDocsTopic
type: Custom Resource
---

The `clusterdocstopics.cms.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define an orchestrator that creates ClusterAsset CRs for a specific asset type. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```bash
kubectl get crd clusterdocstopics.cms.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample ClusterDocsTopic custom resource (CR) that provides details of the ClusterAsset CR for the **markdown** source type.

```yaml
apiVersion: cms.kyma-project.io/v1alpha1
kind: ClusterDocsTopic
metadata:
  name: service-mesh
  labels:
    cms.kyma-project.io/view-context: docs-ui
    cms.kyma-project.io/group-name: components
    cms.kyma-project.io/order: "6"
spec:
  displayName: "Service Mesh"
  description: "Overall documentation for Service Mesh"
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
| **metadata.name** | Yes | Specifies the name of the CR. It also defines the **cms.kyma-project.io/docs-topic** label added to the ClusterAsset CR that the ClusterDocsTopic CR defines. Because of label name limitations, ClusterDocsTopic CR names can have a maximum length of 63 characters. |
| **metadata.labels** | No | Specifies how to filter and group ClusterAsset CRs that the ClusterDocsTopic CR defines. See [this](#details-headless-cms-in-console) document for more details. |
| **spec.displayname** | Yes | Specifies a human-readable name of the ClusterDocsTopic CR. |
| **spec.description** | Yes | Provides more details on the purpose of the ClusterDocsTopic CR. |
| **spec.sources** | Yes  | Defines the type of the asset and the **cms.kyma-project.io/type** label added to the ClusterAsset CR.  |
| **spec.sources.type** | Yes | Specifies the type of assets included in the ClusterDocsTopic CR. |
| **spec.sources.name** | Yes | Defines a unique identifier of a given asset. It must be unique if there is more than one asset of a given type in a ClusterDocsTopic CR. |
| **spec.sources.mode** | Yes | Specifies if the asset consists of one file or a set of compressed files in the ZIP or TAR format. Use `single` for one file and `package` for a set of files.  |
| **spec.sources.parameters** | No | Specifies a set of parameters for the ClusterAsset. For example, use it to define what to render, disable, or modify in the UI. Define it in a valid YAML or JSON format. |
| **spec.sources.url** | Yes  | Specifies the location of a single file or a package. |
| **spec.sources.filter** | No | Specifies a set of assets from the package to upload. The regex used in the filter must be [RE2](https://golang.org/s/re2syntax)-compliant. |
| **status.lastheartbeattime** | Not applicable | Provides the last time when the DocsTopic Controller processed the ClusterDocsTopic CR. |
| **status.message** | Not applicable | Describes a human-readable message on the CR processing progress, success, or failure. |
| **status.phase** | Not applicable | The DocsTopic Controller adds it to the ClusterDocsTopic CR. It describes the status of processing the ClusterDocsTopic CR by the DocsTopic Controller. It can be `Ready`, `Pending`, or `Failed`. |
| **status.reason** | Not applicable | Provides the reason why the ClusterDocsTopic CR processing succeeded, is pending, or failed. See the [**Reasons**](#status-reasons) section for the full list of possible status reasons and their descriptions.  |

> **NOTE:** The DocsTopic Controller automatically adds all parameters marked as **Not applicable** to the ClusterDocsTopic CR.

### Status reasons

Processing of a ClusterDocsTopic CR can succeed, continue, or fail for one of these reasons:

| Reason | Phase | Description |
| --------- | ------------- | ----------- |
| `AssetCreated` | `Pending` | The DocsTopic Controller created the specified asset. |
| `AssetCreationFailed` | `Failed` | The DocsTopic Controller couldn't create the specified asset due to the provided error. |
| `AssetsCreationFailed` | `Failed` | The DocsTopic Controller couldn't create assets due to the provided error. |
| `AssetsListingFailed` | `Failed` | The DocsTopic Controller couldn't list assets due to the provided error. |
| `AssetDeleted` | `Pending` | The DocsTopic Controller deleted specified assets. |
| `AssetDeletionFailed` | `Failed`  | The DocsTopic Controller couldn't delete the specified asset due to the provided error. |
| `AssetsDeletionFailed` | `Failed` | The DocsTopic Controller couldn't delete assets due to the provided error. |
| `AssetUpdated` | `Pending` | The DocsTopic Controller updated the specified asset. |
| `AssetUpdateFailed` | `Failed` | The DocsTopic Controller couldn't upload the specified asset due to the provided error. |
| `AssetsUpdateFailed` | `Failed` | The DocsTopic Controller couldn't update assets due to the provided error. |
| `AssetsReady` | `Ready` | Assets are ready to use. |  
| `WaitingForAssets` | `Pending` | Waiting for assets to be in the `Ready` status phase. |
| `BucketError` | `Failed` | Bucket verification failed due to the provided error. |
| `AssetsWebhookGetFailed` | `Failed` | The DocsTopic Controller failed to obtain proper webhook configuration. |
| `AssetsSpecValidationFailed` | `Failed` | Asset specification is invalid due to the provided error. |

## Related resources and components

These are the resources related to this CR:

| Custom resource |   Description |
|----------|------|
| ClusterAsset | The ClusterDocsTopic CR orchestrates the creation of the ClusterAsset CR and defines its content. |

These components use this CR:

| Component   |   Description |
|----------|------|
| Asset Store |  Manages ClusterAsset CRs created based on the definition in the ClusterDocsTopic CR. |
