---
title: Serverless limitations
---

## Controller limitations
Serverless controller does not serve time-critical requests from users.
It reconciles Function custom resources (CR), stored at the Kubernetes API Server, and has no persistent state on its own.

Serverless controller doesn't build or serve Functions using its allocated runtime resources. It delegates this work to the dedicated Kubernetes workloads. It schedules (build-time) jobs to build the Function Docker image and (runtime) Pods to serve them once they are built. 
Refer to the [architecture](../../05-technical-reference/00-architecture/svls-01-architecture.md) diagram for more details.

Having this in mind Serverless controller does not require horizontal scaling.
It scales vertically up to the `160Mi` of memory and `500m` of CPU time.

## Limitation for the number of Functions
There is no upper limit of Functions that can be run on Kyma (similar to Kubernetes workloads in general). Once a user defines a Function, its build jobs and runtime Pods will always be requested by Serverless controller. It's up to Kubernetes to schedule them based on the available memory and CPU time on the Kubernetes worker nodes. This is determined mainly by the number of the Kubernetes worker nodes (and the node auto-scaling capabilities) and their computational capacity.

## Build phase limitation:
The time necessary to build Function depends on:
 - selected [build profile](../../05-technical-reference/svls-08-available-presets.md#build-jobs-resources) that determines the requested resources (and their limits) for the build phase 
 - number and size of dependencies that must be downloaded and bundled into the Function image.
 - cluster nodes specification (see the note with reference specification at the end of the article)

<div tabs name="build" group="function-build-times">
  <details>
  <summary label="nodejs">
  Node.js 
  </summary>

|                 | local-dev | no profile (no limits for resource) |
|-----------------|-----------|-------------------------------------|
| no dependencies | 24 sec    | 15 sec                              |
| 2 dependencies  | 26 sec    | 16 sec                              |

  </details>
  <details>
  <summary label="python">
  Python
  </summary>

|                 | local-dev | no profile (no limits for resource) |
|-----------------|-----------|-------------------------------------|
| no dependencies | 30 sec    | 16 sec                              |
| 2 dependencies  | 32 sec    | 20 sec                              |

  </details>
</div>

The shortest build time (the limit) is approximately 15 seconds and requires no limitation of the build job resources and a minimum number of dependencies that are pulled in during the build phase.

Running multiple Function build jobs at once (especially with no limits) may drain the cluster resources. To mitigate such risk, there is an additional limit of 5 simultaneous Function builds. If a sixth one is scheduled, it is built once there is a vacancy in the build queue.

This limitation is configurable using [`containers.manager.envs.functionBuildMaxSimultaneousJobs`](../../05-technical-reference/00-configuration-parameters/svls-01-serverless-chart.md#configurable-parameters).


## Runtime phase limitations
In the runtime, the Functions serve user-provided logic wrapped in the WEB framework (`express` for Node.js and `bottle` for Python). Taking the user logic aside, those frameworks have limitations and depend on the selected [runtime profile](../../05-technical-reference/svls-08-available-presets.md#functions-resources) and the Kubernetes nodes specification (see the note with reference specification at the end of this article).

The following describes the response times of the selected runtime profiles for a "hello world" Function requested at 50 requests/second. This describes the overhead of the serving framework itself. Any user logic added on top of that will add extra milliseconds and must be profiled separately.


<div tabs name="steps" group="function-response-times">
  <details>
  <summary label="nodejs">
  Node.js 
  </summary>
|                               | XL     | L      | M      | S      | XS      |
|-------------------------------|--------|--------|--------|--------|---------|
| response time [avarage]       | ~13ms  | 13ms   | ~15ms  | ~60ms  | ~400ms  |
| response time [95 percentile] | ~20ms  | ~30ms  | ~70ms  | ~200ms | ~800ms  |
| response time [99 percentile] | ~200ms | ~200ms | ~220ms | ~500ms | ~1.25ms |


  </details>
  <details>
  <summary label="python">
  Python
  </summary>

|                               | XL     | L      | M      | S      |
|-------------------------------|--------|--------|--------|--------|
| response time [avarage]       | ~11ms  | 12ms   | ~12ms  | ~14ms  |
| response time [95 percentile] | ~25ms  | ~25ms  | ~25ms  | ~25ms  |
| response time [99 percentile] | ~175ms | ~180ms | ~210ms | ~280ms |
  </details>
</div>

Obviously, the bigger the runtime profile, the more resources are available to serve the response quicker. Consider these limits of the serving layer as a baseline - as this does not take your Function logic into account.


### Scaling

Function runtime Pods can be scaled horizontally from zero up to the limits of the available resources at the Kubernetes worker nodes.
See the [Use external scalers](../../03-tutorials/00-serverless/svls-15-use-external-scalers.md) tutorial for more information.

## In-cluster Docker registry limitations

Serverless comes with an in-cluster Docker registry for the Function images.
This registry is only suitable for development because of its [limitations](../serverless/svls-03-container-registries.md), i.e.:
 - Registry capacity is limited to 20GB
 - There is no image lifecycle management. Once an image is stored in the registry, it stays there until it is manually removed.

> **NOTE:** All measurements were done on Kubernetes with five AWS worker nodes of type `m5.xlarge` (four CPU 3.1 GHz x86_64 cores, 16 GiB memory).