---
title: Function
type: Custom Resource
---

The `functions.serverless.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to manage lambda functions within Kyma. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```bash
kubectl get crd functions.serverless.kyma-project.io -o yaml
```

## Sample custom resource

The following Function object creates a lambda function which runs on the Node.js 8 runtime and responds to HTTP requests with "Hello World". It has low requirements in terms of compute resources and is therefore classified as size S.

```yaml
apiVersion: serverless.kyma-project.io/v1alpha1
kind: Function
metadata:
  name: hello-world
spec:
  functionContentType: plaintext
  size: S
  runtime: nodejs8
  function: |
    module.exports = {
      main: function(event, context) {
        return 'Hello World'
      }
    }
```

## Custom resource properties

This table lists all the possible properties of a given resource together with their descriptions:

| Property | Required | Description |
|----------|:---------:|-------------|
| **metadata.name** | Yes | Specifies the name of the CR. |
| **spec.function** | Yes | Provides the source code of the lambda function. |
| **spec.functionContentType** | Yes | Specifies the content type of the function's code defined in the **function** property. The content type can be plaintext or base64-encoded. |
| **spec.runtime** | Yes | Specifies the software runtime used to run the function's code. |
| **spec.size** | Yes | Specifies the compute requirement of the function expressed in size, such as S, M, L, or XL. |
| **spec.deps** | No | Specifies the dependencies of the lambda function. |
| **spec.env** | No | Specifies environment variables you need to export for the lambda function. |
| **spec.timeout** | No | Specifies the duration in seconds after which the function execution is terminated. The default value is `180`. |

## Related resources and components

The Function custom resource relies on resources from [Knative Serving](https://knative.dev/v0.7-docs/serving/) and [Knative Build](https://knative.dev/v0.7-docs/build/).

| Resource | Description |
|----------|-------------|
|[Build](https://knative.dev/v0.7-docs/reference/build-api/) | Builds a container image containing the function code together with its configured runtime. |
|[Service](https://knative.dev/v0.7-docs/reference/serving-api/) | Orchestrates the deployment and availability of the function.|
