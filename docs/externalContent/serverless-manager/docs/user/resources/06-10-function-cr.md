# Function

The `functions.serverless.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to manage Functions within Kyma. To get the up-to-date CRD and show the output in the YAML format, run this command:

```bash
kubectl get crd functions.serverless.kyma-project.io -o yaml
```

## Sample Custom Resource

The following Function object creates a Function which responds to HTTP requests with the "Hello John" message. The Function's code (**source**) and dependencies (**dependencies**) are specified in the Function CR.

```yaml
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
metadata:
  name: my-test-function
  namespace: default
  labels:
    app: my-test-function
spec:
  runtime: nodejs20
  source:
    inline:
      dependencies: |
        {
          "name": "hellowithdeps",
          "version": "0.0.1",
          "dependencies": {
            "end-of-stream": "^1.4.1",
            "from2": "^2.3.0",
            "lodash": "^4.17.5"
          }
        }
      source: |
        module.exports = {
          main: function(event, context) {
            const name = process.env.PERSON_NAME;
            return 'Hello ' + name;
          }
        }
  scaleConfig:
    minReplicas: 3
    maxReplicas: 3
  resourceConfiguration:
    function:
      resources:
        limits:
          cpu: 1
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 500Mi
    build:
      resources:
        limits:
          cpu: 2
          memory: 2Gi
        requests:
          cpu: 1
          memory: 1Gi
  env:
    - name: PERSON_NAME
      value: "John"
  secretMounts:
    - secretName: SECRET_NAME
      mountPath: /secret/mount/path
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
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
metadata:
  name: my-test-function
spec:
  source:
    gitRepository:
      url: github.com/username/repo
      baseDir: "/"
      reference: "branchA"
      auth:
        type: basic
        secretName: secret-name
  runtime: "nodejs20"
