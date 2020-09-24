---
title: Supported webhooks
type: Details
---

A newly created or modified Function CR is first updated by the defaulting webhook and then verified by the validation webhook before the Function Controller starts to process it.

## Defaulting webhook

> **NOTE:** It only applies to the [Function custom resource (CR)](#custom-resource-function).

The defaulting webhook sets the default values for CPU and memory requests and limits of Function, default values for CPU and memory requests and limits of Kubernetes Job responsible for building Function's image, and adds the maximum and the minimum number of replicas, if not specified already in the Function CR. It also sets the default runtime `nodejs12` unless specified otherwise.

   | Parameter         | Default value |
   | ----------------- | ------------- |
   | **resources.requests.cpu**    | `50m`         |
   | **resources.requests.memory** | `64Mi`        |
   | **resources.limits.cpu**     | `100m`        |
   | **resources.limits.memory**  | `128Mi`       |
   | **buildResources.requests.cpu**    | `700m`         |
   | **buildResources.requests.memory** | `700Mi`        |
   | **buildResources.limits.cpu**     | `1100m`        |
   | **buildResources.limits.memory**  | `1100Mi`       |
   | **minReplicas**   | `1`           |
   | **maxReplicas**   | `1`           |
   | **runtime**       | `nodejs12`    |
  
> **NOTE:** Function's resources and replicas and resources for Kubernetes Job are based on presets. Read [available presets](#details-supported-webhooks-available-presets) to find out more.

## Validation webhook

It checks the following conditions for these CRs:

1. [Function CR](#custom-resource-function)

   - Minimum values requested for CPU, memory, and replicas for Function and CPU and memory for Kubernetes Job responsible for building Function's image are not lower than the required ones:

   | Parameter            | Required value |
   | -------------------- | -------------- |
   | **minReplicas** | `1`            |
   | **resources.requests.cpu**    | `10m`          |
   | **resources.requests.memory** | `16Mi`         |
   | **buildResources.requests.cpu**    | `200m`          |
   | **buildResources.requests.memory** | `200Mi`         |

   - Requests are lower than or equal to limits, and the minimum number of replicas is lower than or equal to the maximum one.
   - The Function CR contains all the required parameters.
   - If you decide to set a Git repository as the source of your Function's code and dependencies (**spec.type** set to `git`), the **spec.reference** and **spec.baseDir** fields must contain values.
   - The format of deps, envs, labels, and the Function name ([RFC 1035](https://tools.ietf.org/html/rfc1035)) is correct.
   - The Function CR contains any envs reserved for the Deployment: `FUNC_RUNTIME`, `FUNC_HANDLER`, `FUNC_PORT`, `MOD_NAME`, `NODE_PATH`, `PYTHONPATH`.

2. [GitRepository CR](#custom-resource-git-repository)

   - The **spec.url** parameter must:

      - Not be empty
      - Start with the `http(s)`, `git`, or `ssh` prefix
      - End with the `.git` suffix

   - If you use SSH to authenticate to the repository:

     - **spec.auth.type** must be set to `key`
     - **spec.auth.secretName** must not be empty

   - If you use HTTP(S) to point to the repository that requires authentication (**spec.auth**):
   
      - **spec.auth.type** must be set to either `key` (SSH key) or `basic` (password or token)
      - **spec.auth.secretName** must not be empty

## Available presets

Function's resources and replicas and resources for image-building Job are based on presets. Preset is predefined group of values. Function CR has defined three group of presets, for Function's resources, Function's replicas and image-building Job's resources.

> **NOTE:** To change/add new preset, update configuration for **webhook.values.function.replicas.presets**, **webhook.values.function.resources.presets** or **webhook.values.buildJob.resources.presets** in `values.yaml` file of the Serverless chart. Read the [Serverless chart configuration](#configuration-serverless-chart) to find out more.

> **NOTE:** To override existing values in Function CR for a given preset, in edit mode, remove first appropriate field(s) and use a new preset(s).

### Function's replicas

| Preset | Minimum number | Maximum number |
| - | - | - |
| `S` | 1 | 1 |
| `M` | 1 | 2 |
| `L` | 1 | 5 |
| `XL` | 1 | 10 |

To apply values ​​from a given preset, use `serverless.kyma-project.io/replicas-preset: {PRESET}` label in Function CR.

### Function's resources

| Preset | Request CPU | Request memory | Limit CPU | Limit memory |
| - | - | - | - | - |
| `XS` | `10m` | `16Mi` | `25m` | `32Mi` |
| `S` | `25m` | `32Mi` | `50m` | `64Mi` |
| `M` | `50m` | `64Mi` | `100m` | `128Mi` |
| `L` | `100m` | `128Mi` | `200m` | `256Mi` |
| `XL` | `200m` | `256Mi` | `400m` | `512Mi` |

To apply values ​​from a given preset, use `serverless.kyma-project.io/function-resources-preset: {PRESET}` label in Function CR.

### Build Job's resources

| Preset | Request CPU | Request memory | Limit CPU | Limit memory |
| - | - | - | - | - |
| `local-dev` | `200m` | `200Mi` | `400m` | `400Mi` |
| `slow` | `400m` | `400Mi` | `700m` | `700Mi` |
| `normal` | `700m` | `700Mi` | `1100m` | `1100Mi`|
| `fast` | `1100m` | `1100Mi` | `1700m` | `1700Mi`|

To apply values ​​from a given preset, use `serverless.kyma-project.io/build-resources-preset: {PRESET}` label in Function CR.
