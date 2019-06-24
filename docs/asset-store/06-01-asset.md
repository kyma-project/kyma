---
title: Asset
type: Custom Resource
---

The `assets.assetstore.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define an asset to store in a cloud storage bucket. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd assets.assetstore.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample Asset CR configuration:

```
apiVersion: assetstore.kyma-project.io/v1alpha2
kind: Asset
metadata:
  name: my-package-assets
  namespace: default
spec:
  source:
    mode: package
    url: https://some.domain.com/structure.zip
    filter: .*\.md$
  bucketRef:
    name: my-bucket

```

### Validation and mutation webhook services

You can also define validation and mutation services:
- **Validation webhook** performs the validation of fetched assets before the Asset Controller uploads them into the bucket. It can be a list of several different validation webhooks and all of them should be processed even if one fails. It can refer either to the validation of a specific file against a specification or to the security validation. The validation webhook returns the validation status when the validation completes.
- **Mutation webhook** acts similarly to the validation service. The difference is that it mutates the asset instead of just validating it. For example, this can mean asset rewriting through the `regex` operation or `keyvalue`, or the modification in the JSON specification. The mutation webhook returns modified files instead of information on the status.

```
apiVersion: assetstore.kyma-project.io/v1alpha2
kind: Asset
metadata:
  name: my-package-assets
  namespace: default
spec:
  source:
    mode: single
    metadata:
      disableRelativeLinks: "true"
    url: https://some.domain.com/main.js
    validationWebhookService:
    - name: swagger-operations-svc
      namespace: default
      endpoint: "/validate"
      filter: \.js$
    mutationWebhookService:
    - name: swagger-operations-svc
      namespace: default
      endpoint: "/mutate"
      filter: \.js$
      metadata:
        rewrite: keyvalue
        pattern: \json|yaml
        data:
          basePath: /test/v2
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


| Parameter   |      Mandatory      |  Description |
|----------|:-------------:|------|
| **metadata.name** |    **YES**   | Specifies the name of the CR. |
| **metadata.namespace** |    **YES**   | Defines the Namespace in which the CR is available. |
| **spec.source.mode** |    **YES**   | Specifies if the asset consists of one file or a set of compressed files in the ZIP or TAR formats. Use `single` for one file and `package` for a set of files. |
| **spec.source.metadata** |    **NO**   | Specifies a set of metadata for the Asset. For example, use it to define what to render, disable, or modify in the UI. Define it in a valid YAML or JSON format. |
| **spec.source.url** |    **YES**   | Specifies the location of the file. |
| **spec.source.filter** |    **NO**   | Specifies the regex pattern used to select files to store from the package. |
| **spec.source.validationwebhookservice** |    **NO**   | Provides specification of the validation webhook services. |
| **spec.source.validationwebhookservice.name** |    **YES**   | Provides the name of the validation webhook service. |
| **spec.source.validationwebhookservice.namespace** |    **YES**   | Provides the Namespace in which the service is available. |
| **spec.source.validationwebhookservice.endpoint** |    **NO**   | Specifies the endpoint to which the service sends calls. |
| **spec.source.validationwebhookservice.filter** |    **NO**   | Specifies the regex pattern used to select files sent to the service. |
| **spec.source.mutationwebhookservice** |    **NO**   | Provides specification of the mutation webhook services. |
| **spec.source.mutationwebhookservice.name** |    **YES**   | Provides the name of the mutation webhook service. |
| **spec.source.mutationwebhookservice.namespace** |    **YES**   | Provides the Namespace in which the service is available. |
| **spec.source.mutationwebhookservice.endpoint** |    **NO**   | Specifies the endpoint to which the service sends calls. |
| **spec.source.mutationwebhookservice.filter** |    **NO**   | Specifies the regex pattern used to select files sent to the service. |
| **spec.source.metadatawebhookservice** |    **NO**   | Provides specification of the metadata webhook services. |
| **spec.source.metadatawebhookservice.name** |    **YES**   | Provides the name of the metadata webhook service. |
| **spec.source.metadatawebhookservice.namespace** |    **YES**   | Provides the Namespace in which the service is available. |
| **spec.source.metadatawebhookservice.endpoint** |    **NO**   | Specifies the endpoint to which the service sends calls. |
| **spec.source.metadatawebhookservice.filter** |    **NO**   | Specifies the regex pattern used to select files sent to the service. |
| **spec.bucketref.name** |    **YES**   | Provides the name of the bucket for storing the asset. |
| **status.phase** |    **Not applicable**   | The Asset Controller adds it to the Asset CR. It describes the status of processing the Asset CR by the Asset Controller. It can be `Ready`, `Failed`, or `Pending`. |
| **status.reason** |    **Not applicable**   | Provides the reason why the Asset CR processing failed or is pending.  |
| **status.message** |    **Not applicable**   | Describes a human-readable message on the CR processing progress, success, or failure. |
| **status.lastheartbeattime** |    **Not applicable**   | Provides the last time when the Asset Controller processed the Asset CR. |
| **status.observedGeneration** |    **Not applicable**   | Specifies the most recent generation that the Asset Controller observes. |
| **status.assetref** |    **Not applicable**   | Provides details on the location of the assets stored in the bucket.   |
| **status.assetref.assets** |    **Not applicable**   | Provides the relative path to the given asset in the storage bucket. |
| **status.assetref.baseurl** |    **Not applicable**   | Specifies the absolute path to the location of the assets in the storage bucket. |


> **NOTE:** The Asset Controller automatically adds all parameters marked as **Not applicable** to the Asset CR.


## Related resources and components

These are the resources related to this CR:

| Custom resource |   Description |
|----------|------|
| Bucket |  The Asset CR uses the name of the bucket specified in the definition of the Bucket CR. |

These components use this CR:

| Component   |   Description |
|----------|------|
| Asset Store |  Uses the Asset CR for the detailed asset definition, including its location and the name of the bucket in which it is stored. |
