---
title: Supported webhooks
type: Details
---

A newly created or modified Function CR is first updated by the defaulting webhook and then verified by the validation webhook before the Function Controller starts to process it.

## Defaulting webhook

> **NOTE:** It only applies to the [Function custom resource (CR)](#custom-resource-function).

The defaulting webhook:

- Sets default values for CPU and memory requests and limits for a Function.
- Sets default values for CPU and memory requests and limits for a Kubernetes Job responsible for building the Function's image.
- Adds the maximum and the minimum number of replicas, if not specified already in the Function CR.
- Sets the default `nodejs14` runtime unless specified otherwise.

   | Parameter       | Default value |
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
   | **runtime**       | `nodejs14`    |

> **NOTE:** Function's resources and replicas as well as resources for a Kubernetes Job are based on presets. Read about all [available presets](#details-available-presets) to find out more.

## Validation webhook

It checks the following conditions for these CRs:

1. [Function CR](#custom-resource-function)

   - Minimum values requested for a Function (CPU, memory, and replicas) and a Kubernetes Job (CPU and memory) responsible for building the Function's image must not be lower than the required ones:

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

## Admission webhook

There is also the admission webhook that is only triggered in the scenario when you switch at runtime from one registry to another.

To switch registries at runtime, you must create or update a Secret CR that complies with [specific requirements](#details-switching-registries-at-runtime). This Secret CR, among other details, contains a username and password to the registry. The admission webhook encodes these credentials to base64, a format that is required by Kaniko - the Function's job building tool - to access the registry and push a Function's image to it. These encoded credentials also allow Kubernetes to pull images of deployed Functions from the registry. The admission webhook adds these credentials to the created Secret CR. They take the form of the `.dockerconfigjson` entry with a valid value. The admission webhook also updates this entry's value each time the username and password change.

### Requirements

This admission webhook is triggered every time you `CREATE` or `UPDATE` a Secret in any Namespace, as long as the Secret contains the `serverless.kyma-project.io/remote-registry: config` label.
