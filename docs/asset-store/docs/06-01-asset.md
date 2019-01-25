---
title: Asset
type: Custom Resource
---

The `assets.assetstore.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define an asset to store in a cloud storage bucket. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd assets.assetstore.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource (CR) that provides details of the bucket for storing assets.

```
apiVersion: assetstore.kyma-project.io/v1alpha1
kind: Asset
metadata:
  name: my-package-assets
  namespace: default
spec:
  source:
    mode: single / index / package
    url: https://some.domain.com/main.js
  bucketRef:
    name: my-bucket

```

### Sample index file

The CR must contain the reference to the source file, object, or asset location that the Asset Controller fetches with three different modes, where:
- `single` specifies the direct link to the asset that needs to be fetched.
- `index` specifies the link to the `index.yaml` file that contains a reference to files that need to be separately fetched from a given relative location.
- `package` specifies the link to the zip or tar file that must be unzipped before uploading.

See the sample of the index file with markdown and assets files:

```
apiVersion: v1
files:
  - name: 01-overview.md
    metadata:
      title: MyOverview
      type: Overview
  - name: 02-details.md
    metadata:
      title: MyDetails
      type: Details
  - name: 03-installation.md
    metadata:
      title: MyInstallation
      type: Tutorial
  - name: assets/diagram.svg
```

### Validation and mutation webhook services

You can also define validation and mutation services:
- **Validation webhook** reference to a service that performs the validation of fetched assets before they are uploaded to the bucket. It can be a list of several different validation webhooks and all of them should be processed even if one is failing. It can refer either to the validation of a specific file against a specification or to the security validation. The validation webhook returns the validation status.
- **Mutation webhook** reference to a service that acts similarly to the validation service. The difference is that it mutates the asset instead of just validating it. For example, this can be an asset rewriting through the `regex` operation or `keyvalue`, or the modification in the JSON specification. The mutation webhook returns modified files instead of information on the status.

```
apiVersion: assetstore.kyma-project.io/v1alpha1
kind: Asset
metadata:
  name: my-package-assets
  namespace: default
spec:
  source:
    mode: single | index | package
    url: https://some.domain.com/main.js
    validationWebhookService:
      name: swagger-operations-svc
      namespace: default
      endpoint: "/validate"
    mutationWebhookService:
      name: swagger-operations-svc
      namespace: default
      endpoint: "/mutate"
      metadata:
        rewrite: keyvalue
        pattern: \json|yaml
        data:
          basePath: /test/v2
  bucketRef:
    name: my-bucket
status:
  phase: Ready | Failed | Pending
  reason: ValidationFailure
  message: "The file is not valid against the provided json schema"
  lastHeartbeatTime: "2018-01-03T07:38:24Z"

```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:


| Parameter   |      Mandatory      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR. |
| **metadata.namespace** |    **YES**   | Defines the Namespace in which the CR is available. |
| **spec.source.mode** |    **YES**   | Specifies if the asset is a single file, zip or tar, or a `index.yaml` file. For details on the **index** mode, see [this](#sample-index-file) section. |
| **spec.source.url** |    **YES**   | Specifies the location of the file. |
| **bucketref.name** |    **YES**   | Specifies the name of the bucket for storing the asset. |
| **spec.source.validationwebhookservice** |    **NO**   | Provides specification of the validation webhook service. |
| **spec.source.mutationwebhookservice** |    **NO**   | Provides specification of the validation webhook service. |


## Related resources and components

These are the resources related to this CR:

| Custom resource |   Description |
|:----------:|:------|
| Bucket |  The Asset CR uses the name of the bucket specified in the definition of the Bucket CR. |

These components use this CR:

| Component   |   Description |
|:----------:|:------|
| Asset Store |  Uses the Asset CR for the details definition of the asset, including its location, and the name of the bucket in which it is stored. |
