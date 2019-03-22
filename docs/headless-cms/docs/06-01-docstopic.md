---
title: DocsTopic
type: Custom Resource
---

The `docstopics.cms.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define an orchestrator that creates Asset CRs for a specific asset type. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd docstopics.cms.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample DocsTopic custom resource (CR) that provides details of the Asset CRs for the **redis**, **asyncapi**, **markdown**, and **openapi** source types.

```
apiVersion: cms.kyma-project.io/v1alpha1
kind: DocsTopic
metadata:
  name: service-catalog
  namespace: default
  labels:
    viewContext: docs-view
    groupName: components
spec:
  displayName: "Service Catalog"
  description: "Service Catalog documentation"
  sources:
    redis:
      mode: package
      url: https://github.com/kyma-project/bundles/releases/download/latest/redis-0.0.3.tgz
    asyncapi:
      mode: single
      url: https://raw.githubusercontent.com/asyncapi/asyncapi/master/examples/1.2.0/slack-rtm.yml
    markdown:
      mode: package
      url: https://github.com/kyma-project/kyma/archive/master.zip
      filter: ^kyma-master/docs/
    openapi:
      mode: single
      url: https://petstore.swagger.io/v2/swagger.json
status:
  lastHeartbeatTime: "2019-03-18T13:42:55Z"
  message: Assets are ready to use
  phase: Ready
  reason: AssetsReady
```

## Custom resource parameters

This table lists all possible parameters of a given resource together with their descriptions:


| Parameter   |      Mandatory      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR. It also defines the respective **docsTopic** label that is added to the Asset CR that the DocsTopic CR defines. |
| **metadata.namespace** |    **YES**   | Defines the Namespace in which the CR is available. |
| **metadata.labels** |    **YES**   | Specifies how to filter and group Asset CRs that the DocsTopic CR defines. The available labels include **viewContext** and **groupName**. |
| **metadata.labels.viewcontext** |    **YES**   | Specifies the location in the Console UI where you want to render the given asset. |
| **metadata.labels.groupname** |    **YES**   | Specifies the group under which you want to render the given asset in the Console UI. The value cannot include spaces. |
| **spec.displayname** |    **YES**   | Specifies a human-readable name of the DocsTopic CR. |
| **spec.description** |    **YES**   | Provides more details on the purpose of the DocsTopic CR. |
| **spec.sources** |    **YES**   | Defines the type of the asset and a **type** label added to the Asset CR.  |
| **spec.sources.markdown** |    **YES**   | Specifies the `markdown` type of the assets included in the DocsTopic CR. |
| **spec.sources.markdown.mode** |    **YES**   | Specifies if the asset consists of one file or a set of compressed files in the ZIP or TAR formats. Use `single` for one file and `package` for a set of files.  |
| **spec.sources.markdown.url** |    **YES**   | Specifies the location of a single file or a package. |
| **spec.sources.markdown.filter** |    **NO**   | Specifies the regex pattern used to select files to store from the package. |
| **status.lastheartbeattime** |    **Not applicable**   | Provides the last time when the DocsTopic Controller processed the DocsTopic CR. |
| **status.message** |    **Not applicable**   | Describes a human-readable message on the CR processing progress, success, or failure. |
| **status.phase** |    **Not applicable**   | The DocsTopic Controller adds it to the DocsTopic CR. It describes the status of processing the DocsTopic CR by the DocsTopic Controller. It can be `Ready`, `Pending`, or `Failed`. |
| **status.reason** |    **Not applicable**   | Provides the reason why the DocsTopic CR processing succeeded, is pending, or failed.  |

> **NOTE:** The DocsTopic Controller automatically adds all parameters marked as **Not applicable** to the DocsTopic CR.

## Related resources and components

These are the resources related to this CR:

| Custom resource |   Description |
|:----------:|:------|
| Asset |  The DocsTopic CR orchestrates the creation of the Asset CR and defines its content. |

These components use this CR:

| Component   |   Description |
|:----------:|:------|
| Asset Store | Manages Asset CRs created based on the definition in the DocsTopic CRs. |
