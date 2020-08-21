---
title: Supported webhooks
type: Details
---

A newly created or modified Function CR is first updated by the defaulting webhook and then verified by the validation webhook before the Function Controller starts to process it:

1. **Defaulting webhook** sets the default values for CPU and memory requests and limits, and adds the maximum and the minimum number of replicas, if not specified already in the Function CR.

   | Parameter         | Default value |
   | ----------------- | ------------- |
   | **requestCpu**    | `50m`         |
   | **requestMemory** | `64Mi`        |
   | **limitsCpu**     | `100m`        |
   | **limitsMemory**  | `128Mi`       |
   | **minReplicas**   | `1`           |
   | **maxReplicas**   | `1`           |

2. **Validation webhook** checks if:

   - Minimum values requested for CPU, memory, and replicas are not lower than the required ones:

   | Parameter            | Required value |
   | -------------------- | -------------- |
   | **minRequestCpu**    | `10m`          |
   | **minRequestMemory** | `16Mi`         |
   | **minReplicasValue** | `1`            |

   - Requests are lower than or equal to limits, and the minimum number of replicas is lower than or equal to the maximum one.
   - The Function CR contains all the required parameters.
   - The format of deps, envs, labels, and the Function name ([RFC 1035](https://tools.ietf.org/html/rfc1035)) is correct.
   - The Function CR contains any envs reserved for the Deployment: `FUNC_RUNTIME`, `FUNC_HANDLER`, `FUNC_PORT`, `MOD_NAME`, `NODE_PATH`, `PYTHONPATH`
