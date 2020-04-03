---
title: Function
type: Custom Resource
---

The `functions.serverless.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to manage lambdas within Kyma. To get the up-to-date CRD and show the output in the YAML format, run this command:

```bash
kubectl get crd functions.serverless.kyma-project.io -o yaml
```

## Sample custom resource

The following Function object creates a lambda which responds to HTTP requests with "Hello John".

```yaml
apiVersion: serverless.kyma-project.io/v1alpha1
kind: Function
metadata:
  name: my-test-lambda
spec:
  functionContentType: plaintext
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
  status:
    phase: Running
    conditions:
      [...]
      - type: Deployed
        lasttransitiontime: "2020-03-30T11:57:59+02:00"
        reason: DeploySucceeded
        message: ""
    observedgeneration: 3
    imagetag: dd5f487a-fc44-4f06-90d5-196dd7246b92
```

## Custom resource properties

This table lists all the possible properties of a given resource together with their descriptions:

| Property | Required | Description |
|----------|:---------:|-------------|
| **metadata.name** | Yes | Specifies the name of the CR. |
| **metadata.namespace** | No | Defines the Namespace in which the CR is available. It is set to `default` unless you specify otherwise. |
| **spec.functionContentType** | No | Specifies the content type of the lambda's code defined in the **function** property. The content type can be either `plaintext` or `base64`. It is set to `plaintext` unless you specify otherwise.|
| **spec.timeout** | No | Specifies the duration in seconds after which the lambda execution is terminated. The default value is `180`. |
| **spec.env** | No | Specifies environment variables you need to export for the lambda. |
| **spec.deps** | No | Specifies the lambda's dependencies. |
| **spec.function** | Yes | Provides the lambda's source code. |
| **status.phase** | Not applicable | The Function Controller adds it to the Function CR. It describes the status of processing the Function CR by the Function Controller. It can be `Initializing`, `Building`, `Deploying`, `Running`, `Error`, or `Failed`. |
| **status.conditions.type** | Not applicable | Describes a substage of the Function CR processing phase. |
| **status.conditions.lasttransitiontime** | Not applicable | Provides a timestamp for the last time the lambda's Pod transitioned from one status to another. |
| **status.conditions.reason** | Not applicable | Provides information on the Function CR processing success or failure. See the [**Reasons**](#status-reasons) section for the full list of possible status reasons and their descriptions. |
| **status.conditions.message** | Not applicable | Describes a human-readable message on the CR processing progress, success, or failure.  |
| **status.observedgeneration** | Not applicable | Specifies the number of times the Function Controller processed the Function CR. |
| **status.imagetag** | Not applicable | Specifies the current tag of the image generated for the given lambda. |

### Status reasons

Processing of a Function CR can succeed, continue, or fail for one of these reasons:

| Reason | ConditionType | Phase | Description |
| --------- | ------------- | ----------- |----------- |
| `CreateConfigSucceeded` | `Initialized` | `Initializing` | The ConfigMap with the lambda's source code and dependencies was created. |
| `CreateConfigFailed` | `Error` | `Initializing` | The ConfigMap couldn't be created due to an error. |
| `GetConfigFailed` | `Error` | `Initializing` | Failed to get the current ConfigMap for an update due to an error. |
| `UpdateConfigSucceeded` | `Initialized` | `Initializing` | The ConfigMap with the lambda's source code and dependencies was updated. |
| `UpdateConfigFailed` | `Error` | `Initializing` | The ConfigMap couldn't be updated due to an error. |
| `UpdateRuntimeConfig` | `Initialized` | `Initializing` | Environment variables for a lambda were updated. |
| `BuildSucceeded` | `ImageCreated` | `Building` | The image with the lambda's configuration was created and uploaded to the Docker registry. |
| `BuildFailed` | `Error` | `Building` | The image with the lambda's configuration couldn't be created due to an error. |
| `CreateServiceSucceeded` | `Deploying` | `Deploying` | The KService was created. |
| `UpdateServiceSucceeded` | `Deploying` | `Deploying` | The KService was updated.  |
| `DeploySucceeded` | `Deployed` | `Deploying` | The lambda was deployed in the Namespace. |
| `DeployFailed` | `Error` | `Deploying` | The lambda couldn't be deployed in the Namespace due to an error. |
| `Unknown` | `Error` | `Error` or `Failed` | The Function Controller failed to process the Function CR or stopped processing it due to an unexpected error. |

## Related resources and components

The Function custom resource relies on these resources from [Knative Serving](https://knative.dev/docs/serving/) and [Tekton Pipelines](https://github.com/tektoncd/pipeline):

| Resource | Description |
|----------|-------------|
|[KService CR](https://github.com/knative/docs/blob/master/docs/serving/spec/knative-api-specification-1.0.md#service) | Orchestrates the deployment and availability of the function.|
|[TaskRun CR](https://github.com/tektoncd/pipeline/blob/master/docs/taskruns.md) | Builds an image with the function code on a chosen runtime. |

These components use this CR:

| Component   |   Description |
|-------------|---------------|
| Function Controller |  Uses the Function CR for the detailed lambda definition, including the environment on which it should run. |
