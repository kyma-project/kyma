---
title: Function
type: Custom Resources
---

The `functions.serverless.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to manage Functions within Kyma. To get the up-to-date CRD and show the output in the YAML format, run this command:

```bash
kubectl get crd functions.serverless.kyma-project.io -o yaml
```

## Sample custom resource

The following Function object creates a Function which responds to HTTP requests with the "Hello John" message. The Function's code (**source**) and dependencies (**deps**) are specified in the Function CR.

```yaml
apiVersion: serverless.kyma-project.io/v1alpha1
kind: Function
metadata:
  name: my-test-function
  namespace: default
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
  labels:
    app: my-test-function
  minReplicas: 3
  maxReplicas: 3
  resources:
    limits:
      cpu: 1
      memory: 1Gi
    requests:
      cpu: 500m
      memory: 500Mi
  buildResources:
    limits:
      cpu: 2
      memory: 2Gi
    requests:
      cpu: 1
      memory: 1Gi
  source: |
    module.exports = {
      main: function(event, context) {
        const name = process.env.PERSON_NAME;
        return 'Hello ' + name;
      }
    }
  runtime: nodejs12
  status:
    conditions:
      - lastTransitionTime: "2020-04-14T08:17:11Z"
        message: "Deployment my-test-function-nxjdp is ready"
        reason: DeploymentReady
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

If you store the Function's source code and dependencies in a Git repository and want the Function Controller to fetch them from it, use these parameters in the Function CR:

