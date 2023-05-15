---
title: Function
---

>**WARNING:** The current API version is `serverless.kyma-project.io/v1alpha2`. The `serverless.kyma-project.io/v1alpha1` version is still supported but deprecated. For the v1alpha1 version, see the [previous Function documentation](https://github.com/kyma-project/kyma/blob/release-2.5/docs/05-technical-reference/00-custom-resources/svls-01-function.md).

The `functions.serverless.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to manage Functions within Kyma. To get the up-to-date CRD and show the output in the YAML format, run this command:

```bash
kubectl get crd functions.serverless.kyma-project.io -o yaml
```

## Sample custom resource

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
  runtime: nodejs16
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
  runtime: "nodejs16"
```

## Custom resource parameters
<!-- TABLE-START -->
### Function.serverless.kyma-project.io/v1alpha2

**Spec:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **annotations**  | map\[string\]string | Annotations will be used in Deployment's PodTemplate and will be applied on the function's runtime Pod. |
| **env**  | \[\]object | Env specifies an array of key-value pairs to be used as environment variables for the Function. You can define values as static strings or reference values from ConfigMaps or Secrets. |
| **env.&#x200b;name** (required) | string | Name of the environment variable. Must be a C_IDENTIFIER. |
| **env.&#x200b;value**  | string | Variable references $(VAR_NAME) are expanded using the previously defined environment variables in the container and any service environment variables. If a variable cannot be resolved, the reference in the input string will be unchanged. Double $ are reduced to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e. "$(VAR_NAME)" will produce the string literal "$(VAR_NAME)". Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to "". |
| **env.&#x200b;valueFrom**  | object | Source for the environment variable's value. Cannot be used if value is not empty. |
| **env.&#x200b;valueFrom.&#x200b;configMapKeyRef**  | object | Selects a key of a ConfigMap. |
| **env.&#x200b;valueFrom.&#x200b;configMapKeyRef.&#x200b;key** (required) | string | The key to select. |
| **env.&#x200b;valueFrom.&#x200b;configMapKeyRef.&#x200b;name**  | string | Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid? |
| **env.&#x200b;valueFrom.&#x200b;configMapKeyRef.&#x200b;optional**  | boolean | Specify whether the ConfigMap or its key must be defined |
| **env.&#x200b;valueFrom.&#x200b;fieldRef**  | object | Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels['<KEY>']`, `metadata.annotations['<KEY>']`, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs. |
| **env.&#x200b;valueFrom.&#x200b;fieldRef.&#x200b;apiVersion**  | string | Version of the schema the FieldPath is written in terms of, defaults to "v1". |
| **env.&#x200b;valueFrom.&#x200b;fieldRef.&#x200b;fieldPath** (required) | string | Path of the field to select in the specified API version. |
| **env.&#x200b;valueFrom.&#x200b;resourceFieldRef**  | object | Selects a resource of the container: only resources limits and requests (limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported. |
| **env.&#x200b;valueFrom.&#x200b;resourceFieldRef.&#x200b;containerName**  | string | Container name: required for volumes, optional for env vars |
| **env.&#x200b;valueFrom.&#x200b;resourceFieldRef.&#x200b;divisor**  | UNKNOW TYPE | Specifies the output format of the exposed resources, defaults to "1" |
| **env.&#x200b;valueFrom.&#x200b;resourceFieldRef.&#x200b;resource** (required) | string | Required: resource to select |
| **env.&#x200b;valueFrom.&#x200b;secretKeyRef**  | object | Selects a key of a secret in the pod's namespace |
| **env.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;key** (required) | string | The key of the secret to select from.  Must be a valid secret key. |
| **env.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;name**  | string | Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid? |
| **env.&#x200b;valueFrom.&#x200b;secretKeyRef.&#x200b;optional**  | boolean | Specify whether the Secret or its key must be defined |
| **labels**  | map\[string\]string | Labels will be used in Deployment's PodTemplate and will be applied on the function's runtime Pod. |
| **replicas**  | integer | Replicas defines the exact number of Function's Pods to run at a time. If ScaleConfig is configured, or if Function is targeted by an external scaler, then the Replicas field is used by the relevant HorizontalPodAutoscaler to control the number of active replicas. |
| **resourceConfiguration**  | object | ResourceConfiguration specifies resources requested by Function and build Job. |
| **resourceConfiguration.&#x200b;build**  | object | Build specifies resources requested by the build Job's Pod. |
| **resourceConfiguration.&#x200b;build.&#x200b;profile**  | string | Profile defines name of predefined set of values of resource. Can't be used at the same time with Resources. |
| **resourceConfiguration.&#x200b;build.&#x200b;resources**  | object | Resources defines amount of resources available for the Pod to use. Can't be used at the same time with Profile. |
| **resourceConfiguration.&#x200b;build.&#x200b;resources.&#x200b;limits**  | map\[string\]UNKNOW TYPE | Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| **resourceConfiguration.&#x200b;build.&#x200b;resources.&#x200b;requests**  | map\[string\]UNKNOW TYPE | Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| **resourceConfiguration.&#x200b;function**  | object | Function specifies resources requested by the Function's Pod. |
| **resourceConfiguration.&#x200b;function.&#x200b;profile**  | string | Profile defines name of predefined set of values of resource. Can't be used at the same time with Resources. |
| **resourceConfiguration.&#x200b;function.&#x200b;resources**  | object | Resources defines amount of resources available for the Pod to use. Can't be used at the same time with Profile. |
| **resourceConfiguration.&#x200b;function.&#x200b;resources.&#x200b;limits**  | map\[string\]UNKNOW TYPE | Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| **resourceConfiguration.&#x200b;function.&#x200b;resources.&#x200b;requests**  | map\[string\]UNKNOW TYPE | Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| **runtime** (required) | string | Runtime specifies the runtime of the Function. The available values are `nodejs16`, `nodejs18`, and `python39`. |
| **runtimeImageOverride**  | string | RuntimeImageOverride specifies the runtimes image which must be used instead of the default one. |
| **scaleConfig**  | object | ScaleConfig defines minimum and maximum number of Function's Pods to run at a time. When it is configured, a HorizontalPodAutoscaler will be deployed and will control the Replicas field to scale Function based on the CPU utilisation. |
| **scaleConfig.&#x200b;maxReplicas** (required) | integer | MaxReplicas defines the maximum number of Function's Pods to run at a time. |
| **scaleConfig.&#x200b;minReplicas** (required) | integer | MinReplicas defines the minimum number of Function's Pods to run at a time. |
| **secretMounts**  | \[\]object | SecretMounts specifies Secrets to mount into the Function's container filesystem. |
| **secretMounts.&#x200b;mountPath** (required) | string | MountPath specifies path within the container at which the Secret should be mounted. |
| **secretMounts.&#x200b;secretName** (required) | string | SecretName specifies name of the Secret in the Function's Namespace to use. |
| **source** (required) | object | Source contains the Function's specification. |
| **source.&#x200b;gitRepository**  | object | GitRepository defines Function as git-sourced. Can't be used at the same time with Inline. |
| **source.&#x200b;gitRepository.&#x200b;auth**  | object | Auth specifies that you must authenticate to the Git repository. Required for SSH. |
| **source.&#x200b;gitRepository.&#x200b;auth.&#x200b;secretName** (required) | string | SecretName specifies the name of the Secret with credentials used by the Function Controller to authenticate to the Git repository in order to fetch the Function's source code and dependencies. This Secret must be stored in the same Namespace as the Function CR. |
| **source.&#x200b;gitRepository.&#x200b;auth.&#x200b;type** (required) | string | RepositoryAuthType defines if you must authenticate to the repository with a password or token (`basic`), or an SSH key (`key`). For SSH, this parameter must be set to `key`. |
| **source.&#x200b;gitRepository.&#x200b;baseDir**  | string | BaseDir specifies the relative path to the Git directory that contains the source code from which the Function is built. |
| **source.&#x200b;gitRepository.&#x200b;reference**  | string | Reference specifies either the branch name, tag or the commit revision from which the Function Controller automatically fetches the changes in the Function's code and dependencies. |
| **source.&#x200b;gitRepository.&#x200b;url** (required) | string | URL provides the address to the Git repository with the Function's code and dependencies. Depending on whether the repository is public or private and what authentication method is used to access it, the URL must start with the `http(s)`, `git`, or `ssh` prefix. |
| **source.&#x200b;inline**  | object | Inline defines Function as the inline Function. Can't be used at the same time with GitRepository. |
| **source.&#x200b;inline.&#x200b;dependencies**  | string | Dependencies specifies the Function's dependencies. |
| **source.&#x200b;inline.&#x200b;source** (required) | string | Source provides the Function's full source code. |
| **template**  | object | Deprecated: `.spec.Labels` and `.spec.Annotations` should be used to annotate/label function's pods. |
| **template.&#x200b;annotations**  | map\[string\]string |  |
| **template.&#x200b;labels**  | map\[string\]string |  |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **baseDir**  | string | BaseDir specifies the relative path to the Git directory that contains the source code from which the Function is built. |
| **commit**  | string |  |
| **conditions**  | \[\]object |  |
| **conditions.&#x200b;lastTransitionTime**  | string |  |
| **conditions.&#x200b;message**  | string |  |
| **conditions.&#x200b;reason**  | string |  |
| **conditions.&#x200b;status** (required) | string |  |
| **conditions.&#x200b;type**  | string | TODO: Status related things needs to be developed. |
| **podSelector**  | string |  |
| **reference**  | string | Reference specifies either the branch name, tag or the commit revision from which the Function Controller automatically fetches the changes in the Function's code and dependencies. |
| **replicas**  | integer |  |
| **runtime**  | string | Runtime specifies the name of the Function's runtime. |
| **runtimeImage**  | string |  |
| **runtimeImageOverride**  | string | Deprecated: RuntimeImageOverride exists for historical compatibility and should be removed with v1alpha3 version. RuntimeImage has the override image if it isn't empty. |

<!-- TABLE-END -->

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
| `HorizontalPodAutoscalerCreated` | `Running`            | A new Horizontal Pod Scaler referencing the Function's Deployment was created.                                                                                  |
| `HorizontalPodAutoscalerUpdated` | `Running`            | The existing Horizontal Pod Scaler was updated after applying required changes.                                                                                 |
| `MinimumReplicasUnavailable`     | `Running`            | Insufficient number of available Replicas. The Function is unhealthy.                                                                                                       |

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