```

## Custom Resource Parameters
<!-- TABLE-START -->
### Function.serverless.kyma-project.io/v1alpha2

**Spec:**

| Parameter | Type | Description                                                                                                                                                                                                                                                                                                                                                 |
| ---- | ----------- |-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **annotations**  | map\[string\]string | Defines annotations used in Deployment's PodTemplate and applied on the Function's runtime Pod.                                                                                                                                                                                                                                                             |
| **env**  | \[\]object | Specifies an array of key-value pairs to be used as environment variables for the Function. You can define values as static strings or reference values from ConfigMaps or Secrets. For configuration details, see the [official Kubernetes documentation](https://kubernetes.io/docs/tasks/inject-data-application/define-environment-variable-container/). |
| **labels**  | map\[string\]string | Defines labels used in Deployment's PodTemplate and applied on the Function's runtime Pod.                                                                                                                                                                                                                                                                  |
| **replicas**  | integer | Defines the exact number of Function's Pods to run at a time. If **ScaleConfig** is configured, or if the Function is targeted by an external scaler, then the **Replicas** field is used by the relevant HorizontalPodAutoscaler to control the number of active replicas.                                                                                 |
| **resourceConfiguration**  | object | Specifies resources requested by the Function and the build Job.                                                                                                                                                                                                                                                                                            |
| **resourceConfiguration.&#x200b;build**  | object | **Deprecated: Serverless won't build images a future version** Specifies resources requested by the build Job's Pod.                                                                                                                                                                                                                                                                                                       |
| **resourceConfiguration.&#x200b;build.&#x200b;profile**  | string | Defines the name of the predefined set of values of the resource. Can't be used together with **Resources**.                                                                                                                                                                                                                                                |
| **resourceConfiguration.&#x200b;build.&#x200b;resources**  | object | Defines the amount of resources available for the Pod. Can't be used together with **Profile**. For configuration details, see the [official Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/).                                                                                                     |
| **resourceConfiguration.&#x200b;function**  | object | Specifies resources requested by the Function's Pod.                                                                                                                                                                                                                                                                                                        |
| **resourceConfiguration.&#x200b;function.&#x200b;profile**  | string | Defines the name of the predefined set of values of the resource. Can't be used together with **Resources**.                                                                                                                                                                                                                                                |
| **resourceConfiguration.&#x200b;function.&#x200b;resources**  | object | Defines the amount of resources available for the Pod. Can't be used together with **Profile**. For configuration details, see the [official Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/).                                                                                                     |
| **runtime** (required) | string | Specifies the runtime of the Function. The available values are `nodejs20` and `python312`.                                                                                                                                                                                                   |
| **runtimeImageOverride**  | string | Specifies the runtime image used instead of the default one.                                                                                                                                                                                                                                                                                                |
| **scaleConfig**  | object | **Deprecated: Serverless won't create HPA in a future version** Defines the minimum and maximum number of Function's Pods to run at a time. When it is configured, a HorizontalPodAutoscaler will be deployed and will control the **Replicas** field to scale the Function based on the CPU utilization.                                                                                                                   |
| **scaleConfig.&#x200b;maxReplicas** (required) | integer | Defines the maximum number of Function's Pods to run at a time.                                                                                                                                                                                                                                                                                             |
| **scaleConfig.&#x200b;minReplicas** (required) | integer | Defines the minimum number of Function's Pods to run at a time.                                                                                                                                                                                                                                                                                             |
| **secretMounts**  | \[\]object | Specifies Secrets to mount into the Function's container filesystem.                                                                                                                                                                                                                                                                                        |
| **secretMounts.&#x200b;mountPath** (required) | string | Specifies the path within the container where the Secret should be mounted.                                                                                                                                                                                                                                                                                 |
| **secretMounts.&#x200b;secretName** (required) | string | Specifies the name of the Secret in the Function's namespace.                                                                                                                                                                                                                                                                                               |
| **source** (required) | object | Contains the Function's source code configuration.                                                                                                                                                                                                                                                                                                          |
| **source.&#x200b;gitRepository**  | object | Defines the Function as Git-sourced. Can't be used together with **Inline**.                                                                                                                                                                                                                                                                                |
| **source.&#x200b;gitRepository.&#x200b;auth**  | object | Specifies the authentication method. Required for SSH.                                                                                                                                                                                                                                                                                                      |
| **source.&#x200b;gitRepository.&#x200b;auth.&#x200b;secretName** (required) | string | Specifies the name of the Secret with credentials used by the Function Controller to authenticate to the Git repository in order to fetch the Function's source code and dependencies. This Secret must be stored in the same namespace as the Function CR.                                                                                                 |
| **source.&#x200b;gitRepository.&#x200b;auth.&#x200b;type** (required) | string | Defines the repository authentication method. The value is either `basic` if you use a password or token, or `key` if you use an SSH key.                                                                                                                                                                                                                   |
| **source.&#x200b;gitRepository.&#x200b;baseDir**  | string | Specifies the relative path to the Git directory that contains the source code from which the Function is built.                                                                                                                                                                                                                                            |
| **source.&#x200b;gitRepository.&#x200b;reference**  | string | Specifies either the branch name, tag or commit revision from which the Function Controller automatically fetches the changes in the Function's code and dependencies.                                                                                                                                                                                      |
| **source.&#x200b;gitRepository.&#x200b;url** (required) | string | Specifies the URL of the Git repository with the Function's code and dependencies. Depending on whether the repository is public or private and what authentication method is used to access it, the URL must start with the `http(s)`, `git`, or `ssh` prefix.                                                                                             |
| **source.&#x200b;inline**  | object | Defines the Function as the inline Function. Can't be used together with **GitRepository**.                                                                                                                                                                                                                                                                 |
| **source.&#x200b;inline.&#x200b;dependencies**  | string | Specifies the Function's dependencies.                                                                                                                                                                                                                                                                                                                      |
| **source.&#x200b;inline.&#x200b;source** (required) | string | Specifies the Function's full source code.                                                                                                                                                                                                                                                                                                                  |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **baseDir**  | string | Specifies the relative path to the Git directory that contains the source code from which the Function is built. |
| **commit**  | string | Specifies the commit hash used to build the Function. |
| **conditions**  | \[\]object | Specifies an array of conditions describing the status of the parser. |
| **conditions.&#x200b;lastTransitionTime**  | string | Specifies the last time the condition transitioned from one status to another. |
| **conditions.&#x200b;message**  | string | Provides a human-readable message indicating details about the transition. |
| **conditions.&#x200b;reason**  | string | Specifies the reason for the condition's last transition. |
| **conditions.&#x200b;status** (required) | string | Specifies the status of the condition. The value is either `True`, `False`, or `Unknown`. |
| **conditions.&#x200b;type**  | string | Specifies the type of the Function's condition. |
| **podSelector**  | string | Specifies the Pod selector used to match Pods in the Function's Deployment. |
| **reference**  | string | Specifies either the branch name, tag or commit revision from which the Function Controller automatically fetches the changes in the Function's code and dependencies. |
| **replicas**  | integer | Specifies the total number of non-terminated Pods targeted by this Function. |
| **runtime**  | string | Specifies the **Runtime** type of the Function. |
| **runtimeImage**  | string | Specifies the image version used to build and run the Function's Pods. |
| **runtimeImageOverride**  | string | Deprecated: Specifies the runtime image version which overrides the **RuntimeImage** status parameter. **RuntimeImageOverride** exists for historical compatibility and should be removed with v1alpha3 version. |

<!-- TABLE-END -->

### Status Reasons

Processing of a Function CR can succeed, continue, or fail for one of these reasons:

| Reason                           | Type                 | Description                                                                                                                                                   |
| -------------------------------- | -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ConfigMapCreated`               | `ConfigurationReady` | A new ConfigMap was created based on the Function CR definition.                                                                                              |
| `ConfigMapUpdated`               | `ConfigurationReady` | The existing ConfigMap was updated after changes in the Function CR name, its source code or dependencies.                                                    |
| `SourceUpdated`                  | `ConfigurationReady` | The Function Controller managed to fetch changes in the Functions's source code and configuration from the Git repository.                |
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
| `ServiceFailed`                  | `Running`            | The Function's service could not be created or updated.                                                                                             |
| `HorizontalPodAutoscalerCreated` | `Running`            | A new Horizontal Pod Scaler referencing the Function's Deployment was created.                                                                                  |
| `HorizontalPodAutoscalerUpdated` | `Running`            | The existing Horizontal Pod Scaler was updated after applying required changes.                                                                                 |
| `MinimumReplicasUnavailable`     | `Running`            | Insufficient number of available Replicas. The Function is unhealthy.                                                                                                       |

## Related Resources and Components

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
