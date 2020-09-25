---
title: Supported webhooks
type: Details
---

A newly created or modified Function CR is first updated by the defaulting webhook and then verified by the validation webhook before the Function Controller starts to process it.

## Defaulting webhook

> **NOTE:** It only applies to the [Function custom resource (CR)](#custom-resource-function).

The defaulting webhook sets the default values for CPU and memory requests and limits, and adds the maximum and the minimum number of replicas, if not specified already in the Function CR. It also sets the default runtime `nodejs12` unless specified otherwise.

   | Parameter         | Default value |
   | ----------------- | ------------- |
   | **requestCpu**    | `50m`         |
   | **requestMemory** | `64Mi`        |
   | **limitsCpu**     | `100m`        |
   | **limitsMemory**  | `128Mi`       |
   | **minReplicas**   | `1`           |
   | **maxReplicas**   | `1`           |
   | **runtime**       | `nodejs12`    |

## Validation webhook

It checks the following conditions for these CRs:

1. [Function CR](#custom-resource-function)

   - Minimum values requested for CPU, memory, and replicas are not lower than the required ones:

   | Parameter            | Required value |
   | -------------------- | -------------- |
   | **minRequestCpu**    | `10m`          |
   | **minRequestMemory** | `16Mi`         |
   | **minReplicasValue** | `1`            |

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

   - If you use SSH to authenticate to repository, you have to provide:

     - **spec.auth.type** must be set to `key`
     - **spec.auth.secretName** must not be empty

   - If you use HTTP(s) URL to point to repository and repository requires authentication (**spec.auth**):
   
      - **spec.auth.type** must be set to either `key` (SSH key) or `basic` (password or token)
      - **spec.auth.secretName** must not be empty
