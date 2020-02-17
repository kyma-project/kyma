---
title: Function
type: Custom Resource
---

The `functions.serverless.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to manage lambdas within Kyma. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```bash
kubectl get crd functions.serverless.kyma-project.io -o yaml
```

## Sample custom resource

The following Function object creates a lambda which runs on the Node.js 8 runtime and responds to HTTP requests with "Hello John".

```yaml
apiVersion: serverless.kyma-project.io/v1alpha1
kind: Function
metadata:
  name: my-test-lambda
  namespace: default
spec:
  functionContentType: plaintext
  runtime: nodejs8
  timeout: 360
  env:
  - name: PERSON_NAME
    value: "John"
  deps: |
    {
      "name": "hellowithdeps",
      "version": "0.0.1",
      "dependencies": {
        "end-of-stream": "^1.4.1",
        "from2": "^2.3.0",
        "lodash": "^4.17.5"
      }
    }
  function: |
    module.exports = {
      main: function(event, context) {
        const name = process.env.PERSON_NAME;
        return 'Hello ' + name;
      }
    }
```

## Custom resource properties

This table lists all the possible properties of a given resource together with their descriptions:

| Property | Required | Description |
|----------|:---------:|-------------|
| **metadata.name** | Yes | Specifies the name of the CR. |
| **metadata.namespace** | No | Defines the Namespace in which the CR is available. It is set to `default` unless you specify otherwise. |
| **spec.functionContentType** | No | Specifies the content type of the lambda's code defined in the **function** property. The content type can be either `plaintext` or `base64`. It is set to `plaintext` unless you specify otherwise.|
| **spec.runtime** | No | Specifies the software runtime used to run the lambda's code. The Function Controller supports `nodejs6` and `nodejs8`. It is set to `nodejs8` unless you specify otherwise. |
| **spec.timeout** | No | Specifies the duration in seconds after which the lambda execution is terminated. The default value is `180`. |
| **spec.env** | No | Specifies environment variables you need to export for the lambda. |
| **spec.deps** | No | Specifies the lambda's dependencies. |
| **spec.function** | Yes | Provides the lambda's source code. |

## Related resources and components

The Function custom resource relies on resources from [Knative Serving](https://knative.dev/docs/serving/) and [Tekton Pipelines](https://github.com/tektoncd/pipeline):

| Resource | Description |
|----------|-------------|
|[KService CR](https://github.com/knative/docs/blob/master/docs/serving/spec/knative-api-specification-1.0.md#service) | Orchestrates the deployment and availability of the function.|
|[TestRun CR](https://github.com/tektoncd/pipeline/blob/master/docs/taskruns.md) | Builds a container image with the function code on a chosen runtime. |

These components use this CR:

| Component   |   Description |
|-------------|---------------|
| Function Controller |  Uses the Function CR for the detailed lambda definition, including the environment on which it should run. |