```yaml
apiVersion: serverless.kyma-project.io/v1alpha1
kind: Function
metadata:
  name: my-test-function
spec:
  type: git
  source: auth-basic
  baseDir: "/"
  reference: "branchA"
  runtime: "nodejs12"
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:

| Parameter         |    Required    | Description                                   |
| ---------------------------------------- | :------------: | ---------|
| **metadata.name**              |      Yes       | Specifies the name of the CR.                 |
| **metadata.namespace**     |       No       | Defines the Namespace in which the CR is available. It is set to `default` unless you specify otherwise.      |
| **spec.env**                             |       No       | Specifies environment variables you need to export for the Function. You can export them either directly in the Function CR's spec or define them in a [ConfigMap](#configuration-environment-variables-define-environment-variables-in-a-config-map). |
| **spec.deps**                            |       No       | Specifies the Function's dependencies.  |
| **spec.labels**                          |       No       | Specifies the Function's Pod labels.    |
| **spec.minReplicas**                     |       No       | Defines the minimum number of Function's Pods to run at a time.  |
| **spec.maxReplicas**                     |       No       | Defines the maximum number of Function's Pods to run at a time.    |
| **spec.resources.limits.cpu**            |       No       | Defines the maximum number of CPUs available for the Function's Pod to use.      |
| **spec.resources.limits.memory**         |       No       | Defines the maximum amount of memory available for the Function's Pod to use.      |
| **spec.resources.requests.cpu**          |       No       | Specifies the number of CPUs requested by the Function's Pod to operate.       |
| **spec.resources.requests.memory**       |       No       | Specifies the amount of memory requested by the Function's Pod to operate.               |
| **spec.buildResources.limits.cpu**            |       No       | Defines the maximum number of CPUs available to use for the Kubernetes Job's Pod responsible for building the Function's image.      |
| **spec.buildResources.limits.memory**         |       No       | Defines the maximum amount of memory available for the Job's Pod to use.      |
| **spec.buildResources.requests.cpu**          |       No       | Specifies the number of CPUs requested by the build Job's Pod to operate.       |
| **spec.buildResources.requests.memory**       |       No       | Specifies the amount of memory requested by the build Job's Pod to operate.               |
| **spec.runtime**                         |       No       | Specifies the runtime of the Function. The available values are `nodejs12`, `python38` and `nodejs14`. It is set to `nodejs14` unless specified otherwise.  |
| **spec.type**                          |      No       | Defines that you use a Git repository as the source of Function's code and dependencies. It must be set to `git`. |
| **spec.source**                          |      Yes       | Provides the Function's full source code or the name of the Git directory in which the code and dependencies are stored.     |
| **spec.baseDir**                          |      No       | Specifies the relative path to the Git directory that contains the source code from which the Function will be built​. |
| **spec.reference**                        |      No       | Specifies either the branch name or the commit revision from which the Function Controller automatically fetches the changes in Function's code and dependencies. |
| **status.conditions.lastTransitionTime** | Not applicable | Provides a timestamp for the last time the Function's condition status changed from one to another.    |
| **status.conditions.message**            | Not applicable | Describes a human-readable message on the CR processing progress, success, or failure.   |
| **status.conditions.reason**             | Not applicable | Provides information on the Function CR processing success or failure. See the [**Reasons**](#status-reasons) section for the full list of possible status reasons and their descriptions. All status reasons are in camelCase.   |
| **status.conditions.status**             | Not applicable | Describes the status of processing the Function CR by the Function Controller. It can be `True` for success, `False` for failure, or `Unknown` if the CR processing is still in progress. If the status of all conditions is `True`, the overall status of the Function CR is ready.     |
| **status.conditions.type**               | Not applicable | Describes a substage of the Function CR processing. There are three condition types that a Function has to meet to be ready: `ConfigurationReady`, `BuildReady`, and `Running`. When displaying the Function status in the terminal, these types are shown under `CONFIGURED`, `BUILT`, and `RUNNING` columns respectively. All condition types can change asynchronously depending on the type of Function modification, but all three need to be in the `True` status for the Function to be considered successfully processed. |

### Status reasons

Processing of a Function CR can succeed, continue, or fail for one of these reasons:

| Reason                           | Type                 | Description                                                                                                                                                   |
| -------------------------------- | -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ConfigMapCreated`               | `ConfigurationReady` | A new ConfigMap was created based on the Function CR definition.                                                                                              |
| `ConfigMapUpdated`               | `ConfigurationReady` | The existing ConfigMap was updated after changes in the Function CR name, its source code or dependencies.                                                    |
| `SourceUpdated`                  | `ConfigurationReady` | The Function Controller managed to fetch changes in the Functions's source code and configuration from the Git repository (`type: git`).                |
| `SourceUpdateFailed`             | `ConfigurationReady` | The Function Controller failed to fetch changes in the Functions's source code and configuration from the Git repository.                            |
| `JobFailed`                      | `BuildReady`         | The image with the Function's configuration could not be created due to an error.                                                                             |
| `JobCreated`                     | `BuildReady`         | The Kubernetes Job resource that builds the Function image was created.                                                                                       |
| `JobUpdated`                     | `BuildReady`         | The existing Job was updated after changing the Function's metadata or spec fields that do not affect the way of building the Function image, such as labels. |
| `JobRunning`                     | `BuildReady`         | The Job is in progress.                                                                                                                                       |
| `JobsDeleted`                    | `BuildReady`         | Previous Jobs responsible for building Function images were deleted.                                                                                          |
| `JobFinished`                    | `BuildReady`         | The Job was finished and the Function's image was uploaded to the Docker Registry.                                                                            |
| `DeploymentCreated`              | `Running`            | A new Deployment referencing the Function's image was created.                                                                                                |
| `DeploymentUpdated`              | `Running`            | The existing Deployment was updated after changing the Function's image, scaling parameters, variables, or labels.                                            |
| `DeploymentFailed`               | `Running`            | The Function's Pod crashed or could not start due to an error.                                                                                                |
| `DeploymentWaiting`              | `Running`            | The Function was deployed and is waiting for the Deployment to be ready.                                                                                      |
| `DeploymentReady`                | `Running`            | The Function was deployed and is ready.                                                                                                                       |
| `ServiceCreated`                 | `Running`            | A new Service referencing the Function's Deployment was created.                                                                                              |
| `ServiceUpdated`                 | `Running`            | The existing Service was updated after applying required changes.                                                                                             |
| `HorizontalPodAutoscalerCreated` | `Running`            | A new HorizontalPodScaler referencing the Function's Deployment was created.                                                                                  |
| `HorizontalPodAutoscalerUpdated` | `Running`            | The existing HorizontalPodScaler was updated after applying required changes.                                                                                 |

## Related resources and components

These are the resources related to this CR:

| Custom resource                                                                                              | Description                                                                           |
| ----------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| [ConfigMap](https://kubernetes.io/docs/concepts/configuration/configmap/)                             | Stores the Function's source code and dependencies.                                   |
| [Job](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/)              | Builds an image with the Function's code in a runtime.                                |
| [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)                   | Serves the Function's image as a microservice.                                        |
| [Service](https://kubernetes.io/docs/concepts/services-networking/service/)                           | Exposes the Function's Deployment as a network service inside the Kubernetes cluster. |
| [HorizontalPodAutoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) | Automatically scales the number of Function's Pods.                                   |

These components use this CR:

| Component           | Description                                                                                                  |
| ------------------- | ------------------------------------------------------------------------------------------------------------ |
| Function Controller | Uses the Function CR for the detailed Function definition, including the environment on which it should run. |
