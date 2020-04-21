---
title: Function
type: Custom Resource
---

The `functions.serverless.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to manage functions within Kyma. To get the up-to-date CRD and show the output in the YAML format, run this command:

```bash
kubectl get crd functions.serverless.kyma-project.io -o yaml
```

## Sample custom resource

The following Function object creates a function which responds to HTTP requests with the "Hello John" message.

```yaml
apiVersion: serverless.kyma-project.io/v1alpha1
kind: Function
metadata:
  name: my-test-function
spec:
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
  minReplicas: 3
  maxReplicas: 3
  resources:
    limits:
      cpu: 1
      memory: 1Gi
    requests:
      cpu: 500m
      memory: 500Mi  
  source: |
    module.exports = {
      main: function(event, context) {
        const name = process.env.PERSON_NAME;
        return 'Hello ' + name;
      }
    }
  status:
    conditions:
      - lastTransitionTime: "2020-04-14T08:17:11Z"
        message: "Function my-test-function is ready"
        reason: ServiceReady
        status: "True"
        type: Running
      - lastTransitionTime: "2020-04-14T08:16:55Z"
        message: "Job my-test-function-build-552ft finished"
        reason: JobFinished
        status: "True"
        type: BuildReady
      - lastTransitionTime: "2020-04-14T08:16:16Z"
        message: "ConfigMap my-test-function-xv6pc created"
        reason: ConfigMapCreated
        status: "True"
        type: ConfigurationReady
```

## Custom resource properties

This table lists all the possible properties of a given resource together with their descriptions:

| Property | Required | Description |
|----------|:---------:|-------------|
| **metadata.name** | Yes | Specifies the name of the CR. |
| **metadata.namespace** | No | Defines the Namespace in which the CR is available. It is set to `default` unless you specify otherwise. |
| **spec.env** | No | Specifies environment variables you need to export for the function. |
| **spec.deps** | No | Specifies the function's dependencies. |
| **spec.minReplicas** | No | Defines the minimum number of function's Pods to run at a time. |
| **spec.maxReplicas** | No | Defines the maximum number of function's Pods to run at a time. |
| **spec.resources.limits.cpu** | No | Defines the maximum number of CPUs available for the function's Pod to use. |
| **spec.resources.limits.memory** | No | Defines the maximum amount of memory available for the function's Pod to use. |
| **spec.resources.requests.cpu** | No |  Specifies the number of CPUs requested by the function's Pod to operate. |
| **spec.resources.requests.memory** | No |  Specifies the amount of memory requested by the function's Pod to operate. |
| **spec.source** | Yes | Provides the function's source code. |
| **status.conditions.lastTransitionTime** | Not applicable | Provides a timestamp for the last time the function's condition status changed from one to another. |
| **status.conditions.message** | Not applicable | Describes a human-readable message on the CR processing progress, success, or failure.  |
| **status.conditions.reason** | Not applicable | Provides information on the Function CR processing success or failure. See the [**Reasons**](#status-reasons) section for the full list of possible status reasons and their descriptions. All status reasons are in camelCase. |
| **status.conditions.status** | Not applicable | Describes the status of processing the Function CR by the Function Controller. It can be `True` for success, `False` for failure, or `Unknown` if the CR processing is still in progress. If the status of all conditions is `True`, the overall status of the Function CR is ready. |
| **status.conditions.type** | Not applicable | Describes a substage of the Function CR processing. There are three condition types that a function has to meet to be ready: `ConfigurationReady`, `BuildReady`, and `Running`. When displaying the function status in the terminal, these types are shown under `CONFIGURED`, `BUILT`, and `RUNNING` columns respectively. All condition types can change asynchronously depending on the type of function modification, but all three need to be in the `True` status for the function to be considered successfully processed. |

### Status reasons

Processing of a Function CR can succeed, continue, or fail for one of these reasons:

| Reason | Type | Description |
| --------- | ------------- | ----------- |----------- |
| `ConfigMapCreated` | `ConfigurationReady` | A new ConfigMap was created based on the Function CR definition. |
| `ConfigMapUpdated` | `ConfigurationReady` | The existing ConfigMap was updated after changes in the Function CR name, its source code or dependencies. |
| `ConfigMapError` | `ConfigurationReady` | The ConfigMap could not be created or updated due to an error. |
| `JobFailed` | `BuildReady` | The image with the function's configuration could not be created due to an error. |
| `JobCreated` | `BuildReady` | The Kubernetes Job resource that builds the function image was created. |
| `JobRunning` | `BuildReady` | The Job is in progress.  |
| `JobsDeleted` | `BuildReady` | Previous Jobs responsible for building function images were deleted. |
| `JobFinished` | `BuildReady` | The Job was finished and the function's image was uploaded to the Docker Registry. |
| `ServiceCreated` | `Running` | A new KService referencing the function's image was created. |
| `ServiceUpdated` | `Running` | The existing KService was updated after such changes as the function's image, scaling parameters, variables, or labels. |
| `ServiceFailed` | `Running` | The function's Pod crashed or could not start due to an error. |
| `ServiceWaiting` | `Running` | Creation or update of the KService is in progress. |
| `ServiceReady` | `Running` | The function was deployed in the Namespace. |

## Related resources and components

The Function custom resource relies on these Kubernetes and [Knative Serving](https://knative.dev/docs/serving/) resources:

| Resource | Description |
|----------|-------------|
|[KService CR](https://github.com/knative/docs/blob/master/docs/serving/spec/knative-api-specification-1.0.md#service) | Orchestrates the deployment and availability of the function.|
|[Kubernetes Job](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/) | Builds an image with the function code on a runtime. |

These components use this CR:

| Component   |   Description |
|-------------|---------------|
| Function Controller |  Uses the Function CR for the detailed function definition, including the environment on which it should run. |
