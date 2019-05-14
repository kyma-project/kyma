---
title: DocsTopic
type: Custom Resource
---

The `docstopics.cms.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define an orchestrator that creates Asset CRs for a specific asset type. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd docstopics.cms.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample DocsTopic custom resource (CR) that provides details of the Asset CRs for the **markdown**, **asyncapi**, and **openapi** source types.

```
apiVersion: cms.kyma-project.io/v1alpha1
kind: DocsTopic
metadata:
  name: slack
  labels:
    cms.kyma-project.io/view-context: service-catalog
spec:
  displayName: Slack
  description: "Slack documentation"
  sources:
    - type: markdown
      name: markdown-slack
      mode: single
      url: https://raw.githubusercontent.com/slackapi/slack-api-specs/master/README.md
    - type: asyncapi
      name: asyncapi-slack
      mode: single
      url: https://raw.githubusercontent.com/slackapi/slack-api-specs/master/events-api/slack_events_api_async_v1.json
    - type: openapi
      name: openapi-slack
      mode: single
      url: https://raw.githubusercontent.com/slackapi/slack-api-specs/master/web-api/slack_web_openapi_v2.json
status:
  lastHeartbeatTime: "2019-03-18T13:42:55Z"
  message: Assets are ready to use
  phase: Ready
  reason: AssetsReady
```

## Custom resource parameters

This table lists all possible parameters of a given resource together with their descriptions:


| Parameter   |      Mandatory      |  Description |
|----------|:-------------:|------|
| **metadata.name** |    **YES**   | Specifies the name of the CR. It also defines the **cms.kyma-project.io/docs-topic** label added to the Asset CR that the DocsTopic CR defines. Because of label name limitations, DocsTopic CR names can have a maximum length of 63 characters. |
| **metadata.labels** |    **NO**   | Specifies how to filter and group Asset CRs that the DocsTopic CR defines. See [this](#details-headless-cms-in-console) document for more details. |
| **spec.displayname** |    **YES**   | Specifies a human-readable name of the DocsTopic CR. |
| **spec.description** |    **YES**   | Provides more details on the purpose of the DocsTopic CR. |
| **spec.sources** |    **YES**   | Defines the type of the asset and the **cms.kyma-project.io/type** label added to the Asset CR.  |
| **spec.sources.type** |    **YES**   | Specifies the type of assets included in the DocsTopic CR. |
| **spec.sources.name** |    **YES**   | Defines an identifier of a given asset. It must be unique if there is more than one asset of a given type in a DocsTopic CR. |
| **spec.sources.mode** |    **YES**   | Specifies if the asset consists of one file or a set of compressed files in the ZIP or TAR format. Use `single` for one file and `package` for a set of files.  |
| **spec.sources.url** |    **YES**   | Specifies the location of a single file or a package. |
| **spec.sources.filter** |    **NO**   | Specifies a set of assets from the package to upload. The regex used in the filter must be [RE2](https://golang.org/s/re2syntax)-compliant. |
| **status.lastheartbeattime** |    **Not applicable**   | Provides the last time when the DocsTopic Controller processed the DocsTopic CR. |
| **status.message** |    **Not applicable**   | Describes a human-readable message on the CR processing progress, success, or failure. |
| **status.phase** |    **Not applicable**   | The DocsTopic Controller adds it to the DocsTopic CR. It describes the status of processing the DocsTopic CR by the DocsTopic Controller. It can be `Ready`, `Pending`, or `Failed`. |
| **status.reason** |    **Not applicable**   | Provides the reason why the DocsTopic CR processing succeeded, is pending, or failed.  |

> **NOTE:** The DocsTopic Controller automatically adds all parameters marked as **Not applicable** to the DocsTopic CR.

## Related resources and components

These are the resources related to this CR:

| Custom resource |   Description |
|----------|------|
| Asset |  The DocsTopic CR orchestrates the creation of the Asset CR and defines its content. |

These components use this CR:

| Component   |   Description |
|----------|------|
| Asset Store | Manages Asset CRs created based on the definition in the DocsTopic CR. |
